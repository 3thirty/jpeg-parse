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

	return pos == 0
}

func (jpeg JpegFile) HasEOI() bool {
	// EOI is 0xFF, 0xD9
	// for now, let's naively assume that it's the last 2 bytes of the file

	info, err := jpeg.Stat()

	if err != nil {
		return false
	}

	fileSize := info.Size()

	pos, err := jpeg.findMarker(0xD9, fileSize-2)

	if err != nil {
		fmt.Println("Error seeking to EOI marker:", err)
		return false
	}

	return pos == fileSize-2
}

func (jpeg JpegFile) GetAppData() []int64 {
	var ret []int64
	var offset = int64(2) // because SOI marker is fixed length, we can just start searching from offset 2

	markers := []byte{0xE0, 0xE1, 0xE2, 0xE3, 0xE4, 0xE5, 0xE6, 0xE7, 0xE8, 0xE9, 0xEA, 0xEB, 0xEC, 0xED, 0xEE, 0xEF}

	// App data markers are 0xFF 0xE0 to 0xFF 0xEF
	for _, marker := range markers {
		pos, err := jpeg.findMarker(marker, offset)

		// if we hit EOF or other errors, that's fine, just move on to th next marker
		if err != nil {
			continue
		}

		//fmt.Printf("Found marker 0x%X at position %d\n", marker, pos)

		ret = append(ret, pos)

		offset = pos + 2
	}

	return ret
}

func (jpeg JpegFile) GetHeight() int64 {
	offset, err := jpeg.getSOFOffset()

	if err != nil {
		return int64(-1)
	}

	var buf = make([]byte, 2)

	// height is stored in bytes 5 and 6 of the SOF segment
	jpeg.Seek(offset+5, 0)

	if _, err = jpeg.Read(buf); err != nil {
		return int64(-1)
	}

	return int64(buf[0])<<8 | int64(buf[1])
}

func (jpeg JpegFile) GetWidth() int64 {
	offset, err := jpeg.getSOFOffset()

	if err != nil {
		return int64(-1)
	}

	var buf = make([]byte, 2)

	// width is stored in bytes 7 and 8 of the SOF segment
	jpeg.Seek(offset+7, 0)

	if _, err = jpeg.Read(buf); err != nil {
		return int64(-1)
	}

	return int64(buf[0])<<8 | int64(buf[1])
}

func (jpeg JpegFile) getSOFOffset() (int64, error) {
	return jpeg.findMarker(0xC0, int64(2))
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
			return pos - 2, nil // Return position of the marker
		}
	}
}
