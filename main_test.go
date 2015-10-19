package main

import (
	"reflect"
	"sort"
	"testing"
)

func TestPrettyListify(t *testing.T) {
	cases := []struct {
		in   []string
		want string
	}{
		{
			[]string{"Alice", "Bob", "Charlie", "Dianne"},
			"Alice, Bob, Charlie and Dianne",
		},
		{
			[]string{"Adrian", "Brianna", "Charlotte"},
			"Adrian, Brianna and Charlotte",
		},
		{
			[]string{"Arnold", "Bertie"},
			"Arnold and Bertie",
		},
		{
			[]string{"Andy"},
			"Andy",
		},
		{
			[]string{},
			"",
		},
	}

	for _, c := range cases {
		sort.Strings(c.in) // iteration order is nondeterminant
		got := prettyListify(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("prettyListify(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
