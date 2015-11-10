package main

import "testing"

func TestReplace(t *testing.T) {
	cases := []struct {
		content, search, replace, expected string
	}{
		{"foobar", "foo", "bar", "barbar"},
	}
	for _, c := range cases {
		actual := Replace{Search: c.search, Replace: c.replace}.Execute(c.content, nil)
		if actual != c.expected {
			t.Errorf(
				"Replace{Search: %v, Replace: %v}.Execute(%v) == %v, expected %v",
				c.search, c.replace, c.content, actual, c.expected)
		}
	}
}

func TestReplaceRegexp(t *testing.T) {
	cases := []struct {
		content, search, replace, expected string
	}{
		{
			content:  "foobar",
			search:   "fo+",
			replace:  "bar",
			expected: "barbar",
		},
		{
			content:  "foobar",
			search:   "(...)(...)",
			replace:  "$2$1",
			expected: "barfoo",
		},
	}
	for index, c := range cases {
		replace := Replace{Search: c.search, Replace: c.replace, Regexp: true}
		actual := replace.Execute(c.content, nil)
		if actual != c.expected {
			t.Errorf(
				"Case: #%d - content: %s, search: %s, replace: %s\n"+
					"  actual: %#v\n"+
					"expected: %#v\n",
				index, c.content, c.search, c.replace, actual, c.expected)
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
		actual := Replace{Search: "foo", Replace: "bar"}.Execute("foobar", func(info ReplacementInfo) bool {
			return c.callbackResult
		})
		if actual != c.expected {
			t.Errorf(
				"callbackResult: %v, expected: %v, actual: %v",
				c.callbackResult, c.expected, actual)
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
