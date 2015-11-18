package main

import (
	"os"
	"path/filepath"
)

type Finder struct {
	output *Output
}

func (f *Finder) Find(searchDir string, filterFunc FilterFunc) []string {
	fileList := []string{}
	filepath.Walk(searchDir, func(path string, fi os.FileInfo, err error) error {
		if path == searchDir {
			return nil
		}
		if fi == nil {
			f.output.reportError("Could not read fileinfo: " + path)
			return nil
		}
		if filterFunc(path) {
			if fi.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}
		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			return filepath.SkipDir
		}
		fileList = append(fileList, path)
		return nil
	})
	return fileList
}

type FilterFunc func(path string) bool
