package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/mgutz/ansi"
)

func main() {
	dir, _ := os.Getwd()
	mainSub(dir, os.Stdout, os.Args[1:])
}

func mainSub(workingDir string, stdout io.Writer, args []string) {
	flagSet := flag.NewFlagSet("main", flag.ExitOnError)

	dryRun := flagSet.Bool("dry-run", false, "If true -> do not change anything [default: false]")
	verbose := flagSet.Bool("verbose", false, "If true -> increase verbosity [default: false]")

	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: search-and-replace [flags] search replace\n")
		flagSet.PrintDefaults()
		os.Exit(1)
	}

	flagSet.Parse(args)

	if flagSet.NArg() != 2 {
		flagSet.Usage()
	}

	program := Program{
		RootDirectory: workingDir,
		Search:        flagSet.Arg(0),
		Replace:       flagSet.Arg(1),
		Stdout:        stdout,
		DryRun:        *dryRun,
		Verbose:       *verbose,
	}
	err := program.Execute()
	if err != nil {
		panic(fmt.Errorf("Program-Execution error(%s)", err))
	}
}

type Program struct {
	RootDirectory string
	Search        string
	Replace       string
	Stdout        io.Writer
	DryRun        bool
	Verbose       bool
}

func (p Program) Execute() (err error) {
	p.reportInfo(
		"(search: %s, replace: %s, dry-run: %v)", p.Search, p.Replace, p.DryRun)
	p.reportVerbose("Root-Directory: %s", p.RootDirectory)

	red := ansi.ColorFunc("red")
	redBold := ansi.ColorFunc("red+b")
	green := ansi.ColorFunc("green")
	greenBold := ansi.ColorFunc("green+b")

	replace := Replace{
		Search:  p.Search,
		Replace: p.Replace,
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
			content := string(bytes)
			newContent := replace.Execute(content, func(info ReplaceInfo) bool {
				p.reportInfo("Match #%d", 1)
				p.print(info.LinesBeforeMatch)

				p.print(red(info.MatchLine[:info.MatchLineMatchIndex[0]]))
				p.print(redBold(info.MatchLine[info.MatchLineMatchIndex[0]:info.MatchLineMatchIndex[1]]))
				p.print(red(info.MatchLine[info.MatchLineMatchIndex[1]:]))

				p.print(green(info.ReplacementLine[:info.ReplacementLineReplacementIndex[0]]))
				p.print(greenBold(info.ReplacementLine[info.ReplacementLineReplacementIndex[0]:info.ReplacementLineReplacementIndex[1]]))
				p.print(green(info.ReplacementLine[info.ReplacementLineReplacementIndex[1]:]))

				p.print(info.LinesAfterMatch)
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
		newName := replace.Execute(baseName, func(info ReplaceInfo) bool {
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

func (p Program) print(s string) {
	fmt.Fprint(p.Stdout, s)
}

func (p Program) reportVerbose(format string, a ...interface{}) {
	if !p.Verbose {
		return
	}
	p.reportInfo(format, a...)
}
