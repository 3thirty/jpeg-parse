package jpeg

import (
	"testing"
)

func TestMarkerToString(t *testing.T) {
	var m Marker = DQT

	got := m.String()
	want := "Define Quantzation Table"

	if got != want {
		t.Errorf("Marker.String() = %q; want %q", got, want)
	}
}

func TestConstToByte(t *testing.T) {
	var m Marker = SOS

	got := m.Byte()
	want := byte(0xDA)

	if got != want {
		t.Errorf("Marker %q; want %q", got, want)
	}
}
