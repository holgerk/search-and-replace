package main

// This is an integration Test which runs against real directories and
// compares the transformation against golden directories, containing
// the expected content and structure.
//
// This was heavily inspired by oracle_test.go, thanks.

import (
	"bytes"
	"flag"
	"fmt"
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
		referenceDir string
		dryRun       bool
	}{
		{"testdata/t1", false},
		{"testdata/t2", false},
		{"testdata/t1", true},
	}
	for _, c := range cases {
		referenceDir := c.referenceDir
		workingDir := referenceDir + ".got"
		goldenDir := referenceDir + ".golden"

		os.RemoveAll(workingDir)
		err := copyDirectory(referenceDir, workingDir)
		if err != nil {
			panic(err)
		}

		run(workingDir, c.dryRun, false)

		// compare directory.got with directory.golden
		compareDir := goldenDir
		if c.dryRun {
			compareDir = referenceDir
		}
		cmd := exec.Command("diff", "-ru", workingDir, compareDir)
		cmd.Stdout = new(bytes.Buffer)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			t.Errorf("Test for %s (dryRun: %v) failed: %s.\n%s\n",
				workingDir, c.dryRun, err, cmd.Stdout)

			if *updateFlag {
				t.Logf("Updating golden: %s...", goldenDir)
				os.RemoveAll(goldenDir)
				err := copyDirectory(workingDir, goldenDir)
				if err != nil {
					t.Errorf("Update failed: %s", err)
				}
			}
		}
	}
}

func TestPermissionProblems(t *testing.T) {
	stdout := run("testdata/t3", false, true)
	os.Chmod("testdata/t3/foo-not-writable.txt", 0444)
	if !strings.Contains(stdout, "Could not write: testdata/t3/foo-not-writable") {
		t.Errorf("Missing [Could not write...] message in(%s)", stdout)
	}
	if !strings.Contains(stdout, "Could not move: testdata/t3/foo-not-moveable") {
		t.Errorf("Missing [Could not open...] message in(%s)", stdout)
	}
	// fmt.Println(stdout)
}

func run(workingDir string, dryRun, verbose bool) string {
	var stdout bytes.Buffer
	program := Program{
		RootDirectory: workingDir,
		Search:        "foo",
		Replace:       "bar",
		Stdout:        &stdout,
		DryRun:        dryRun,
		Verbose:       verbose,
	}
	err := program.Execute()
	if err != nil {
		panic(fmt.Errorf("Program-Execution error(%s)", err))
	}

	return stdout.String()
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
	if err != nil {
		return fmt.Errorf("Could not copy directory(%s)", err)
	}
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
