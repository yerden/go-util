package bcd

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"testing/iotest"
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
	assert(n == 3)
	assert(bytes.Equal(output[:n], []byte{0x21, 0x43, 0xf5}))

	input = "abc"
	n, err = enc.Encode(output, []byte(input))
	assert(err == nil)
	assert(n == 2)
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

	// try not ignoring the fillers
	input = []byte{0x21, 0xf3, 0xf4, 0x65, 0xf7}
	enc.IgnoreFiller = false
	n, err = enc.Decode(output, input)
	assert(err == ErrBadBCD)

	// try ignoring the fillers
	enc.IgnoreFiller = true
	n, err = enc.Decode(output, input)
	assert(string(output[:n]) == "1234567")
}

func testDecodeReader(t *testing.T, srcS []byte, dstS string, expected bool) {
	assert := newAssert(t, false)
	enc := NewDecoder(enc)

	var dst bytes.Buffer
	src := iotest.OneByteReader(enc.NewReader(bytes.NewBuffer(srcS)))
	_, err := io.Copy(&dst, src)

	if expected {
		assert(dst.String() == dstS)
		assert(err == nil)
	} else {
		assert(err == ErrBadBCD)
	}

	dst.Reset()
	src = enc.NewReader(bytes.NewBuffer(srcS))
	_, err = io.Copy(&dst, src)

	if expected {
		assert(dst.String() == dstS)
		assert(err == nil)
	} else {
		assert(err == ErrBadBCD)
	}
}

func testEncodeReader(t *testing.T, srcS []byte, dstS string, expected bool) {
	assert := newAssert(t, false)
	src := iotest.OneByteReader(bytes.NewBufferString(dstS))
	dst := new(bytes.Buffer)
	w := NewEncoder(enc).NewWriter(dst)

	_, err := io.Copy(w, src)
	if expected {
		assert(err == nil)
	} else {
		assert(err != nil)
		return
	}
	if len(dstS)%2 == 0 {
		assert(w.Buffered() == 0)
	} else {
		assert(w.Buffered() == 1)
	}
	err = w.Flush()
	assert(err == nil)
	assert(w.Buffered() == 0)
	if expected {
		assert(bytes.Equal(dst.Bytes(), srcS))
	} else {
		assert(!bytes.Equal(dst.Bytes(), srcS))
	}
}

func TestDecodeReader(t *testing.T) {
	testDecodeReader(t, []byte{0x21, 0x43, 0xf5}, "12345", true)
	testDecodeReader(t, []byte{0x21, 0x43}, "1234", true)
	testDecodeReader(t, []byte{0x21, 0xf3}, "123", true)
	testDecodeReader(t, []byte{0xdc, 0xfe}, "abc", true)
	testDecodeReader(t, []byte{0xff, 0xff}, "abc", false)
	testDecodeReader(t, []byte{0xfe, 0xff}, "abc", false)
}

func TestEncodeReader(t *testing.T) {
	testEncodeReader(t, []byte{0x21, 0x43, 0xf5}, "12345", true)
	testEncodeReader(t, []byte{0x21, 0x43}, "1234", true)
	testEncodeReader(t, []byte{0x21, 0xf3}, "123", true)
	testEncodeReader(t, []byte{0xdc, 0xfe}, "abc", true)
	testEncodeReader(t, []byte{0xfe, 0xff}, "hrhdsg", false)
	testDecodeReader(t, []byte{0xfe, 0xff}, "abc", false)
}

func TestReaderWriter(t *testing.T) {
	assert := newAssert(t, false)
	s := "1234521293476283476196" +
		"9384729845768676325420" +
		"9077754234857329847513" +
		"3028573488274556781023" +
		"0876416235676495867679" +
		"0"

	src := bytes.NewBufferString(s)
	dst := new(bytes.Buffer)
	enc := NewEncoder(AikenEncoding).NewWriter(dst)
	dec := NewDecoder(AikenEncoding).NewReader(dst)

	_, err := io.Copy(enc, src)
	assert(enc.Buffered() == 1)
	assert(enc.Flush() == nil)
	assert(enc.Buffered() == 0)
	assert(err == nil)

	_, err = io.Copy(src, dec)
	assert(err == nil)
	assert(src.String() == s)
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
