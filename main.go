package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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
		"(search: %s, replace: %s, dry-run: %v)",
		p.Search, p.Replace, p.DryRun)

	replace := Replace{
		Search:  p.Search,
		Replace: p.Replace,
	}

	entries := Find(p.RootDirectory, Filter)

	// iterate reversed, so directories are renamed after files are written
	for i := len(entries) - 1; i >= 0; i-- {
		path := entries[i]

		if p.Verbose {
			p.reportInfo("Processing %s...", path)
		}

		file, err := os.Open(path)
		if err != nil {
			p.reportError("Could not open: %s (%s)", path, err)
			continue
		}
		fileInfo, err := file.Stat()
		if err != nil {
			p.reportError("Could not stat: %s (%s)", path, err)
			continue
		}
		// close file directly (no defer) to prevent to many open files error
		file.Close()

		// Step 1 - Replace search string in files content
		if !fileInfo.IsDir() {
			bytes, err := ioutil.ReadFile(path)
			if err != nil {
				p.reportError("Could not read: %s (%s)", path, err)
				continue
			}
			content := string(bytes)
			newContent := replace.Execute(content)
			if newContent != content {
				p.reportInfo("Write: %s", path)
				if !p.DryRun {
					err = ioutil.WriteFile(path, []byte(newContent), fileInfo.Mode())
					if err != nil {
						p.reportError("Could not write: %s (%s)", path, err)
						continue
					}
				}
			}
		}

		// Step 2 - Replace search string in file or directory name
		baseName := filepath.Base(path)
		newName := replace.Execute(baseName)
		if newName != baseName {
			newPath := filepath.Join(filepath.Dir(path), newName)
			p.reportInfo("Move to: %s", newPath)
			if !p.DryRun {
				err = os.Rename(path, newPath)
				p.reportError("Could not move: %s (%s)", path, err)
				continue
			}
		}
	}
	return
}

func (p Program) reportError(format string, a ...interface{}) {
	fmt.Fprintf(p.Stdout, "[ERROR] "+format+"\n", a...)
}

func (p Program) reportInfo(format string, a ...interface{}) {
	fmt.Fprintf(p.Stdout, "[INFO] "+format+"\n", a...)
}
