package main

import (
	"reflect"
	"testing"
)

type testCase struct {
	input    []string
	expected map[string][]string
}

func TestGroupAnagrams(t *testing.T) {
	var tests = []testCase{
		{
			input: []string{"пятак", "пятка", "тяпка", "листок", "слиток", "столик", "арбуз"},
			expected: map[string][]string{
				"пятак":  {"пятак", "пятка", "тяпка"},
				"листок": {"листок", "слиток", "столик"},
			},
		},
	}

	for _, c := range tests {
		groups := GroupAnagrams(c.input)
		if !reflect.DeepEqual(*groups, c.expected) {
			t.Errorf("unexpected result: %v (expected %v)", *groups, c.expected)
		}
	}
}
