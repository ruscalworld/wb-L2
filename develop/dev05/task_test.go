package main

import (
	"bytes"
	"reflect"
	"testing"
)

type searchTestCase struct {
	search    *Search
	predicate func(string) bool
	expected  map[int]bool
}

func TestSearch_Find(t *testing.T) {
	tests := []searchTestCase{
		{
			search: &Search{
				sourceLines: []string{"2", "2", "1", "1", "2", "2"},
				matches:     make(map[int]bool),
				invert:      false,
			},
			predicate: func(s string) bool {
				return s == "1"
			},
			expected: map[int]bool{
				0: false, 1: false,
				2: true, 3: true,
				4: false, 5: false,
			},
		},
		{
			search: &Search{
				sourceLines: []string{"2", "2", "1", "1", "2", "2"},
				matches:     make(map[int]bool),
				invert:      true,
			},
			predicate: func(s string) bool {
				return s == "1"
			},
			expected: map[int]bool{
				0: true, 1: true,
				2: false, 3: false,
				4: true, 5: true,
			},
		},
	}

	for _, c := range tests {
		c.search.Find(c.predicate)
		if !reflect.DeepEqual(c.search.matches, c.expected) {
			t.Errorf("unexpected result: %v (expected %v)", c.search.matches, c.expected)
		}
	}
}

type printerTestCase struct {
	search   *Search
	printer  *Printer
	expected string
}

func TestPrinter_Print(t *testing.T) {
	tests := []printerTestCase{
		{
			search: &Search{
				sourceLines: []string{"2", "2", "1", "1", "2", "2"},
				matches: map[int]bool{
					0: false, 1: false,
					2: true, 3: true,
					4: false, 5: false,
				},
			},
			printer:  &Printer{},
			expected: "1\n1\n",
		},
		{
			search: &Search{
				sourceLines: []string{"2", "2", "1", "1", "2", "2"},
				matches: map[int]bool{
					0: false, 1: false,
					2: true, 3: true,
					4: false, 5: false,
				},
			},
			printer: &Printer{
				linesBefore: 1,
				linesAfter:  1,
			},
			expected: "2\n1\n1\n2\n",
		},
		{
			search: &Search{
				sourceLines: []string{"2", "2", "1", "1", "2", "2"},
				matches: map[int]bool{
					0: false, 1: false,
					2: true, 3: true,
					4: false, 5: false,
				},
			},
			printer: &Printer{
				linesBefore: 0,
				linesAfter:  1,
			},
			expected: "1\n1\n2\n",
		},
		{
			search: &Search{
				sourceLines: []string{"2", "2", "1", "1", "2", "2"},
				matches: map[int]bool{
					0: false, 1: false,
					2: true, 3: true,
					4: false, 5: false,
				},
			},
			printer: &Printer{
				linesBefore: 1,
				linesAfter:  0,
			},
			expected: "2\n1\n1\n",
		},
		{
			search: &Search{
				sourceLines: []string{"2", "2", "1", "1", "2", "2"},
				matches: map[int]bool{
					0: true, 1: true,
					2: true, 3: true,
					4: true, 5: true,
				},
			},
			printer: &Printer{
				linesBefore: 2,
				linesAfter:  2,
			},
			expected: "2\n2\n1\n1\n2\n2\n",
		},
	}

	for _, c := range tests {
		buf := &bytes.Buffer{}
		_, _ = c.printer.Print(c.search, buf)

		if buf.String() != c.expected {
			t.Errorf("unexpected result: %s (expected %s)", buf.String(), c.expected)
		}
	}
}
