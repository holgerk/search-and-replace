package main

import "path/filepath"

func Filter(path string) bool {
	basename := filepath.Base(path)
	if basename == ".git" {
		return true
	}
	return false
}
