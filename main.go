package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/mgutz/ansi"
)

var styleRed = ansi.ColorFunc("red")
var styleRedUnderline = ansi.ColorFunc("red+u")
var styleGreen = ansi.ColorFunc("green")
var styleGreenUnderline = ansi.ColorFunc("green+u")
var styleHeader = ansi.ColorFunc("white+b:black")
var styleBold = ansi.ColorFunc("+b")

func main() {
	dir, _ := os.Getwd()
	exitCode := mainSub(dir, os.Stdout, os.Stdin, os.Args[1:])
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

type options struct {
	DryRun      bool `short:"d" long:"dry-run"     description:"Do not change anything"`
	Regexp      bool `short:"r" long:"regexp"      description:"Treat search string as regular expression"`
	Verbose     bool `short:"v" long:"verbose"     description:"Show verbose debug information"`
	Interactive bool `short:"i" long:"interactive" description:"Confirm every replacement"`
	Args        struct {
		Search  string
		Replace string
	} `positional-args:"yes" required:"yes"`
}

func mainSub(workingDir string, stdout io.Writer, stdin io.Reader, args []string) int {

	opts, exitCode := parseOptions(stdout, args)
	if opts == nil {
		return exitCode
	}

	program := Program{
		RootDirectory: workingDir,
		Search:        opts.Args.Search,
		Replace:       opts.Args.Replace,
		Stdout:        stdout,
		Stdin:         stdin,
		DryRun:        opts.DryRun,
		Verbose:       opts.Verbose,
		Regexp:        opts.Regexp,
		Interactive:   opts.Interactive,
	}
	program.Execute()

	return 0
}

type Program struct {
	RootDirectory string
	Search        string
	Replace       string
	Stdout        io.Writer
	Stdin         io.Reader
	DryRun        bool
	Verbose       bool
	Regexp        bool
	Interactive   bool
}

func (p Program) Execute() {
	p.reportInfo(
		"(search: %s, replace: %s, dry-run: %v, regexp: %v)",
		p.Search, p.Replace, p.DryRun, p.Regexp)
	p.reportVerbose("Root-Directory: %s", p.RootDirectory)

	if p.Regexp {
		_, err := regexp.Compile(p.Search)
		if err != nil {
			p.reportError(
				"Could not compile regular expression: %s - %s",
				p.Search, err)
			return
		}
	}

	replace := Replace{
		Search:  p.Search,
		Replace: p.Replace,
		Regexp:  p.Regexp,
	}

	ask := Ask{
		Stdin:  p.Stdin,
		Stdout: p.Stdout,
	}

	entries := Find(p.RootDirectory, Filter)

	// iterate reversed, so directories are renamed after files are written
	for i := len(entries) - 1; i >= 0; i-- {
		path := entries[i]

		p.reportVerbose(
			"Processing(%d/%d) %s...", len(entries)-i, len(entries), p.shortenPath(path))

		file, err := os.Open(path)
		if err != nil {
			p.reportError("Could not open: %s (%s)", p.shortenPath(path), err)
			continue
		}
		fileInfo, err := file.Stat()
		if err != nil {
			p.reportError("Could not stat: %s (%s)", p.shortenPath(path), err)
			continue
		}
		// close file directly (no defer) to prevent to many open files error
		file.Close()

		// Step 1 - Replace search string in files content
		if !fileInfo.IsDir() {
			bytes, err := ioutil.ReadFile(path)
			if err != nil {
				p.reportError("Could not read: %s (%s)", p.shortenPath(path), err)
				continue
			}

			matchCount := 0

			content := string(bytes)
			newContent := replace.Execute(content, func(info ReplacementInfo) bool {
				matchCount++

				p.printHeader("Match #%d in %s", matchCount, p.shortenPath(path))
				p.reportReplacement(info)

				if p.Interactive && !ask.question(styleBold("Replace?")) {
					return false
				}

				return true
			})
			if newContent != content {
				p.reportInfo("Write: %s", p.shortenPath(path))
				if !p.DryRun {
					err = ioutil.WriteFile(path, []byte(newContent), fileInfo.Mode())
					if err != nil {
						p.reportError("Could not write: %s (%s)", p.shortenPath(path), err)
						continue
					}
				}
			}
		}

		// Step 2 - Replace search string in file or directory name
		baseName := filepath.Base(path)
		newName := replace.Execute(baseName, func(info ReplacementInfo) bool {

			p.printf("\nRename %s to %s\n", info.Match, info.Repl)

			if p.Interactive && !ask.question(styleBold("Rename?")) {
				return false
			}

			return true
		})
		if newName != baseName {
			newPath := filepath.Join(filepath.Dir(path), newName)
			p.printHeader("Move to: %s", p.shortenPath(newPath))
			if !p.DryRun {
				err = os.Rename(path, newPath)
				if err != nil {
					p.reportError("Could not move: %s (%s)", p.shortenPath(path), err)
					continue
				}
			}
		}
	}
	return
}

func (p Program) shortenPath(path string) string {
	return strings.Replace(path, p.RootDirectory+"/", "", 1)
}

func (p Program) reportError(format string, a ...interface{}) {
	fmt.Fprintf(p.Stdout, "[ERROR] "+format+"\n", a...)
}

func (p Program) reportInfo(format string, a ...interface{}) {
	fmt.Fprintf(p.Stdout, "[INFO] "+format+"\n", a...)
}

func (p Program) reportReplacement(info ReplacementInfo) {
	p.print(info.LinesBeforeMatch)

	p.print(styleRed(info.MatchLine[:info.MatchLineMatchIndex[0]]))
	p.print(styleRedUnderline(info.Match))
	p.print(styleRed(info.MatchLine[info.MatchLineMatchIndex[1]:]))

	p.print(styleGreen(info.ReplLine[:info.ReplLineReplIndex[0]]))
	p.print(styleGreenUnderline(info.Repl))
	p.print(styleGreen(info.ReplLine[info.ReplLineReplIndex[1]:]))

	p.print(info.LinesAfterMatch)
}

func (p Program) print(s string) {
	fmt.Fprint(p.Stdout, s)
}

func (p Program) printf(format string, a ...interface{}) {
	fmt.Fprintf(p.Stdout, format, a...)
}

func (p Program) printHeader(format string, a ...interface{}) {
	fmt.Fprintf(p.Stdout, styleHeader("\n "+format)+"\n", a...)
}

func (p Program) reportVerbose(format string, a ...interface{}) {
	if !p.Verbose {
		return
	}
	p.reportInfo(format, a...)
}

func parseOptions(stdout io.Writer, args []string) (*options, int) {
	opts := options{}

	parser := flags.NewParser(&opts, flags.PassDoubleDash|flags.HelpFlag)
	args, err := parser.ParseArgs(args)
	if err != nil {
		if parserErr, ok := err.(*flags.Error); ok {
			if parserErr.Type != flags.ErrHelp {
				fmt.Printf("%s\n\n", parserErr.Message)
			}

			var b bytes.Buffer
			parser.WriteHelp(&b)
			fmt.Fprint(stdout, b.String())

			if parserErr.Type == flags.ErrHelp {
				return nil, 0
			}
		} else {
			panic(err)
		}
		return nil, 2
	}

	return &opts, 0
}
