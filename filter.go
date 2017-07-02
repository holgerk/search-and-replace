package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sabhiram/go-git-ignore"
)

type Filterer interface {
	Filter(s string) bool
}

var directoryBlacklist = map[string]bool{
	".git":      true,
	".hg":       true,
	".svn":      true,
	".DS_Store": true,
}

type Filter struct {
	rootDirectory string
	gitIgnore     *ignore.GitIgnore
}

func NewFilter(rootDirectory string) *Filter {
	gitIgnore, err := ignore.CompileIgnoreLines()
	if err != nil {
		panic(err)
	}

	ignoreFilePath := filepath.Join(rootDirectory, ".gitignore")
	if fi, err := os.Stat(ignoreFilePath); err == nil && fi.IsDir() == false {
		gitIgnore, err = ignore.CompileIgnoreFile(ignoreFilePath)
		if err != nil {
			panic(err)
		}
	}

	return &Filter{
		rootDirectory: rootDirectory,
		gitIgnore:     gitIgnore,
	}
}

func (f *Filter) Filter(path string) bool {
	basename := filepath.Base(path)
	if directoryBlacklist[basename] {
		return true
	}
	if f.gitIgnore.MatchesPath(f.shortenPath(path)) {
		return true
	}
	return false
}

func (f *Filter) shortenPath(path string) string {
	return strings.Replace(path, f.rootDirectory+"/", "", 1)
}
