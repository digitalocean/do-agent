package structuredstream

import (
	"encoding/binary"
	"io"
	"time"
)

// Writer wraps a writer with calls for writing binary data types
// If any error is encountered, all subsequent calls will fail
// error checking must be done in a separate call to error
type Writer struct {
	w         io.Writer
	err       error
	byteOrder binary.ByteOrder
}

// NewWriter returns a new writer
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w:         w,
		err:       nil,
		byteOrder: binary.LittleEndian,
	}
}

// Write writes a given type to the stream
func (s *Writer) Write(x interface{}) {
	if s.err == nil {
		s.err = binary.Write(s.w, s.byteOrder, x)
	}
}

// WriteUint16PrefixedBytes writes a uint16 specifying the length of the bytes buffer, followed by the payload
func (s *Writer) WriteUint16PrefixedBytes(x []byte) {
	l := len(x)
	if l > 0xFFFF {
		panic("overflow")
	}
	s.Write(uint16(l))
	s.Write(x)
}

// WriteUint16PrefixedString writes a uint16 specifying the length of the string, followed by the actual string
func (s *Writer) WriteUint16PrefixedString(x string) {
	s.WriteUint16PrefixedBytes([]byte(x))
}

// WriteUnixTime64UTC writes a time as unix epoch in UTC; sub-second accuracy is truncated
func (s *Writer) WriteUnixTime64UTC(x time.Time) {
	s.Write(int64(x.UTC().Unix()))
}

// Error returns any errors the occurred since the writer was first constructed
func (s *Writer) Error() error {
	return s.err
}
