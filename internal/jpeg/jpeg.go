package jpeg

import (
	"fmt"
	"os"
)

type JpegFile struct {
	*os.File
}

func Open(filename string) (JpegFile, error) {
	fh, err := os.Open(filename)

	if err != nil {
		return JpegFile{}, err
	}

	return JpegFile{fh}, nil
}

func (jpeg JpegFile) HasSOI() bool {
	// SOI is first two bytes: 0xFF, 0xD8
	pos, err := jpeg.findMarker(0xD8, 0)

	if err != nil {
		fmt.Println("Error seeking to SOI marker:", err)
		return false
	}

	return pos == 1
}

func (jpeg JpegFile) HasEOI() bool {
	// EOI is 0xFF, 0xD9
	// for now, let's naively assume that it's the last 2 bytes of the file

	info, err := jpeg.Stat()

	if err != nil {
		return false
	}

	fileSize := info.Size()

	pos, err := jpeg.findMarker(0xD9, 2)

	if err != nil {
		fmt.Println("Error seeking to EOI marker:", err)
		return false
	}

	return pos == fileSize-1
}

func (jpeg JpegFile) findMarker(marker byte, offset int64) (int64, error) {
	var err error
	var pos int64

	// Move to the specified offset
	if _, err = jpeg.Seek(offset, 0); err != nil {
		return -1, err
	}

	// Read until we find the marker or reach EOF
	for {
		buf := make([]byte, 2)
		if _, err = jpeg.Read(buf); err != nil {
			return -1, err
		}

		if buf[0] == 0xFF && buf[1] == marker {
			pos, err = jpeg.Seek(0, 1) // Get current position
			if err != nil {
				return -1, err
			}
			return pos - 1, nil // Return position of the marker
		}
	}
}
