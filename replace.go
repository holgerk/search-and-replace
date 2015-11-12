package main

import "regexp"

const LineFeed = 10
const ContextLineCount = 3

type Replace struct {
	Search, Replace string
	Regexp          bool
}

func (r Replace) Execute(in string, callback ReplaceCallback) string {
	result := ""
	remainder := in
	search := r.Search
	if !r.Regexp {
		search = regexp.QuoteMeta(search)
	}
	rgx := regexp.MustCompile(search)

	var match []int
	replacement := []byte{}

	replacementInfo := func() ReplacementInfo {
		content := result + remainder
		matchOffset := len(result)
		matchStart := match[0] + matchOffset
		matchEnd := match[1] + matchOffset
		return newReplacementInfo(content, string(replacement), matchStart, matchEnd)
	}

	for {
		match = rgx.FindStringSubmatchIndex(remainder)
		if match == nil {
			result += remainder
			break
		}

		replacement = []byte{}
		replacement = rgx.ExpandString(replacement, r.Replace, in, match)

		if callback != nil && !callback(replacementInfo()) {
			result += remainder[0:match[1]]
		} else {
			result += remainder[0:match[0]] + string(replacement)
		}

		remainder = remainder[match[1]:]
	}
	return result
}

type ReplaceCallback func(info ReplacementInfo) bool

type ReplacementInfo struct {
	LinesBeforeMatch    string
	Match               string
	MatchLine           string
	MatchLineMatchIndex []int
	Repl                string
	ReplLine            string
	ReplLineReplIndex   []int
	LinesAfterMatch     string
}

func newReplacementInfo(content, replacement string, matchStart, matchEnd int) ReplacementInfo {
	lineStartIndex := findLineStartIndex(content, matchStart)
	lineEndIndex := findLineEndIndex(content, matchEnd)

	matchLine := content[lineStartIndex : lineEndIndex+1]
	lineContentBeforeMatch := matchLine[:matchStart-lineStartIndex]
	lineContentAfterMatch := matchLine[matchEnd-lineStartIndex:]
	replacementLine := lineContentBeforeMatch + replacement + lineContentAfterMatch

	matchLineMatchIndex := []int{
		matchStart - lineStartIndex,
		matchEnd - lineStartIndex,
	}
	replacementLineReplacementIndex := []int{
		matchStart - lineStartIndex,
		matchStart - lineStartIndex + len(replacement),
	}

	return ReplacementInfo{
		LinesBeforeMatch:    linesBeforeMatch(content, lineStartIndex),
		Match:               matchLine[matchLineMatchIndex[0]:matchLineMatchIndex[1]],
		MatchLine:           matchLine,
		MatchLineMatchIndex: matchLineMatchIndex,
		Repl:                replacementLine[replacementLineReplacementIndex[0]:replacementLineReplacementIndex[1]],
		ReplLine:            replacementLine,
		ReplLineReplIndex:   replacementLineReplacementIndex,
		LinesAfterMatch:     linesAfterMatch(content, lineEndIndex),
	}
}

func linesBeforeMatch(content string, lineStartIndex int) string {
	from := findPreviousLinesStartIndex(content, lineStartIndex, ContextLineCount)
	to := lineStartIndex
	return content[from:to]
}

func linesAfterMatch(content string, lineEndIndex int) string {
	if lineEndIndex == len(content)-1 {
		return ""
	}
	from := lineEndIndex
	to := findNextLinesEndIndex(content, lineEndIndex, ContextLineCount)
	if from < len(content)-1 {
		from++
	}
	return content[from : to+1]
}

func findLineStartIndex(content string, fromIndex int) int {
	index := fromIndex
	if index <= 0 {
		return 0
	}
	for index > 0 {
		index--
		if content[index] == LineFeed {
			return index + 1
		}
	}
	return 0
}

func findLineEndIndex(content string, fromIndex int) int {
	index := fromIndex
	if index >= len(content)-1 {
		return len(content) - 1
	}
	for index <= len(content)-1 {
		if content[index] == LineFeed {
			return index
		}
		index++
	}
	return len(content) - 1
}

func findPreviousLinesStartIndex(content string, fromIndex int, n int) int {
	// go to current lines start index
	index := findLineStartIndex(content, fromIndex)
	for i := 0; i < n; i++ {
		index = findLineStartIndex(content, index-1)
	}
	return index
}

func findNextLinesEndIndex(content string, fromIndex int, n int) int {
	// go to current lines end index
	index := findLineEndIndex(content, fromIndex)
	for i := 0; i < n; i++ {
		index = findLineEndIndex(content, index+1)
	}
	return index
}
