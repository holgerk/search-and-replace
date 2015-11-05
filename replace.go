package main

import "strings"

type Replace struct {
	Search, Replace string
}

func (r Replace) Execute(in string) string {
	return strings.Replace(in, r.Search, r.Replace, -1)
}
