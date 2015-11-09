package main

import "testing"

func TestReplace(t *testing.T) {
	cases := []struct {
		in, search, replace, want string
	}{
		{"foobar", "foo", "bar", "barbar"},
	}
	for _, c := range cases {
		got := Replace{Search: c.search, Replace: c.replace}.Execute(c.in, nil)
		if got != c.want {
			t.Errorf(
				"Replace{Search: %v, Replace: %v}.Execute(%v) == %v, want %v",
				c.search, c.replace, c.in, got, c.want)
		}
	}
}

func TestReplaceCallback(t *testing.T) {
	cases := []struct {
		callbackResult bool
		expected       string
	}{
		{false, "foobar"},
		{true, "barbar"},
	}
	for _, c := range cases {
		got := Replace{Search: "foo", Replace: "bar"}.Execute("foobar", func(info ReplacementInfo) bool {
			return c.callbackResult
		})
		if got != c.expected {
			t.Errorf(
				"callbackResult: %v, expected: %v, got: %v",
				c.callbackResult, c.expected, got)
		}
	}
}

func TestReplacementInfoContextLines(t *testing.T) {
	cases := []struct {
		content                  string
		expectedLinesBeforeMatch string
		expectedLinesAfterMatch  string
	}{
		{"foo", "", ""},
		{
			"line1\n" +
				"line2\n" +
				"line3\n" +
				"line4\n" +
				"line5 foo\n" +
				"line6\n" +
				"line7\n" +
				"line8\n" +
				"line9\n",

			"line2\n" +
				"line3\n" +
				"line4\n",

			"line6\n" +
				"line7\n" +
				"line8\n",
		},
		{
			"line1\n" +
				"line2 foo\n" +
				"line3\n",

			"line1\n",

			"line3\n",
		},
	}
	for index, c := range cases {
		Replace{Search: "foo", Replace: "bar"}.Execute(c.content, func(info ReplacementInfo) bool {
			if info.LinesBeforeMatch != c.expectedLinesBeforeMatch {
				t.Errorf(
					"Case: #%d - LinesBeforeMatch\n"+
						"  actual: %#v\n"+
						"expected: %#v\n",
					index, info.LinesBeforeMatch, c.expectedLinesBeforeMatch)
			}
			if info.LinesAfterMatch != c.expectedLinesAfterMatch {
				t.Errorf(
					"Case: #%d - LinesAfterMatch\n"+
						"  actual: %#v\n"+
						"expected: %#v\n",
					index, info.LinesAfterMatch, c.expectedLinesAfterMatch)
			}
			return true
		})
	}
}
