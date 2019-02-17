package bcd

import (
	"bytes"
	"io"
)

type word [2]byte
type dword [4]byte
type qword [8]byte

// Decoder is used to decode BCD converted bytes into
// decimal string.
type Decoder struct {
	// two nibbles (1 byte) to 2 symbols mapping; example:
	// 0x45 -> '45' or '54' depending on nibble swapping
	// additional 2 bytes of dword should be 0, otherwise
	// given byte is unacceptable
	hashWord [0x100]dword

	// one finishing byte with filler nibble
	// to 1 symbol mapping; example:
	// 0x4f -> '4' (filler=0xf, swap=false)
	// additional byte of word should 0, otherise
	// given nibble is unacceptable
	hashByte [0x100]word
}

func newHashDecWord(config *BCD) (res [0x100]dword) {
	var w dword
	var b byte
	for i, _ := range res {
		// invalidating all bytes by default
		res[i] = dword{0xff, 0xff, 0xff, 0xff}
	}

	for c1, nib1 := range config.Map {
		for c2, nib2 := range config.Map {
			b = (nib1 << 4) + nib2&0xf
			if config.SwapNibble {
				w = dword{c2, c1, 0, 0}
			} else {
				w = dword{c1, c2, 0, 0}
			}
			res[b] = w
		}
	}
	return
}

func newHashDecByte(config *BCD) (res [0x100]word) {
	var b byte
	for i, _ := range res {
		// invalidating all nibbles by default
		res[i] = word{0xff, 0xff}
	}
	for c, nib := range config.Map {
		if config.SwapNibble {
			b = (config.Filler << 4) + nib&0xf
		} else {
			b = (nib << 4) + config.Filler&0xf
		}
		res[b] = word{c, 0}
	}
	return
}

func (dec *Decoder) unpackWord(b byte) (word, error) {
	dw := dec.hashWord[b]
	if dw[2] != 0 {
		return word{}, ErrBadBCD
	}

	return word{dw[0], dw[1]}, nil
}

func (dec *Decoder) unpackLastByte(b byte) (byte, error) {
	w := dec.hashByte[b]
	if w[1] != 0 {
		return 0, ErrBadBCD
	}

	return w[0], nil
}

// NewDecoder creates new Decoder from BCD configuration.
// If the configuration is invalid NewDecoder will panic.
func NewDecoder(config *BCD) *Decoder {
	if !checkBCD(config) {
		panic("BCD table is incorrect")
	}

	return &Decoder{
		hashWord: newHashDecWord(config),
		hashByte: newHashDecByte(config)}
}

// DecodedLen tells how much space is needed to
// store decoded string. Please note that it returns
// the max amount of possibly needed space because
// last octet may contain only one encoded digit.
// In that case the decoded length will be less by 1.
// For example, 4 octets may encode 7 or 8 digits.
// Please examine the result of Decode to obtain
// the real value.
func DecodedLen(x int) int {
	return 2 * x
}

// Decode parses BCD encoded bytes from src and tries to
// decode them to dst. Number of decoded bytes and possible
// error is returned.
func (dec *Decoder) Decode(dst, src []byte) (int, error) {
	dst = dst[:0]

	if len(src) == 0 {
		return 0, nil
	}

	for _, c := range src[:len(src)-1] {
		w, err := dec.unpackWord(c)
		if err != nil {
			return 0, ErrBadBCD
		}
		dst = append(dst, w[:]...)
	}

	c := src[len(src)-1]
	w, err := dec.unpackWord(c)
	if err != nil {
		// unable to unpack a word
		// maybe it's a finishing byte
		b, err := dec.unpackLastByte(c)
		if err != nil {
			return 0, ErrBadBCD
		}
		dst = append(dst, b)
	} else {
		dst = append(dst, w[:]...)
	}
	return len(dst), nil
}

type Reader struct {
	*Decoder
	src io.Reader
	buf *bytes.Buffer
}
