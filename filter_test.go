package main

import "testing"

func TestFilter(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"t1", false},
		{"p/t1", false},
		{"p/.git/d", false},

		{"p/.git", true},
		{".git", true},
		{".hg", true},
		{".svn", true},
	}
	for _, c := range cases {
		filter := NewFilter("")
		got := filter.Filter(c.in)
		if got != c.want {
			t.Errorf("Filter(%v) == %v, want %v", c.in, got, c.want)
		}
	}
}
