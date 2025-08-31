package jpeg

import (
	"fmt"
	"os"
)

// File represents a JPEG file on disk and provides methods to interact with it
type File struct {
	*os.File
	fields map[Marker]Field
}

// Field represents a segment or field within the JPEG file
type Field struct {
	Offset int64
	Name   string
	Length int64
}

// Open opens a JPEG file and returns a File struct
func Open(filename string) (File, error) {
	fh, err := os.Open(filename)

	if err != nil {
		return File{}, err
	}

	return File{fh, make(map[Marker]Field)}, nil
}

// HasSOI checks if the JPEG file has a valid Start of Image (SOI) marker
func (jpeg File) HasSOI() bool {
	_, err := jpeg.GetFieldByMarker(SOI)

	if err != nil {
		fmt.Println(err)
	}

	return err == nil
}

// HasEOI checks if the JPEG file has a valid End of Image (EOI) marker
func (jpeg File) HasEOI() bool {
	_, err := jpeg.GetFieldByMarker(EOI)

	return err == nil
}

// extractFieldByMarker returns a Field struct for a given JPEG marker
func (jpeg File) extractFieldByMarker(marker Marker) (Field, error) {
	pos, err := jpeg.findMarker(marker.Byte())

	if err != nil {
		return Field{}, err
	}

	if pos == -1 {
		return Field{}, fmt.Errorf("could not find marker %X: %s", marker.Byte(), marker.String())
	}

	var length = int64(0)

	if marker.HasLength() {
		length = jpeg.getSegmentLength(pos + 2)

		if length == -1 {
			return Field{}, fmt.Errorf("could not find length for marker %X: %s", marker.Byte(), marker.String())
		}
	} else {
		// fields with no legnth field are 2 bytes
		length = 2
	}

	return Field{pos, marker.String(), length}, nil
}

// GetFieldByMarker returns a Field struct for a given JPEG marker, either from cache or by extracting it
func (jpeg File) GetFieldByMarker(marker Marker) (Field, error) {
	field, ok := jpeg.fields[marker]

	if ok {
		return field, nil
	}

	var ret Field
	var err error
	if marker == SOF {
		// SOF is a special case since there are multiple possible markers
		ret, err = jpeg.getSOF()
	} else {
		// extract standard fields
		ret, err = jpeg.extractFieldByMarker(marker)
	}

	if err == nil {
		// cache the field data
		jpeg.fields[marker] = ret
	}

	return ret, err

}

// getEOIOffset returns the offset of the EOI marker
func (jpeg File) getEOI() (Field, error) {
	return jpeg.GetFieldByMarker(EOI)
}

// GetAppData returns a slice of Fields representing the APPn segments in the JPEG file
func (jpeg File) GetAppData() []Field {
	var ret []Field

	for i := 0; i < 16; i++ {
		field, err := jpeg.extractFieldByMarker(Marker(0xE0 + byte(i)))

		if err != nil {
			continue
		}

		ret = append(ret, field)
	}

	return ret
}

// GetHeight returns the height of the JPEG image in pixels, or -1 if it cannot be determined
func (jpeg File) GetHeight() int64 {
	field, err := jpeg.GetFieldByMarker(SOF)

	if err != nil {
		return int64(-1)
	}

	var buf = make([]byte, 2)

	// height is stored in bytes 5 and 6 of the SOF segment
	jpeg.Seek(field.Offset+5, 0)

	if _, err = jpeg.Read(buf); err != nil {
		return int64(-1)
	}

	return bytesToInt64(buf)
}

// GetWidth returns the width of the JPEG image in pixels, or -1 if it cannot be determined
func (jpeg File) GetWidth() int64 {
	field, err := jpeg.GetFieldByMarker(SOF)

	if err != nil {
		return int64(-1)
	}

	var buf = make([]byte, 2)

	// width is stored in bytes 7 and 8 of the SOF segment
	jpeg.Seek(field.Offset+7, 0)

	if _, err = jpeg.Read(buf); err != nil {
		return int64(-1)
	}

	return bytesToInt64(buf)
}

// GetSOS returns the Start of Scan (SOS) Field in the JPEG file
func (jpeg File) GetSOS() (Field, error) {
	return jpeg.GetFieldByMarker(SOS)
}

// HasSOF checks if the JPEG file has a Start of Frame (SOF) marker
func (jpeg File) HasSOF() bool {
	sof, err := jpeg.getSOF()

	return sof.Offset != -1 && sof.Length != -1 && err == nil
}

// GetDQT returns the Define Quantization Table (DQT) Field in the JPEG file
func (jpeg File) GetDQT() (Field, error) {
	return jpeg.GetFieldByMarker(DQT)
}

// getSOFOffset returns the offset of the Start of Frame (SOF) marker
// The SOF marker may be anything between C0 and CF, we will just look for the first
func (jpeg File) getSOF() (Field, error) {
	for i := 0; i < 16; i++ {
		pos, err := jpeg.findMarker(0xC0 + byte(i))

		if err == nil {
			length := jpeg.getSegmentLength(pos + 2)

			return Field{pos, SOF.String(), length}, nil
		}
	}

	return Field{}, fmt.Errorf("could not find SOF marker")
}

// findMarker searches for a specific JPEG marker
func (jpeg File) findMarker(marker byte) (int64, error) {
	var err error

	// Start scanning from the beginning of the file
	if _, err := jpeg.Seek(0, 0); err != nil {
		return -1, err
	}

	// read 1kb at a time for now
	buf := make([]byte, 1024)

	// Read until we find the marker or reach EOF
	foundMarkerStart := false
	for {
		if _, err = jpeg.Read(buf); err != nil {
			return -1, err
		}

		for i := 0; i < len(buf); i++ {
			if buf[i] == 0xFF {
				foundMarkerStart = true
				continue
			}

			if foundMarkerStart && buf[i] == marker {
				pos, err := jpeg.Seek(0, 1)

				if err != nil {
					return -1, err
				}

				// position of the marker is calculated by:
				// current read position (end of buffer) - 1 (for the 0xFF byte) - length of buffer - current index
				return pos - 1 - int64(len(buf)-i), nil
			}

			foundMarkerStart = false
		}
	}
}

// getSegmentLength reads the length of a JPEG segment from the file at the given offset
// this is blindly assumed to be the two bytes after the given offset
func (jpeg File) getSegmentLength(offset int64) int64 {
	var readBuf = make([]byte, 2)

	jpeg.Seek(offset, 0)

	_, err := jpeg.Read(readBuf)

	if err != nil {
		return -1
	}

	return bytesToInt64(readBuf)
}

// bytesToInt64 converts a 2-byte slice to an int64
func bytesToInt64(b []byte) int64 {
	if len(b) < 2 {
		return -1
	}
	return int64(b[0])<<8 | int64(b[1])
}

// ParseAppData reads the data from an APPn Field in the JPEG file and returns it as a map of strings
// for now, we only extract the identifier string
func ParseAppData(field Field, jpeg File) map[string]string {
	ret := make(map[string]string)
	buf := make([]byte, field.Length)

	jpeg.Seek(field.Offset, 0)
	jpeg.Read(buf)

	// parse
	length := bytesToInt64(buf[2:4])

	// extract the identifier
	for i := int64(4); i < length; i++ {
		if (buf[i]) == 0x00 {
			ret["identifier"] = string(buf[4:i])
			break
		}
	}

	return ret
}

// GetCompressedImageData returns the compressed image data between the SOS and EOI markers
func GetCompressedImageData(jpeg File) []byte {
	sos, err := jpeg.GetSOS()

	if err != nil {
		return nil
	}

	// we identify the start of the compressed image data by finding the SOS marker (0xFF 0xDA)
	// skipping and skipping the length of that field
	jpeg.Seek(sos.Offset+sos.Length, 0)

	// we will read until we hit the EOI marker (0xFF 0xD9)
	// note that this does not support progressive jpegs
	eoi, err := jpeg.getEOI()

	if err != nil {
		return nil
	}

	var buf = make([]byte, eoi.Offset-(sos.Offset-1000))

	jpeg.Seek(sos.Offset+sos.Length, 0)
	if _, err = jpeg.Read(buf); err != nil {
		return nil
	}

	return buf
}
