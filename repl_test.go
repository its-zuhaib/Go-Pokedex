package main

import "testing"

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    " hello world",
			expected: []string{"hello", "world"},
		},
	}
	for _, c := range cases {
		actual := cleanInput(c.input)
		if len(actual) != len(c.expected) {
			t.Errorf("Length mismatch")
		}

		for i := range actual {
			if c.expected[i] != actual[i] {
				t.Errorf("word mismatch")
			}
		}
	}
}
