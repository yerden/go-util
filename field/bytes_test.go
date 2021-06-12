package field

import (
	"bytes"
	"math/rand"
	"testing"
	"time"
)

func assert(t testing.TB, expected bool, args ...interface{}) {
	if !expected {
		t.Helper()
		t.Fatal(args...)
	}
}

func TestBeWrite(t *testing.T) {
	data := []byte{}

	data = WriteUint8(data, 10)
	assert(t, bytes.Equal(data, []byte{10}))

	data = BigEndian.WriteUint16(data, 0xaabb)
	assert(t, bytes.Equal(data, []byte{10, 0xaa, 0xbb}))

	data = BigEndian.WriteUint32(data, 0xaabbccdd)
	assert(t, bytes.Equal(data, []byte{10, 0xaa, 0xbb, 0xaa, 0xbb, 0xcc, 0xdd}))

	data = BigEndian.WriteUint64(data[:0], 0x11223344aabbccdd)
	assert(t, bytes.Equal(data, []byte{0x11, 0x22, 0x33, 0x44, 0xaa, 0xbb, 0xcc, 0xdd}))

	data = BigEndian.WriteUint24(data[:0], 0x112233)
	assert(t, bytes.Equal(data, []byte{0x11, 0x22, 0x33}))

	data = LittleEndian.WriteUint24(data[:0], 0x112233)
	assert(t, bytes.Equal(data, []byte{0x33, 0x22, 0x11}))
}

func TestBeRead(t *testing.T) {
	data := []byte{0xaa, 0xbb, 0xcc, 0xdd, 0x11, 0x22, 0x33, 0x44}

	{
		x := uint8(0)
		res, ok := ReadUint8(data, &x)
		assert(t, ok)
		assert(t, bytes.Equal(res, []byte{0xbb, 0xcc, 0xdd, 0x11, 0x22, 0x33, 0x44}))
		assert(t, x == 0xaa)
	}

	{
		x := uint16(0)
		res, ok := BigEndian.ReadUint16(data, &x)
		assert(t, ok)
		assert(t, bytes.Equal(res, []byte{0xcc, 0xdd, 0x11, 0x22, 0x33, 0x44}))
		assert(t, x == 0xaabb)
	}

	{
		x := uint32(0)
		res, ok := BigEndian.ReadUint24(data, &x)
		assert(t, ok)
		assert(t, bytes.Equal(res, []byte{0xdd, 0x11, 0x22, 0x33, 0x44}))
		assert(t, x == 0xaabbcc)

		res, ok = LittleEndian.ReadUint24(res, &x)
		assert(t, ok)
		assert(t, bytes.Equal(res, []byte{0x33, 0x44}))
		assert(t, x == 0x2211dd)
	}

	{
		x := uint32(0)
		res, ok := BigEndian.ReadUint32(data, &x)
		assert(t, ok)
		assert(t, bytes.Equal(res, []byte{0x11, 0x22, 0x33, 0x44}))
		assert(t, x == 0xaabbccdd)

		y := uint64(0)
		res, ok = BigEndian.ReadUint64(res, &y)
		assert(t, !ok)
	}

	{
		x := uint64(0)
		res, ok := BigEndian.ReadUint64(data, &x)
		assert(t, ok)
		assert(t, bytes.Equal(res, []byte{}))
		assert(t, x == 0xaabbccdd11223344)
	}
}

func TestCopyBytes(t *testing.T) {
	data := make([]byte, 256)
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	rnd.Read(data)

	for n := 0; n <= 256; n++ {
		x := make([]byte, n)
		buf, ok := CopyBytes(data, x)
		assert(t, ok)
		assert(t, bytes.Equal(buf, data[n:]))
		assert(t, bytes.Equal(x, data[:n]))
	}

	for n := 257; n < 1000; n++ {
		x := make([]byte, n)
		buf, ok := CopyBytes(data, x)
		assert(t, !ok)
		assert(t, bytes.Equal(buf, []byte{}))
		assert(t, bytes.Equal(x[:256], data))
	}
}

func BenchmarkCopyBytes16(b *testing.B) {
	data := make([]byte, 256)
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	rnd.Read(data)

	for n := 0; n <= b.N; n++ {
		x := make([]byte, 16)
		CopyBytes(data, x)
	}
}

func BenchmarkReadUint24(b *testing.B) {
	data := make([]byte, 3)
	var x uint32

	for n := 0; n < b.N; n++ {
		BigEndian.ReadUint24(data, &x)
	}
}

func BenchmarkReadUint24_4Bytes(b *testing.B) {
	data := make([]byte, 4)
	var x uint32

	for n := 0; n < b.N; n++ {
		BigEndian.ReadUint24(data, &x)
	}
}

func BenchmarkWriteUint24(b *testing.B) {
	data := make([]byte, 3)
	var x uint32

	for n := 0; n < b.N; n++ {
		data = BigEndian.WriteUint24(data[:0], x)
	}
}

func BenchmarkReadUint32(b *testing.B) {
	data := make([]byte, 4)
	var x uint32

	for n := 0; n < b.N; n++ {
		BigEndian.ReadUint32(data, &x)
	}
}

func BenchmarkWriteUint32(b *testing.B) {
	data := make([]byte, 4)
	var x uint32

	for n := 0; n < b.N; n++ {
		data = BigEndian.WriteUint32(data[:0], x)
	}
}
