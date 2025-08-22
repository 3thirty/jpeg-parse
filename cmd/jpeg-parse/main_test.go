package main

import (
	"os"
	"testing"
)

func TestArgParsing(t *testing.T) {
	tests := []struct {
		input         string
		expected      string
		expectedError bool
	}{
		{"test.jpg", "test.jpg", false},
		{"missing.jpg", "", true},
		{"", "", true},
	}

	for _, test := range tests {
		// mock filesystem operations
		if test.expectedError {
			getFileName = func(name string) (string, error) {
				return "", os.ErrNotExist
			}
		} else {
			getFileName = func(name string) (string, error) {
				return name, nil
			}
		}

		res, err := parseArgs([]string{"parse-jpeg", test.input})

		if (test.expectedError && err == nil) || (!test.expectedError && err != nil) {
			t.Errorf("Input: %v, expected error: %v, got: %v", test.input, test.expectedError, err)
		}

		if test.expected != "" && res != test.expected {
			t.Errorf("Input: %v, Expected: %s, got: %s", test.input, test.expected, res)
		}
	}
}
