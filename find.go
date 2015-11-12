package main

import (
	"os"
	"path/filepath"
)

func Find(searchDir string, filterFunc FilterFunc) []string {
	fileList := []string{}
	filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		if path == searchDir {
			return nil
		}
		if f == nil {
			// TODO report fileinfo not readable error
			return nil
		}
		if filterFunc(path) {
			if f.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}
		if f.Mode()&os.ModeSymlink == os.ModeSymlink {
			return filepath.SkipDir
		}
		fileList = append(fileList, path)
		return nil
	})
	return fileList
}

type FilterFunc func(path string) bool
