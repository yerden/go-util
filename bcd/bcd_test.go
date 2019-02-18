package bcd

import (
	"bytes"
	"fmt"
	"testing"
)

func newAssert(t *testing.T, fail bool) func(bool) {
	return func(expected bool) {
		if t.Helper(); !expected {
			t.Error("Something's not right")
			if fail {
				t.FailNow()
			}
		}
	}
}

var (
	// TBCD as in 3GPP TS 29.002
	enc = &BCD{
		Map: map[byte]byte{
			'0': 0x0, '1': 0x1, '2': 0x2, '3': 0x3,
			'4': 0x4, '5': 0x5, '6': 0x6, '7': 0x7,
			'8': 0x8, '9': 0x9, '*': 0xa, '#': 0xb,
			'a': 0xc, 'b': 0xd, 'c': 0xe,
		},
		SwapNibble: true,
		Filler:     0xf}
)

func TestPlainEncode(t *testing.T) {
	assert := newAssert(t, false)

	input := "12345"
	output, err := enc.PlainEncode([]byte(input), []byte{})
	assert(err == nil)
	assert(bytes.Equal(output, []byte{0x21, 0x43, 0xf5}))

	input = "abc"
	output, err = enc.PlainEncode([]byte(input), []byte{})
	assert(err == nil)
	assert(bytes.Equal(output, []byte{0xdc, 0xfe}))

	input = "unacceptable"
	output, err = enc.PlainEncode([]byte(input), []byte{})
	assert(err == ErrBadInput)
}

func TestEncodeOpt(t *testing.T) {
	assert := newAssert(t, false)
	enc := NewEncoder(enc)
	output := make([]byte, 30)

	input := "12345"
	n, err := enc.Encode(output, []byte(input))
	assert(err == nil)
	assert(bytes.Equal(output[:n], []byte{0x21, 0x43, 0xf5}))

	input = "abc"
	n, err = enc.Encode(output, []byte(input))
	assert(err == nil)
	assert(bytes.Equal(output[:n], []byte{0xdc, 0xfe}))

	input = "unacceptable"
	n, err = enc.Encode(output, []byte(input))
	assert(err == ErrBadInput)
}

func ExampleEncoder_Encode() {
	enc := NewEncoder(TBCDEncoding)

	src := []byte("12345")
	dst := make([]byte, EncodedLen(len(src)))
	n, err := enc.Encode(dst, src)
	if err != nil {
		return
	}

	fmt.Println(bytes.Equal(dst[:n], []byte{0x21, 0x43, 0xf5}))
	// Output: true
}

func ExampleEncoder_Encode_second() {
	enc := NewEncoder(StdEncoding)

	src := []byte("1234")
	dst := make([]byte, EncodedLen(len(src)))
	n, err := enc.Encode(dst, src)
	if err != nil {
		return
	}

	fmt.Println(bytes.Equal(dst[:n], []byte{0x12, 0x34}))
	// Output: true
}

func ExampleDecoder_Decode() {
	dec := NewDecoder(TBCDEncoding)

	src := []byte{0x21, 0x43, 0xf5}
	dst := make([]byte, DecodedLen(len(src)))
	n, err := dec.Decode(dst, src)
	if err != nil {
		return
	}

	fmt.Println(string(dst[:n]))
	// Output: 12345
}

func TestPlainDecode(t *testing.T) {
	assert := newAssert(t, true)

	input := []byte{0x21, 0x43, 0xf5}
	output, err := enc.PlainDecode(input, []byte{})
	assert(err == nil)
	assert(string(output) == "12345")

	input = []byte{0x21, 0x43}
	output, err = enc.PlainDecode(input, []byte{})
	assert(err == nil)
	assert(string(output) == "1234")

	input = []byte{0x21, 0xf3}
	output, err = enc.PlainDecode(input, []byte{})
	assert(err == nil)
	assert(string(output) == "123")

	input = []byte{0xdc, 0xfe}
	output, err = enc.PlainDecode(input, []byte{})
	assert(err == nil)
	assert(string(output) == "abc")

	input = []byte{0xff, 0xff}
	output, err = enc.PlainDecode(input, []byte{})
	assert(err == ErrBadBCD)
}

func TestDecodeOpt(t *testing.T) {
	assert := newAssert(t, true)
	enc := NewDecoder(enc)
	output := make([]byte, 20)

	input := []byte{0x21, 0x43, 0xf5}
	n, err := enc.Decode(output, input)
	assert(err == nil)
	assert(string(output[:n]) == "12345")

	input = []byte{0x21, 0x43}
	n, err = enc.Decode(output, input)
	assert(err == nil)
	assert(string(output[:n]) == "1234")

	input = []byte{0x21, 0xf3}
	n, err = enc.Decode(output, input)
	assert(err == nil)
	assert(string(output[:n]) == "123")

	input = []byte{0xdc, 0xfe}
	n, err = enc.Decode(output, input)
	assert(err == nil)
	assert(string(output[:n]) == "abc")

	input = []byte{0xff, 0xff}
	n, err = enc.Decode(output, input)
	assert(err == ErrBadBCD)
}

func BenchmarkPlainEncode(b *testing.B) {
	out := make([]byte, 0, 24)

	for i := 0; i < b.N; i++ {
		enc.PlainEncode([]byte("123456789"), out)
	}
}

func BenchmarkEncodeOpt(b *testing.B) {
	out := make([]byte, 0, 24)
	enc := NewEncoder(enc)

	for i := 0; i < b.N; i++ {
		enc.Encode(out, []byte("123456789"))
	}
}

func BenchmarkPlainDecode(b *testing.B) {
	out := make([]byte, 0, 24)

	for i := 0; i < b.N; i++ {
		enc.PlainDecode([]byte{0x21, 0x43, 0x65, 0x87}, out)
	}
}

func BenchmarkDecodeOpt(b *testing.B) {
	out := make([]byte, 0, 24)
	enc := NewDecoder(enc)

	for i := 0; i < b.N; i++ {
		enc.Decode(out, []byte{0x21, 0x43, 0x65, 0x87})
	}
}
