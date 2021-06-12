/*
Package field is useful for reading/writing byte fields e.g. from network
protocol formatted data.
*/
package field

import (
	"encoding/binary"
	"unsafe"
)

// BigEndian reads/writes integer value as Big Endian.
var BigEndian = bigEndian{}

// LittleEndian reads/writes integer value as Little Endian.
var LittleEndian = littleEndian{}

type bigEndian struct{}

type littleEndian struct{}

// Endianness allows to read/write integer fields with different
// endianness.
type Endianness interface {
	// Append integer number to specified slice and return resulting
	// slice.
	WriteUint16([]byte, uint16) []byte
	WriteUint24([]byte, uint32) []byte
	WriteUint32([]byte, uint32) []byte
	WriteUint64([]byte, uint64) []byte

	// Read integer value of specified length from the top of the
	// slice. Return remaining data and true if reading was ok. If
	// false return nil or a slice of zero-length.
	ReadUint16([]byte, *uint16) ([]byte, bool)
	ReadUint24([]byte, *uint32) ([]byte, bool)
	ReadUint32([]byte, *uint32) ([]byte, bool)
	ReadUint64([]byte, *uint64) ([]byte, bool)
}

var _ Endianness = BigEndian
var _ Endianness = LittleEndian

// WriteUint8 writes a Uint8 value to specified slice and returns
// the resulted slice.
func WriteUint8(data []byte, x uint8) []byte {
	return append(data, x)
}

// ReadUint8 reads a Uint8 value from specified slice and returns
// the resulting slice and boolean flag telling if it was a success.
//
// Value is treated as big endian.
func ReadUint8(data []byte, x *uint8) ([]byte, bool) {
	if len(data) >= 1 {
		*x = data[0]
		return data[1:], true
	}

	return nil, false
}

// WriteUint16 writes a Uint16 value to specified slice and returns
// the resulting slice.
//
// Value is treated as big endian.
func (bigEndian) WriteUint16(data []byte, x uint16) []byte {
	d := (*[unsafe.Sizeof(x)]byte)(unsafe.Pointer(&x))[:]
	binary.BigEndian.PutUint16(d, x)
	return append(data, d...)
}

// WriteUint32 writes a Uint32 value to specified slice and returns
// the resulting slice.
//
// Value is treated as big endian.
func (bigEndian) WriteUint32(data []byte, x uint32) []byte {
	d := (*[unsafe.Sizeof(x)]byte)(unsafe.Pointer(&x))[:]
	binary.BigEndian.PutUint32(d, x)
	return append(data, d...)
}

func (bigEndian) WriteUint24(data []byte, x uint32) []byte {
	d := (*[unsafe.Sizeof(x)]byte)(unsafe.Pointer(&x))[:]
	binary.BigEndian.PutUint32(d, x)
	return append(data, d[1:]...)
}

func (littleEndian) WriteUint24(data []byte, x uint32) []byte {
	d := (*[unsafe.Sizeof(x)]byte)(unsafe.Pointer(&x))[:]
	binary.LittleEndian.PutUint32(d, x)
	return append(data, d[:3]...)
}

// WriteUint64 writes a Uint64 value to specified slice and returns
// the resulting slice.
//
// Value is treated as big endian.
func (bigEndian) WriteUint64(data []byte, x uint64) []byte {
	d := (*[unsafe.Sizeof(x)]byte)(unsafe.Pointer(&x))[:]
	binary.BigEndian.PutUint64(d, x)
	return append(data, d...)
}

// WriteInt64 writes a Int64 value to specified slice and returns
// the resulting slice.
//
// Value is treated as big endian.
func (bigEndian) WriteInt64(data []byte, x int64) []byte {
	d := (*[unsafe.Sizeof(x)]byte)(unsafe.Pointer(&x))[:]
	binary.BigEndian.PutUint64(d, uint64(x))
	return append(data, d...)
}

// ReadUint16 reads a Uint16 value from specified slice and returns
// the resulting slice and boolean flag telling if it was a success.
//
// Value is treated as big endian.
func (bigEndian) ReadUint16(data []byte, x *uint16) ([]byte, bool) {
	if d := int(unsafe.Sizeof(*x)); len(data) >= d {
		*x = binary.BigEndian.Uint16(data)
		return data[d:], true
	}

	return nil, false
}

// ReadUint24 reads a 24-bit value from specified slice and returns
// the resulting slice and boolean flag telling if it was a success.
//
// Value is treated as big endian.
func (bigEndian) ReadUint24(data []byte, x *uint32) ([]byte, bool) {
	if d := 3; len(data) >= d {
		*x = uint32(data[0]) << 16
		*x += uint32(data[1]) << 8
		*x += uint32(data[2])
		return data[3:], true
	}

	return nil, false
}

// ReadUint32 reads a Uint32 value from specified slice and returns
// the resulting slice and boolean flag telling if it was a success.
//
// Value is treated as big endian.
func (bigEndian) ReadUint32(data []byte, x *uint32) ([]byte, bool) {
	if d := int(unsafe.Sizeof(*x)); len(data) >= d {
		*x = binary.BigEndian.Uint32(data)
		return data[d:], true
	}

	return nil, false
}

// ReadUint64 reads a Uint64 value from specified slice and returns
// the resulting slice and boolean flag telling if it was a success.
//
// Value is treated as big endian.
func (bigEndian) ReadUint64(data []byte, x *uint64) ([]byte, bool) {
	if d := int(unsafe.Sizeof(*x)); len(data) >= d {
		*x = binary.BigEndian.Uint64(data)
		return data[d:], true
	}

	return nil, false
}

// WriteUint16 writes a Uint16 value to specified slice and returns
// the resulting slice.
//
// Value is treated as little endian.
func (littleEndian) WriteUint16(data []byte, x uint16) []byte {
	d := (*[unsafe.Sizeof(x)]byte)(unsafe.Pointer(&x))[:]
	binary.LittleEndian.PutUint16(d, x)
	return append(data, d...)
}

// WriteUint32 writes a Uint32 value to specified slice and returns
// the resulting slice.
//
// Value is treated as little endian.
func (littleEndian) WriteUint32(data []byte, x uint32) []byte {
	d := (*[unsafe.Sizeof(x)]byte)(unsafe.Pointer(&x))[:]
	binary.LittleEndian.PutUint32(d, x)
	return append(data, d...)
}

// WriteUint64 writes a Uint64 value to specified slice and returns
// the resulting slice.
//
// Value is treated as little endian.
func (littleEndian) WriteUint64(data []byte, x uint64) []byte {
	d := (*[unsafe.Sizeof(x)]byte)(unsafe.Pointer(&x))[:]
	binary.LittleEndian.PutUint64(d, x)
	return append(data, d...)
}

// WriteInt64 writes a Int64 value to specified slice and returns
// the resulting slice.
//
// Value is treated as little endian.
func (littleEndian) WriteInt64(data []byte, x int64) []byte {
	d := (*[unsafe.Sizeof(x)]byte)(unsafe.Pointer(&x))[:]
	binary.LittleEndian.PutUint64(d, uint64(x))
	return append(data, d...)
}

// ReadUint16 reads a Uint16 value from specified slice and returns
// the resulting slice and boolean flag telling if it was a success.
//
// Value is treated as little endian.
func (littleEndian) ReadUint16(data []byte, x *uint16) ([]byte, bool) {
	if d := int(unsafe.Sizeof(*x)); len(data) >= d {
		*x = binary.LittleEndian.Uint16(data)
		return data[d:], true
	}

	return nil, false
}

// ReadUint24 reads a 24-bit value from specified slice and returns
// the resulting slice and boolean flag telling if it was a success.
//
// Value is treated as little endian.
func (littleEndian) ReadUint24(data []byte, x *uint32) ([]byte, bool) {
	if d := 3; len(data) >= d {
		*x = uint32(data[0])
		*x += uint32(data[1]) << 8
		*x += uint32(data[2]) << 16
		return data[3:], true
	}

	return nil, false
}

// ReadUint32 reads a Uint32 value from specified slice and returns
// the resulting slice and boolean flag telling if it was a success.
//
// Value is treated as little endian.
func (littleEndian) ReadUint32(data []byte, x *uint32) ([]byte, bool) {
	if d := int(unsafe.Sizeof(*x)); len(data) >= d {
		*x = binary.LittleEndian.Uint32(data)
		return data[d:], true
	}

	return nil, false
}

// ReadUint64 reads a Uint64 value from specified slice and returns
// the resulting slice and boolean flag telling if it was a success.
//
// Value is treated as little endian.
func (littleEndian) ReadUint64(data []byte, x *uint64) ([]byte, bool) {
	if d := int(unsafe.Sizeof(*x)); len(data) >= d {
		*x = binary.LittleEndian.Uint64(data)
		return data[d:], true
	}

	return nil, false
}

// ReadBytes checks if data has n bytes at the top and puts them into
// x. Returns the resulting slice and true if reslicing was
// successful.
func ReadBytes(data []byte, x *[]byte, n int) ([]byte, bool) {
	if len(data) >= n {
		*x = data[:n]
		return data[n:], true
	}

	return nil, false
}

// CopyBytes copies up to len(x) bytes from the top of data to x.
// Returns the resulting slice and true if len(x) bytes were copied.
func CopyBytes(data []byte, x []byte) ([]byte, bool) {
	n := copy(x, data)
	return data[n:], len(x) == n
}

// SkipBytes skips n bytes in data, returning resulting slice and true
// if data had at least n bytes.
func SkipBytes(data []byte, n int) ([]byte, bool) {
	if n <= len(data) {
		return data[n:], true
	}
	return nil, false
}
