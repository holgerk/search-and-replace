package main

import (
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

func main() {
	dir, _ := os.Getwd()
	mainSub(dir, os.Stdout, os.Args[1:])
}

func mainSub(workingDir string, stdout io.Writer, args []string) {
	var opts struct {
		DryRun  bool `short:"d" long:"dry-run" description:"Do not change anything"`
		Regexp  bool `short:"r" long:"regexp" description:"Treat search string as regular expression"`
		Verbose bool `short:"v" long:"verbose" description:"Show verbose debug information"`
		Args    struct {
			Search  string
			Replace string
		} `positional-args:"yes" required:"yes"`
	}

	parser := flags.NewParser(&opts, flags.Default)
	args, err := parser.ParseArgs(args)
	if err != nil {
		return
	}

	program := Program{
		RootDirectory: workingDir,
		Search:        opts.Args.Search,
		Replace:       opts.Args.Replace,
		Stdout:        stdout,
		DryRun:        opts.DryRun,
		Verbose:       opts.Verbose,
		Regexp:        opts.Regexp,
	}
	program.Execute()
}

type Program struct {
	RootDirectory string
	Search        string
	Replace       string
	Stdout        io.Writer
	DryRun        bool
	Verbose       bool
	Regexp        bool
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

				p.reportInfo("Match #%d in %s", matchCount, p.shortenPath(path))
				p.reportReplacement(info)
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
			return true
		})
		if newName != baseName {
			newPath := filepath.Join(filepath.Dir(path), newName)
			p.reportInfo("Move to: %s", p.shortenPath(newPath))
			if !p.DryRun {
				err = os.Rename(path, newPath)
				p.reportError("Could not move: %s (%s)", p.shortenPath(path), err)
				continue
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

var red = ansi.ColorFunc("red")
var redUnderline = ansi.ColorFunc("red+u")
var green = ansi.ColorFunc("green")
var greenUnderline = ansi.ColorFunc("green+u")

func (p Program) reportReplacement(info ReplacementInfo) {
	p.print(info.LinesBeforeMatch)

	p.print(red(info.MatchLine[:info.MatchLineMatchIndex[0]]))
	p.print(redUnderline(info.MatchLine[info.MatchLineMatchIndex[0]:info.MatchLineMatchIndex[1]]))
	p.print(red(info.MatchLine[info.MatchLineMatchIndex[1]:]))

	p.print(green(info.ReplLine[:info.ReplLineReplIndex[0]]))
	p.print(greenUnderline(info.ReplLine[info.ReplLineReplIndex[0]:info.ReplLineReplIndex[1]]))
	p.print(green(info.ReplLine[info.ReplLineReplIndex[1]:]))

	p.print(info.LinesAfterMatch)
}

func (p Program) print(s string) {
	fmt.Fprint(p.Stdout, s)
}

func (p Program) reportVerbose(format string, a ...interface{}) {
	if !p.Verbose {
		return
	}
	p.reportInfo(format, a...)
}
