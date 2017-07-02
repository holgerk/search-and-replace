package main

import (
	"reflect"
	"testing"
)

type FilterStub struct {
	response bool
}

func (f FilterStub) Filter(path string) bool {
	return f.response
}

func TestFind(t *testing.T) {
	cases := []struct {
		directory string
		expected  []string
		filterAll bool
	}{
		{
			directory: "testdata/t1",
			expected:  []string{"testdata/t1/foo.css"},
			filterAll: false,
		},
		{
			directory: "testdata/t1",
			expected:  []string{}, // <- nothing because filterAll is true
			filterAll: true,
		},
		{
			directory: "testdata/t4",
			expected:  []string{},
			filterAll: false,
		},
	}
	for index, c := range cases {
		actual := (&Finder{filter: &FilterStub{c.filterAll}}).Find(c.directory)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf(
				"Case: #%d - directory: %s\n"+
					"  actual: %#v\n"+
					"expected: %#v\n",
				index, c.directory, actual, c.expected)
		}
	}
}
