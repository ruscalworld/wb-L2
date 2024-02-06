package main

import (
	"reflect"
	"testing"
)

func TestExtractResources(t *testing.T) {
	tests := map[string][]string{
		`<a href="https://google.com"></a>`:                              {"https://google.com"},
		`<link rel="stylesheet" href="/assets/main.css">`:                {"/assets/main.css"},
		`<link rel="stylesheet" href="main.css">`:                        {"main.css"},
		`<link rel='stylesheet' href='/assets/main.css'>`:                {"/assets/main.css"},
		`<link rel='stylesheet' href='main.css'>`:                        {"main.css"},
		`<script type="text/javascript" src="/assets/main.js"></script>`: {"/assets/main.js"},
		`<script type="text/javascript" src="main.js"></script>`:         {"main.js"},
		`<script type='text/javascript' src='/assets/main.js'></script>`: {"/assets/main.js"},
		`<script type='text/javascript' src='main.js'></script>`:         {"main.js"},
	}

	for input, expected := range tests {
		actual := ExtractResources([]byte(input))

		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("unexpected result %v for test %s (expected %v)", actual, input, expected)
		}
	}
}
