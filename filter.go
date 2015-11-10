package main

import "path/filepath"

var directoryBlacklist = map[string]bool{
	".git":      true,
	".hg":       true,
	".svn":      true,
	".DS_Store": true,
}

func Filter(path string) bool {
	basename := filepath.Base(path)
	return directoryBlacklist[basename]
}
