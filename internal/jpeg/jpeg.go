package jpeg

import (
	"fmt"
	"os"
)

type JpegFile struct {
	*os.File
}

type Field struct {
	offset int64
	name   string
	length int64
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
	pos, err := jpeg.findMarker(0xD8, 0, true)

	if err != nil {
		fmt.Println("Error seeking to SOI marker:", err)
		return false
	}

	return pos == 0
}

func (jpeg JpegFile) HasEOI() bool {
	info, err := jpeg.Stat()

	if err != nil {
		return false
	}

	fileSize := info.Size()

	eoi, err := jpeg.getEOIOffset()

	return eoi.offset == fileSize-2
}

func (jpeg JpegFile) getEOIOffset() (Field, error) {
	// EOI is 0xFF, 0xD9
	// for now, let's naively assume that it's the last 2 bytes of the file

	info, err := jpeg.Stat()

	if err != nil {
		return Field{}, err
	}

	fileSize := info.Size()

	pos, err := jpeg.findMarker(0xD9, fileSize-2, true)

	if err != nil {
		return Field{}, err
	}

	return Field{pos, "End of Image", 2}, nil
}

func (jpeg JpegFile) GetAppData() []Field {
	var ret []Field
	var offset = int64(2) // because SOI marker is fixed length, we can just start searching from offset 2

	markers := []byte{0xE0, 0xE1, 0xE2, 0xE3, 0xE4, 0xE5, 0xE6, 0xE7, 0xE8, 0xE9, 0xEA, 0xEB, 0xEC, 0xED, 0xEE, 0xEF}

	// App data markers are 0xFF 0xE0 to 0xFF 0xEF
	var readBuf = make([]byte, 2)
	for _, marker := range markers {
		pos, err := jpeg.findMarker(marker, offset, true)

		// if we hit EOF or other errors, that's fine, just move on to the next marker
		if err != nil {
			continue
		}

		// read the length, this will also move the file pointer forward by 4 bytes
		length := jpeg.getSegmentLength(pos + 2)

		// read the name field
		var name string
		for {
			if _, err = jpeg.Read(readBuf); err != nil {
				break
			}

			// if we hit a null byte, we're done
			if readBuf[0] == 0x00 {
				break
			}

			// if the last byte is null, we want to include the first byte only
			if readBuf[1] == 0x00 {
				name += string(readBuf[0])
				break
			}

			// otherwise we want both bytes
			name += string(readBuf)
		}

		ret = append(ret, Field{int64(pos), name, length})

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

	return bytesToInt64(buf)
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

	return bytesToInt64(buf)
}

func (jpeg JpegFile) GetSOS() (Field, error) {
	// SOS marker is 0xFF 0xDA
	pos, err := jpeg.findMarker(0xDA, int64(3), false) // TODO: 3 is not the right offset; but example image must have an earlier field with an odd number of bytes

	if err != nil {
		fmt.Println("Error seeking to SOS marker:", err)
		return Field{}, err
	}

	return Field{pos, "Start of Scan", jpeg.getSegmentLength(pos + 2)}, nil
}

func (jpeg JpegFile) GetCompressedImageData() []byte {
	sos, err := jpeg.GetSOS()

	if err != nil {
		return nil
	}

	// we identify the start of the compressed image data by finding the SOS marker (0xFF 0xDA)
	// skipping and skipping the length of that field
	jpeg.Seek(sos.offset+sos.length, 0)

	// we will read until we hit the EOI marker (0xFF 0xD9)
	// note that this does not support progressive jpegs
	eoi, err := jpeg.getEOIOffset()

	if err != nil {
		return nil
	}

	var buf = make([]byte, eoi.offset-(sos.offset-1000))

	jpeg.Seek(sos.offset+sos.length, 0)
	if _, err = jpeg.Read(buf); err != nil {
		return nil
	}

	return buf
}

func (jpeg JpegFile) HasSOF() bool {
	pos, err := jpeg.getSOFOffset()

	if err != nil {
		return false
	}

	return pos != -1
}

func (jpeg JpegFile) GetDQT() []Field {
	pos, err := jpeg.findMarker(0xDB, int64(2), false)

	if err != nil {
		return nil
	}

	length := jpeg.getSegmentLength(pos + 2)

	return []Field{{pos, "Define Quantization Table", length}}
}

func (jpeg JpegFile) getSOFOffset() (int64, error) {
	markers := []byte{0xC0, 0xC1, 0xC2, 0xC3, 0xC4, 0xC5, 0xC6, 0xC7, 0xC8, 0xC9, 0xCA, 0xCB, 0xCC, 0xCD, 0xCC, 0xCF}

	for _, marker := range markers {
		pos, err := jpeg.findMarker(marker, int64(2), false)

		// if we hit EOF or other errors, that's fine, just move on to the next marker
		if err != nil {
			continue
		}

		// just return the first SOF marker we find
		return pos, nil
	}

	return -1, fmt.Errorf("No SOF marker found")
}

func (jpeg JpegFile) findMarker(marker byte, offset int64, trustOffset bool) (int64, error) {
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

		// if we are trusting the offset, we can stop if we hit a different marker
		// 0xFF and 0xBB seem to not be markers, so ignore them. TODO: confirm this
		if trustOffset && buf[0] == 0xFF && buf[1] != marker && buf[1] != 0xFF && buf[1] != 0xBB {
			return -1, fmt.Errorf("Found unexpected marker %x %x", buf[0], buf[1])
		}
	}
}

func (jpeg JpegFile) getSegmentLength(offset int64) int64 {
	// read the subsequent two bytes to get the length of the segment
	var readBuf = make([]byte, 2)

	jpeg.Seek(offset, 0)

	_, err := jpeg.Read(readBuf)

	if err != nil {
		return -1
	}

	return bytesToInt64(readBuf)
}

func bytesToInt64(b []byte) int64 {
	if len(b) < 2 {
		return -1
	}
	return int64(b[0])<<8 | int64(b[1])
}
