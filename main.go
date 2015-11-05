package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: search-and-replace [flags] search replace\n")
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	flag.Usage = usage
	dryRun := flag.Bool("dry-run", false, "If true -> do not change anything [default: false]")
	flag.Parse()
	if flag.NArg() != 2 {
		usage()
	}

	dir, _ := os.Getwd()
	program := Program{
		RootDirectory: dir,
		Search:        flag.Arg(0),
		Replace:       flag.Arg(1),
		Stdout:        os.Stdout,
		DryRun:        *dryRun,
	}
	err := program.Execute()
	if err != nil {
		panic(err)
	}
}

type Program struct {
	RootDirectory string
	Search        string
	Replace       string
	Stdout        io.Writer
	DryRun        bool
}

func (p Program) Execute() (err error) {
	fmt.Fprintf(p.Stdout,
		"Searching for: %s and replacing with: %s (dry-run: %v)...\n",
		p.Search, p.Replace, p.DryRun)

	replace := Replace{
		Search:  p.Search,
		Replace: p.Replace,
	}

	entries := Find(p.RootDirectory, Filter)

	// iterate reversed, so directories are renamed after files are written
	for i := len(entries) - 1; i >= 0; i-- {
		path := entries[i]

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		fileInfo, err := file.Stat()
		if err != nil {
			return err
		}
		// close file directly (no defer) to prevent to many open files error
		file.Close()

		// Step 1 - Replace search string in files content
		if !fileInfo.IsDir() {
			bytes, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			content := string(bytes)
			newContent := replace.Execute(content)
			if newContent != content {
				fmt.Fprintln(p.Stdout, "Write: "+path)
				if !p.DryRun {
					err = ioutil.WriteFile(path, []byte(newContent), fileInfo.Mode())
					panicIfErr(err)
				}
			}
		}

		// Step 2 - Replace search string in file or directory name
		baseName := filepath.Base(path)
		newName := replace.Execute(baseName)
		if newName != baseName {
			newPath := filepath.Join(filepath.Dir(path), newName)
			fmt.Fprintln(p.Stdout, "Move to: "+newPath)
			if !p.DryRun {
				err = os.Rename(path, newPath)
				panicIfErr(err)
			}
		}
	}
	return
}

func panicIfErr(err interface{}) {
	if err != nil {
		panic(err)
	}
}
