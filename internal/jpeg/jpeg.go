package jpeg

import (
	"fmt"
	"os"
)

// File represents a JPEG file on disk.
type File struct {
	*os.File
	fields map[Marker]Field
}

// Field represents a segment or marker field within the JPEG file.
type Field struct {
	Offset int64
	Name   string
	Length int64
}

// Open opens a JPEG file and returns a File.
func Open(filename string) (*File, error) {
	fh, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return &File{File: fh, fields: make(map[Marker]Field)}, nil
}

// HasSOI reports whether the file contains a Start of Image marker.
func (f *File) HasSOI() bool {
	_, err := f.Field(SOI)
	return err == nil
}

// HasEOI reports whether the file contains an End of Image marker.
func (f *File) HasEOI() bool {
	_, err := f.Field(EOI)
	return err == nil
}

// field extracts a Field for a given marker.
func (f *File) field(marker Marker) (Field, error) {
	pos, err := f.findMarker(marker.Byte())
	if err != nil {
		return Field{}, err
	}
	if pos == -1 {
		return Field{}, fmt.Errorf("marker not found %X: %s", marker.Byte(), marker.String())
	}

	var length int64
	if marker.HasLength() {
		length = f.segmentLength(pos + 2)
		if length == -1 {
			return Field{}, fmt.Errorf("no length for marker %X: %s", marker.Byte(), marker.String())
		}
	} else {
		// fields with no length field are 2 bytes
		length = 2
	}

	return Field{pos, marker.String(), length}, nil
}

// Field returns a Field for the given marker, from cache if available.
func (f *File) Field(marker Marker) (Field, error) {
	if field, ok := f.fields[marker]; ok {
		return field, nil
	}

	var (
		ret Field
		err error
	)
	if marker == SOF {
		ret, err = f.sof()
	} else {
		ret, err = f.field(marker)
	}

	if err == nil {
		f.fields[marker] = ret
	}
	return ret, err
}

// eoi returns the End of Image marker field.
func (f *File) eoi() (Field, error) {
	return f.Field(EOI)
}

// AppData returns the APPn segments in the file.
func (f *File) AppData() []Field {
	var ret []Field
	for i := 0; i < 16; i++ {
		field, err := f.field(Marker(0xE0 + byte(i)))
		if err != nil {
			continue
		}
		ret = append(ret, field)
	}
	return ret
}

// Height reports the image height in pixels.
func (f *File) Height() (int64, error) {
	field, err := f.Field(SOF)
	if err != nil {
		return -1, err
	}

	buf := make([]byte, 2)
	if _, err = f.Seek(field.Offset+5, 0); err != nil {
		return -1, err
	}
	if _, err = f.Read(buf); err != nil {
		return -1, err
	}
	return bytesToInt64(buf), nil
}

// Width reports the image width in pixels.
func (f *File) Width() (int64, error) {
	field, err := f.Field(SOF)
	if err != nil {
		return -1, err
	}

	buf := make([]byte, 2)
	if _, err = f.Seek(field.Offset+7, 0); err != nil {
		return -1, err
	}
	if _, err = f.Read(buf); err != nil {
		return -1, err
	}
	return bytesToInt64(buf), nil
}

// SOS returns the Start of Scan field.
func (f *File) SOS() (Field, error) {
	return f.Field(SOS)
}

// HasSOF reports whether the file has a Start of Frame marker.
func (f *File) HasSOF() bool {
	sof, err := f.sof()
	return sof.Offset != -1 && sof.Length != -1 && err == nil
}

// DQT returns the Define Quantization Table field.
func (f *File) DQT() (Field, error) {
	return f.Field(DQT)
}

// sof returns the first Start of Frame field (C0–CF).
func (f *File) sof() (Field, error) {
	for i := 0; i < 16; i++ {
		pos, err := f.findMarker(0xC0 + byte(i))
		if err == nil {
			length := f.segmentLength(pos + 2)
			return Field{pos, SOF.String(), length}, nil
		}
	}
	return Field{}, fmt.Errorf("SOF marker not found")
}

// findMarker searches for the given marker byte.
func (f *File) findMarker(marker byte) (int64, error) {
	if _, err := f.Seek(0, 0); err != nil {
		return -1, err
	}

	buf := make([]byte, 1024)
	foundFF := false

	for {
		n, err := f.Read(buf)
		if err != nil {
			return -1, err
		}
		for i := 0; i < n; i++ {
			if buf[i] == 0xFF {
				foundFF = true
				continue
			}
			if foundFF && buf[i] == marker {
				pos, err := f.Seek(0, 1)
				if err != nil {
					return -1, err
				}
				return pos - 1 - int64(n-i), nil
			}
			foundFF = false
		}
	}
}

// segmentLength reads a segment length from the given offset.
func (f *File) segmentLength(offset int64) int64 {
	buf := make([]byte, 2)
	if _, err := f.Seek(offset, 0); err != nil {
		return -1
	}
	if _, err := f.Read(buf); err != nil {
		return -1
	}
	return bytesToInt64(buf)
}

// bytesToInt64 converts two bytes to int64.
func bytesToInt64(b []byte) int64 {
	if len(b) < 2 {
		return -1
	}
	return int64(b[0])<<8 | int64(b[1])
}

// ParseAppData extracts identifier strings from an APPn field.
func ParseAppData(field Field, f *File) map[string]string {
	ret := make(map[string]string)
	buf := make([]byte, field.Length)

	_, _ = f.Seek(field.Offset, 0)
	_, _ = f.Read(buf)

	length := bytesToInt64(buf[2:4])
	for i := int64(4); i < length; i++ {
		if buf[i] == 0x00 {
			ret["identifier"] = string(buf[4:i])
			break
		}
	}
	return ret
}

// CompressedData returns the compressed image data between SOS and EOI.
func CompressedData(f *File) []byte {
	sos, err := f.SOS()
	if err != nil {
		return nil
	}

	_, _ = f.Seek(sos.Offset+sos.Length, 0)
	eoi, err := f.eoi()
	if err != nil {
		return nil
	}

	buf := make([]byte, eoi.Offset-(sos.Offset-1000))
	_, _ = f.Seek(sos.Offset+sos.Length, 0)
	if _, err = f.Read(buf); err != nil {
		return nil
	}
	return buf
}