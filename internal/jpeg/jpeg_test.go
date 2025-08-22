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

func TestGetAppData(t *testing.T) {
	tests := []struct {
		file          string
		expectedCount int
	}{
		{"../../test/data/minneapolis.jpg", 3},
		{"../../test/data/bad.jpg", 0},
	}

	for _, test := range tests {
		jpeg, err := Open(test.file)

		if err != nil {
			t.Errorf("Error opening file %s: %v", test.file, err)
			return
		}

		appData := jpeg.GetAppData()

		if len(appData) != test.expectedCount {
			t.Errorf("Did not find expected number of app data file=%s expected=%d found=%d", test.file, test.expectedCount, len(appData))
		}
	}
}

func TestGetHeight(t *testing.T) {
	tests := []struct {
		file     string
		expected int64
	}{
		{"../../test/data/minneapolis.jpg", 600},
		{"../../test/data/bad.jpg", -1},
	}

	for _, test := range tests {
		jpeg, err := Open(test.file)

		if err != nil {
			t.Errorf("Error opening file %s: %v", test.file, err)
			return
		}

		height := jpeg.GetHeight()

		if height != test.expected {
			t.Errorf("Height mismatch match file=%s expected=%d found=%d", test.file, test.expected, height)
		}
	}
}

func TestGetWidth(t *testing.T) {
	tests := []struct {
		file     string
		expected int64
	}{
		{"../../test/data/minneapolis.jpg", 800},
		{"../../test/data/bad.jpg", -1},
	}

	for _, test := range tests {
		jpeg, err := Open(test.file)

		if err != nil {
			t.Errorf("Error opening file %s: %v", test.file, err)
			return
		}

		width := jpeg.GetWidth()

		if width != test.expected {
			t.Errorf("Width mismatch match file=%s expected=%d found=%d", test.file, test.expected, width)
		}
	}
}
