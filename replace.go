package main

import "regexp"

const LF = 10

type Replace struct {
	Search, Replace string
}

func (r Replace) Execute(in string, callback ReplaceCallback) string {
	result := ""
	remainder := in
	search := regexp.QuoteMeta(r.Search)
	rgx := regexp.MustCompile(search)
	for {
		match := rgx.FindStringSubmatchIndex(remainder)
		if match == nil {
			result += remainder
			break
		}

		replacement := []byte{}
		replacement = rgx.ExpandString(replacement, r.Replace, in, match)

		if callback != nil && !callback(newReplaceInfo(result, remainder, string(replacement), match)) {
			result += remainder[0:match[1]]
		} else {
			result += remainder[0:match[0]] + string(replacement)
		}

		remainder = remainder[match[1]:]
	}
	return result
}

type ReplaceCallback func(info ReplaceInfo) bool

type ReplaceInfo struct {
	LinesBeforeMatch                string
	MatchLine                       string
	MatchLineMatchIndex             []int
	ReplacementLine                 string
	ReplacementLineReplacementIndex []int
	LinesAfterMatch                 string
}

func newReplaceInfo(result, remainder, replacement string, match []int) ReplaceInfo {
	matchOffset := len(result)
	matchStart := match[0] + matchOffset
	matchEnd := match[1] + matchOffset

	totalString := result + remainder

	// find line start index (excluding linefeed)
	lineStartIndex := matchStart
	for lineStartIndex >= 0 && totalString[lineStartIndex] != LF {
		lineStartIndex--
	}
	lineStartIndex++

	// find line end index (including linefeed)
	lineEndIndex := matchEnd
	for lineEndIndex < len(totalString) {
		if totalString[lineEndIndex] == LF {
			lineEndIndex++
			break
		}
		lineEndIndex++
	}

	matchLine := totalString[lineStartIndex:lineEndIndex]
	replacementLine := matchLine[:matchStart-lineStartIndex] + replacement + matchLine[matchEnd-lineStartIndex:]

	matchLineMatchIndex := []int{matchStart - lineStartIndex, matchEnd - lineStartIndex}
	replacementLineReplacementIndex := []int{matchStart - lineStartIndex, matchStart - lineStartIndex + len(replacement)}

	linesBeforeMatch := ""
	index := lineStartIndex
	count := 0
	for index > 0 {
		index--
		if totalString[index] == LF {
			count++
			if count == 4 {
				index++
				linesBeforeMatch = totalString[index:lineStartIndex]
				break
			}
		}
	}

	linesAfterMatch := ""
	index = lineEndIndex
	count = 0
	for index < len(totalString)-1 {
		index++
		if totalString[index] == LF {
			count++
			if count == 3 {
				linesAfterMatch = totalString[lineEndIndex : index+1]
				break
			}
		}
	}

	return ReplaceInfo{
		LinesBeforeMatch:                linesBeforeMatch,
		MatchLine:                       matchLine,
		MatchLineMatchIndex:             matchLineMatchIndex,
		ReplacementLine:                 replacementLine,
		ReplacementLineReplacementIndex: replacementLineReplacementIndex,
		LinesAfterMatch:                 linesAfterMatch,
	}
}
