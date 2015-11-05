package main

import "testing"

func TestReplace(t *testing.T) {
	cases := []struct {
		in, search, replace, want string
	}{
		{"foobar", "foo", "bar", "barbar"},
	}
	for _, c := range cases {
		got := Replace{Search: c.search, Replace: c.replace}.Execute(c.in)
		if got != c.want {
			t.Errorf("Replace(%v, %v, %v) == %v, want %v", c.in, c.search, c.replace, got, c.want)
		}
	}
}
