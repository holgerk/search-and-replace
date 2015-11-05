package main

// This is an integration Test which runs against real directories and
// compares the transformation against golden directories, containing
// the expected content and structure.
//
// This was heavily inspired by oracle_test.go, thanks.

import (
	"bytes"
	"flag"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var updateFlag = flag.Bool("update", false, "Update golden directories")

func TestMain(t *testing.T) {
	cases := []struct {
		directory string
	}{
		{"testdata/t1"},
		{"testdata/t2"},
	}
	for _, c := range cases {
		referenceDirectory := c.directory
		workingDirectory := referenceDirectory + ".got"
		goldenDirectory := referenceDirectory + ".golden"

		os.RemoveAll(workingDirectory)
		err := copyDirectory(referenceDirectory, workingDirectory)
		if err != nil {
			panic(err)
		}

		var stdout bytes.Buffer
		program := Program{
			RootDirectory: workingDirectory,
			Search:        "foo",
			Replace:       "bar",
			Stdout:        &stdout,
		}
		program.Execute()

		// compare directory.got with directory.golden
		cmd := exec.Command("diff", "-ru", workingDirectory, goldenDirectory)
		cmd.Stdout = new(bytes.Buffer)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			t.Errorf("Test for %s failed: %s.\n%s\n",
				workingDirectory, err, cmd.Stdout)

			if *updateFlag {
				t.Logf("Updating golden: %s...", goldenDirectory)
				os.RemoveAll(goldenDirectory)
				err := copyDirectory(workingDirectory, goldenDirectory)
				if err != nil {
					t.Errorf("Update failed: %s", err)
				}
			}
		}
	}

	// fmt.Println(stdout.String())
}

func copyDirectory(source, target string) (err error) {
	err = filepath.Walk(source, func(path string, f os.FileInfo, err error) (werr error) {
		newPath := strings.Replace(path, source, target, 1)
		if f.IsDir() {
			if werr = os.Mkdir(newPath, f.Mode()); werr != nil {
				return
			}
		} else {
			if werr = copyFile(path, newPath); werr != nil {
				return
			}
		}
		return
	})
	return
}

// see: http://stackoverflow.com/a/21067803
func copyFile(source, target string) (err error) {
	in, err := os.Open(source)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(target)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()

	fi, err := in.Stat()
	if err != nil {
		return
	}
	if err = os.Chmod(target, fi.Mode()); err != nil {
		return
	}

	return
}
