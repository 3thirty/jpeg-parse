package jpeg

import "fmt"

// Marker represents a JPEG marker.
type Marker byte

// JPEG markers.
const (
	SOI   Marker = 0xD8 // Start of image
	EOI   Marker = 0xD9 // End of image
	SOS   Marker = 0xDA // Start of scan
	DQT   Marker = 0xDB // Define quantization table
	DHT   Marker = 0xC4 // Define Huffman table
	APP0  Marker = 0xE0 // APPn segments
	APP1  Marker = 0xE1
	APP2  Marker = 0xE2
	APP3  Marker = 0xE3
	APP4  Marker = 0xE4
	APP5  Marker = 0xE5
	APP6  Marker = 0xE6
	APP7  Marker = 0xE7
	APP8  Marker = 0xE8
	APP9  Marker = 0xEA
	APP10 Marker = 0xEB
	APP11 Marker = 0xEC
	APP12 Marker = 0xED
	APP13 Marker = 0xEE
	APP14 Marker = 0xEF
	SOF   Marker = 0xC0 // Start of frame
)

// markerNames maps markers to a human-readable string
var markerNames = map[Marker]string{
	SOI: "Start Of Image",
	EOI: "End Of Image",
	SOS: "Start Of Scan",
	DQT: "Define Quantzation Table",
	DHT: "Define Huffman Table",
	SOF: "Start Of Frame",
}

// String returns the string representation of the Marker.
func (m Marker) String() string {
	if name, ok := markerNames[m]; ok {
		return name
	}

	if (m >= APP0) && (m <= APP14) {
		return fmt.Sprintf("APP%d", int(m-APP0))
	}

	return fmt.Sprintf("UnknownMarker(0x%X)", byte(m))
}

// Byte returns the byte of the Marker
func (m Marker) Byte() byte {
	return byte(m)
}

// HasLength returns true if the marker has a length field following it.
func (m Marker) HasLength() bool {
	// Markers that do not have a length field
	switch m {
	case SOI, EOI:
		return false
	}
	return true
}
