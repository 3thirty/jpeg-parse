package jpeg

import (
	"testing"
)

var testFiles = []struct {
	file  string
	valid bool
}{
	{"../../test/data/minneapolis.jpg", true},
	{"../../test/data/bad.jpg", false},
}

func TestGetSOI(t *testing.T) {
	for _, test := range testFiles {
		jpeg, err := Open(test.file)

		if err != nil {
			t.Errorf("Error opening file %s: %v", test.file, err)
			return
		}

		if jpeg.HasSOI() != test.valid {
			t.Errorf("No SOI marker in file %s", test.file)
		}
	}
}

func TestGetEOI(t *testing.T) {
	for _, test := range testFiles {
		jpeg, err := Open(test.file)

		if err != nil {
			t.Errorf("Error opening file %s: %v", test.file, err)
			return
		}

		if jpeg.HasEOI() != test.valid {
			t.Errorf("No EOI marker in file %s", test.file)
		}
	}
}
