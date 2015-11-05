package main

import (
	"os"
	"path/filepath"
)

func Find(searchDir string) []string {
	fileList := []string{}
	filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return nil
	})
	return fileList
}
