package structuredstream

import (
	"encoding/binary"
	"io"
	"time"
)

// Reader wraps a reader with calls for reading binary data types
// If any error is encountered, all subsequent calls will fail
// error checking must be done in a separate call to error
type Reader struct {
	r         io.Reader
	err       error
	byteOrder binary.ByteOrder
}

// NewReader returns a new reader
func NewReader(r io.Reader) *Reader {
	return &Reader{
		r:         r,
		err:       nil,
		byteOrder: binary.LittleEndian,
	}
}

// Read takes a pointer to a data type (e.g. uint16, int64, []byte) and reads
// data from the wrapped reader, and advances the reader offset to the next value.
func (s *Reader) Read(x interface{}) {
	if s.err == nil {
		s.err = binary.Read(s.r, s.byteOrder, x)
	}
}

// ReadInt8 returns a int8
func (s *Reader) ReadInt8() int8 {
	var x int8
	s.Read(&x)
	return x
}

// ReadUint8 returns a uint8
func (s *Reader) ReadUint8() uint8 {
	var x uint8
	s.Read(&x)
	return x
}

// ReadInt16 returns an int16
func (s *Reader) ReadInt16() int16 {
	var x int16
	s.Read(&x)
	return x
}

// ReadUint16 returns a uint16
func (s *Reader) ReadUint16() uint16 {
	var x uint16
	s.Read(&x)
	return x
}

// ReadInt32 returns an int32
func (s *Reader) ReadInt32() int32 {
	var x int32
	s.Read(&x)
	return x
}

// ReadUint32 retursn a uint32
func (s *Reader) ReadUint32() uint32 {
	var x uint32
	s.Read(&x)
	return x
}

// ReadInt64 retursn a int64
func (s *Reader) ReadInt64() int64 {
	var x int64
	s.Read(&x)
	return x
}

// ReadUint64 retursn a uint64
func (s *Reader) ReadUint64() uint64 {
	var x uint64
	s.Read(&x)
	return x
}

// ReadFloat64 retursn a float64
func (s *Reader) ReadFloat64() float64 {
	var x float64
	s.Read(&x)
	return x
}

// ReadBytes returns l many bytes
func (s *Reader) ReadBytes(l int) []byte {
	buf := make([]byte, l)

	var n int
	n, s.err = io.ReadFull(s.r, buf)
	if n != l {
		panic("underflow")
	}
	return buf
}

// ReadUint16PrefixedBytes first reads a uint16, then reads that many following bytes
func (s *Reader) ReadUint16PrefixedBytes() []byte {
	l := s.ReadUint16()
	if s.err == nil {
		x := s.ReadBytes(int(l))
		return x
	}
	return nil
}

// ReadUint16PrefixedString first reads a uint16, then reads that many following chars
func (s *Reader) ReadUint16PrefixedString() string {
	return string(s.ReadUint16PrefixedBytes())
}

// ReadUnixTime64UTC reads a uint64 representing unix epoch time in UTC and converts it to a time.time
func (s *Reader) ReadUnixTime64UTC() time.Time {
	var x int64
	s.Read(&x)

	// can't use time.Unix which assumes timezone is local
	t := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Second * time.Duration(x))
	return t
}

// Error returns the last encountered error
func (s *Reader) Error() error {
	return s.err
}
