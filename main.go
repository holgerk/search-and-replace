package main

import (
	"bytes"
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

	output := &Output{
		stdout:  stdout,
		verbose: false,
	}

	opts, exitCode := parseOptions(output, args)
	if opts == nil {
		return exitCode
	}

	output.verbose = opts.Verbose

	finder := &Finder{
		output: output,
	}

	program := &Program{
		Output: output,
		Finder: finder,

		RootDirectory: workingDir,
		Stdout:        stdout,
		Stdin:         stdin,

		// args
		Search:  opts.Args.Search,
		Replace: opts.Args.Replace,

		// options
		DryRun:      opts.DryRun,
		Verbose:     opts.Verbose,
		Regexp:      opts.Regexp,
		Interactive: opts.Interactive,
	}
	program.Execute()

	return 0
}

type Program struct {
	Output *Output
	Finder *Finder

	RootDirectory string
	Stdout        io.Writer
	Stdin         io.Reader

	Search  string
	Replace string

	DryRun      bool
	Verbose     bool
	Regexp      bool
	Interactive bool
}

func (p *Program) Execute() {
	p.Output.reportVerbose(
		"(search: %s, replace: %s, dry-run: %v, regexp: %v)",
		p.Search, p.Replace, p.DryRun, p.Regexp)
	p.Output.reportVerbose("Root-Directory: %s", p.RootDirectory)

	if p.Regexp {
		_, err := regexp.Compile(p.Search)
		if err != nil {
			p.Output.reportError(
				"Could not compile regular expression: %s - %s",
				p.Search, err)
			return
		}
	}

	replace := &Replace{
		Search:  p.Search,
		Replace: p.Replace,
		Regexp:  p.Regexp,
	}

	ask := &Ask{
		Stdin:  p.Stdin,
		Stdout: p.Stdout,
	}

	entries := p.Finder.Find(p.RootDirectory, Filter)

	// iterate reversed, so directories are renamed after files are written
	for i := len(entries) - 1; i >= 0; i-- {
		path := entries[i]

		p.Output.reportVerbose(
			"Processing(%d/%d) %s...", len(entries)-i, len(entries), p.shortenPath(path))

		file, err := os.Open(path)
		if err != nil {
			p.Output.reportError("Could not open: %s (%s)", p.shortenPath(path), err)
			continue
		}
		fileInfo, err := file.Stat()
		if err != nil {
			p.Output.reportError("Could not stat: %s (%s)", p.shortenPath(path), err)
			continue
		}
		// close file directly (no defer) to prevent to many open files error
		file.Close()

		// Step 1 - Replace search string in files content
		if !fileInfo.IsDir() {
			bytes, err := ioutil.ReadFile(path)
			if err != nil {
				p.Output.reportError("Could not read: %s (%s)", p.shortenPath(path), err)
				continue
			}

			matchCount := 0

			content := string(bytes)
			newContent := replace.Execute(content, func(info ReplacementInfo) bool {
				matchCount++

				p.Output.printHeader("Match #%d in %s", matchCount, p.shortenPath(path))
				p.Output.reportReplacement(info)

				if p.Interactive && !ask.question(styleBold("Replace?")) {
					return false
				}

				return true
			})
			if newContent != content {
				p.Output.reportInfo("Write: %s", p.shortenPath(path))
				if !p.DryRun {
					err = ioutil.WriteFile(path, []byte(newContent), fileInfo.Mode())
					if err != nil {
						p.Output.reportError(
							"Could not write: %s (%s)", p.shortenPath(path), err)
						continue
					}
				}
			}
		}

		// Step 2 - Replace search string in file or directory name
		baseName := filepath.Base(path)
		newName := replace.Execute(baseName, func(info ReplacementInfo) bool {

			p.Output.printHeader("Rename %s to %s", p.shortenPath(path), info.ReplLine)

			if p.Interactive && !ask.question(styleBold("Rename?")) {
				return false
			}

			return true
		})
		if newName != baseName {
			newPath := filepath.Join(filepath.Dir(path), newName)
			p.Output.reportInfo("Rename: %s", p.shortenPath(newPath))
			if !p.DryRun {
				err = os.Rename(path, newPath)
				if err != nil {
					p.Output.reportError("Could not move: %s (%s)", p.shortenPath(path), err)
					continue
				}
			}
		}
	}
	return
}

func (p *Program) shortenPath(path string) string {
	return strings.Replace(path, p.RootDirectory+"/", "", 1)
}

func parseOptions(output *Output, args []string) (*options, int) {
	opts := options{}

	parser := flags.NewParser(&opts, flags.PassDoubleDash|flags.HelpFlag)
	args, err := parser.ParseArgs(args)
	if err != nil {
		if parserErr, ok := err.(*flags.Error); ok {
			if parserErr.Type != flags.ErrHelp {
				output.printf("%s\n\n", parserErr.Message)
			}

			var b bytes.Buffer
			parser.WriteHelp(&b)
			output.printf("%s", b.String())

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
