package main

import (
	"fmt"
	"io"
)

type Output struct {
	stdout  io.Writer
	verbose bool
}

func (o *Output) reportError(format string, a ...interface{}) {
	fmt.Fprintf(o.stdout, "[ERROR] "+format+"\n", a...)
}

func (o *Output) reportInfo(format string, a ...interface{}) {
	fmt.Fprintf(o.stdout, "[INFO] "+format+"\n", a...)
}

func (o *Output) reportReplacement(info ReplacementInfo) {
	o.print(info.LinesBeforeMatch)

	o.print(styleRed(info.MatchLine[:info.MatchLineMatchIndex[0]]))
	o.print(styleRedUnderline(info.Match))
	o.print(styleRed(info.MatchLine[info.MatchLineMatchIndex[1]:]))

	o.print(styleGreen(info.ReplLine[:info.ReplLineReplIndex[0]]))
	o.print(styleGreenUnderline(info.Repl))
	o.print(styleGreen(info.ReplLine[info.ReplLineReplIndex[1]:]))

	o.print(info.LinesAfterMatch)
}

func (o *Output) print(s string) {
	fmt.Fprint(o.stdout, s)
}

func (o *Output) printf(format string, a ...interface{}) {
	fmt.Fprintf(o.stdout, format, a...)
}

func (o *Output) printHeader(format string, a ...interface{}) {
	fmt.Fprintf(o.stdout, styleHeader("\n "+format)+"\n", a...)
}

func (o *Output) reportVerbose(format string, a ...interface{}) {
	if !o.verbose {
		return
	}
	o.reportInfo(format, a...)
}
