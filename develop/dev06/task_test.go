package main

import (
	"reflect"
	"testing"
)

type testCase[I any, E any] struct {
	input    I
	expected E
}

func TestParseRange(t *testing.T) {
	tests := []testCase[string, Range]{
		{
			input:    "1",
			expected: NewRange(NewSpecificRangeBound(1)),
		},
		{
			input:    "1-3",
			expected: NewRange(NewSpecificRangeBound(1), NewSpecificRangeBound(3)),
		},
		{
			input:    "-",
			expected: NewRange(NewRangeBound(ConstraintMin), NewRangeBound(ConstraintMax)),
		},
		{
			input:    "1-",
			expected: NewRange(NewSpecificRangeBound(1), NewRangeBound(ConstraintMax)),
		},
		{
			input:    "-3",
			expected: NewRange(NewRangeBound(ConstraintMin), NewSpecificRangeBound(3)),
		},
	}

	for _, c := range tests {
		actual, err := ParseRange(c.input)
		if err != nil {
			t.Errorf("error in test for \"%s\": %s", c.input, err)
			return
		}

		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("unexpected result in test for \"%s\": %v (expected %v)", c.input, actual, c.expected)
		}
	}
}

func TestParseColumnSelector(t *testing.T) {
	tests := []testCase[string, ColumnSelector]{
		{
			input: "1-3,5-9",
			expected: NewColumnSelector(
				NewRange(NewSpecificRangeBound(1), NewSpecificRangeBound(3)),
				NewRange(NewSpecificRangeBound(5), NewSpecificRangeBound(9)),
			),
		},
		{
			input: "-3,5-",
			expected: NewColumnSelector(
				NewRange(NewRangeBound(ConstraintMin), NewSpecificRangeBound(3)),
				NewRange(NewSpecificRangeBound(5), NewRangeBound(ConstraintMax)),
			),
		},
	}

	for _, c := range tests {
		actual, err := ParseColumnSelector(c.input)
		if err != nil {
			t.Errorf("error in test for \"%s\": %s", c.input, err)
			return
		}

		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("unexpected result in test for \"%s\": %v (expected %v)", c.input, actual, c.expected)
		}
	}
}

type getColumnsInput struct {
	selector string
	max      int
}

func TestColumnSelector_GetColumns(t *testing.T) {
	tests := []testCase[getColumnsInput, []int]{
		{
			input: getColumnsInput{
				selector: "1-3,5-9",
				max:      8,
			},
			expected: []int{1, 2, 3, 5, 6, 7, 8},
		},
		{
			input: getColumnsInput{
				selector: "-2,6-",
				max:      7,
			},
			expected: []int{1, 2, 6, 7},
		},
		{
			input: getColumnsInput{
				selector: "2-,5-",
				max:      7,
			},
			expected: []int{2, 3, 4, 5, 6, 7},
		},
	}

	for _, c := range tests {
		selector, err := ParseColumnSelector(c.input.selector)
		if err != nil {
			t.Errorf("error in test for \"%s\": %s", c.input.selector, err)
			return
		}

		actual := selector.GetColumns(c.input.max)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("unexpected result in test for \"%s\": %v (expected %v)", c.input.selector, actual, c.expected)
		}
	}
}
