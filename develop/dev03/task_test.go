package main

import (
	"bytes"
	"sort"
	"testing"
)

const input = "5 0\n10 1\n4 2\n3 3\n3 3\n2 4\n2 5"

type testCase struct {
	input    string
	expected string
	holder   *FileHolder
}

func TestFileHolder(t *testing.T) {
	tests := []testCase{
		{
			holder: &FileHolder{
				columnIndex: -1,
			},
			input:    input,
			expected: "10 1\n2 4\n2 5\n3 3\n3 3\n4 2\n5 0\n",
		},
		{
			holder: &FileHolder{
				columnIndex: -1,
				unique:      true,
			},
			input:    input,
			expected: "10 1\n2 4\n2 5\n3 3\n4 2\n5 0\n",
		},
		{
			holder: &FileHolder{
				columnIndex:  -1,
				reverseOrder: true,
			},
			input:    input,
			expected: "5 0\n4 2\n3 3\n3 3\n2 5\n2 4\n10 1\n",
		},
		{
			holder: &FileHolder{
				columnIndex:  -1,
				reverseOrder: true,
				unique:       true,
			},
			input:    input,
			expected: "5 0\n4 2\n3 3\n2 5\n2 4\n10 1\n",
		},
		{
			holder: &FileHolder{
				columnIndex: 1,
			},
			input:    input,
			expected: "5 0\n10 1\n4 2\n3 3\n3 3\n2 4\n2 5\n",
		},
		{
			holder: &FileHolder{
				columnIndex: 1,
				unique:      true,
			},
			input:    input,
			expected: "5 0\n10 1\n4 2\n3 3\n2 4\n2 5\n",
		},
		{
			holder: &FileHolder{
				columnIndex:  1,
				reverseOrder: true,
			},
			input:    input,
			expected: "2 5\n2 4\n3 3\n3 3\n4 2\n10 1\n5 0\n",
		},
	}

	for i, c := range tests {
		c.holder.ReadLines(bytes.NewReader([]byte(c.input)))
		sort.Sort(c.holder)

		buf := &bytes.Buffer{}
		_, err := c.holder.WriteOutput(buf)
		if err != nil {
			t.Errorf("error in test %d: %s", i, err)
		}

		if buf.String() != c.expected {
			t.Errorf("unexpected value in test %d:\n %s", i, buf.String())
		}
	}
}
