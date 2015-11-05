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
		if filterFunc(path) {
			return filepath.SkipDir
		}
		fileList = append(fileList, path)
		return nil
	})
	return fileList
}

type FilterFunc func(path string) bool
