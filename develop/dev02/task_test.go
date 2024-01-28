package main

import "testing"

func TestUnpack(t *testing.T) {
	cases := map[string]string{
		"a4bc2d5e": "aaaabccddddde",
		"abcd":     "abcd",
		"45":       "",
		"":         "",
		`qwe\4\5`:  "qwe45",
		`qwe\45`:   "qwe44444",
		`qwe\\5`:   `qwe\\\\\`,
	}

	for input, expected := range cases {
		actual, _ := Unpack(input)
		if expected != actual {
			t.Errorf("unexpected result for input \"%s\": \"%s\" (expected \"%s\")", input, actual, expected)
		}
	}
}
