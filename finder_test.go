package main

import (
	"reflect"
	"testing"
)

func TestFind(t *testing.T) {
	cases := []struct {
		directory string
		expected  []string
	}{
		{
			directory: "testdata/t1",
			expected:  []string{"testdata/t1/foo.css"},
		},
		{
			directory: "testdata/t4",
			expected:  []string{},
		},
	}
	for index, c := range cases {
		actual := Finder{}.Find(c.directory, Filter)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf(
				"Case: #%d - directory: %s\n"+
					"  actual: %#v\n"+
					"expected: %#v\n",
				index, c.directory, actual, c.expected)
		}
	}
}
