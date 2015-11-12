package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type Ask struct {
	Stdin  io.Reader
	Stdout io.Writer
}

func (a Ask) question(question string) bool {
	fmt.Fprintf(a.Stdout, "%s [Yn]: ", question)
	reader := bufio.NewReader(a.Stdin)
	reply, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	reply = strings.ToLower(strings.TrimSpace(reply))
	if reply == "y" || reply == "" {
		return true
	}
	return false
}
