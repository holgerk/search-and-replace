package main

import (
	"reflect"
	"testing"
)

func TestFind(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"testdata/t1", []string{"testdata/t1/foo.css"}},
	}
	for _, c := range cases {
		got := Find(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("Finder(%v) == %v, want %v", c.in, got, c.want)
		}
	}
}
