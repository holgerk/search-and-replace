package main

// This is an integration Test which runs against real directories and
// compares the transformation against directories containing
// the expected content and structure (golden directories).
//
// This was heavily inspired by oracle_test.go, thx.

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

func TestMainExecute(t *testing.T) {
	cases := []struct {
		referenceDir string
		search       string
		dryRun       bool
		regexp       bool
	}{
		{
			referenceDir: "testdata/t1",
			search:       "foo",
			dryRun:       false,
			regexp:       false,
		},
		{
			referenceDir: "testdata/t2",
			search:       "foo",
			dryRun:       false,
			regexp:       false,
		},
		{
			referenceDir: "testdata/t1",
			search:       "foo",
			dryRun:       true,
			regexp:       false,
		},
		{
			referenceDir: "testdata/t1",
			search:       "fo{2}",
			dryRun:       false,
			regexp:       true,
		},
	}
	for index, c := range cases {
		referenceDir := c.referenceDir
		workingDir := referenceDir + ".got"
		goldenDir := referenceDir + ".golden"

		os.RemoveAll(workingDir)
		copyDirectory(referenceDir, workingDir)

		args := []string{}
		if c.dryRun {
			args = append(args, "--dry-run")
		}
		if c.regexp {
			args = append(args, "--regexp")
		}
		args = append(args, c.search)
		args = append(args, "bar")
		run(workingDir, []string{}, args)
		// fmt.Println(output)

		// compare directory.got with directory.golden
		compareDir := goldenDir
		if c.dryRun {
			compareDir = referenceDir
		}
		compare(t, index, compareDir, workingDir)

		if *updateFlag {
			t.Logf("Updating golden: %s...", goldenDir)
			os.RemoveAll(goldenDir)
			copyDirectory(workingDir, goldenDir)
		}
	}
}

func TestNotMovableFile(t *testing.T) {
	workingDir := "testdata/t3"
	os.Chmod(workingDir, 0555)
	stdout := run(workingDir, []string{}, []string{"foo", "bar"})
	assertContains(t, stdout, "Could not move: foo-not-moveable")
}

func TestNotCompilableRegexp(t *testing.T) {
	stdout := run("testdata/t3", []string{}, []string{"--regexp", "(", "bar"})
	assertContains(t, stdout, "Could not compile regular expression: (")
}

func TestHelpFlag(t *testing.T) {
	stdout := run("testdata/t3", []string{}, []string{"--help"})
	assertContains(t, stdout, "Application Options")
}

func TestInteractiveMode(t *testing.T) {
	cases := []struct {
		referenceDir string
		expectedDir  string
		answers      []string
	}{
		{
			referenceDir: "testdata/t1",
			expectedDir:  "testdata/t1.golden",
			// provide 4 yes-answers
			answers: []string{"y\n", "Y\n", "\n", "\n"},
		},
		{
			referenceDir: "testdata/t1",
			expectedDir:  "testdata/t1",
			// provide 4 no-answers
			answers: []string{"n\n", "N\n", "w\n", "P\n"},
		},
	}
	for index, c := range cases {
		workingDir := c.referenceDir + ".got"

		os.RemoveAll(workingDir)
		copyDirectory(c.referenceDir, workingDir)

		run(workingDir, c.answers, []string{"--interactive", "foo", "bar"})

		compare(t, index, c.expectedDir, workingDir)
	}
}

func compare(t *testing.T, index int, compareDir, workingDir string) {
	cmd := exec.Command("diff", "-ru", workingDir, compareDir)
	cmd.Stdout = new(bytes.Buffer)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Errorf(
			"Case #%d - workingDir: %s, compareDir: %v failed: %s.\n%s\n",
			index, workingDir, compareDir, err, cmd.Stdout)
	}
}

func run(workingDir string, stdinStr, args []string) string {
	stdin := &StringReader{data: stdinStr}
	var stdout bytes.Buffer
	mainSub(workingDir, &stdout, stdin, args)

	return stdout.String()
}

func copyDirectory(source, target string) {
	err := filepath.Walk(source, func(path string, f os.FileInfo, err error) (werr error) {
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
		panic(fmt.Errorf("Could not copy directory(%s)", err))
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

// copied from standard library (bufio_test.go) - don't how to import :-(
// A StringReader delivers its data one string segment at a time via Read.
type StringReader struct {
	data []string
	step int
}

func (r *StringReader) Read(p []byte) (n int, err error) {
	if r.step < len(r.data) {
		s := r.data[r.step]
		n = copy(p, s)
		r.step++
	} else {
		err = io.EOF
	}
	return
}

func assertContains(t *testing.T, haystack, needle string) {
	if !strings.Contains(haystack, needle) {
		t.Errorf("Expected string(%s) in(%s)", needle, haystack)
	}
}
