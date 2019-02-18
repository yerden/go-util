package bcd

import (
// "io"
)

// Encoder is used to encode decimal string into
// BCD bytes.
type Encoder struct {
	// symbol to nibble mapping; example:
	// '*' -> 0xA
	// the value > 0xf means no mapping, i.e. invalid symbol
	hash [0x100]byte

	// nibble used to fill if the number of bytes is odd
	filler byte

	// if true the 0x45 translates to '54' and vice versa
	swap bool
}

func checkBCD(config *BCD) bool {
	nibbles := make(map[byte]bool)
	// check all nibbles
	for _, nib := range config.Map {
		if _, ok := nibbles[nib]; ok || nib > 0xf {
			// already in map or not a nibble
			return false
		}
		nibbles[nib] = true
	}
	return config.Filler <= 0xf
}

func newHashEnc(config *BCD) (res [0x100]byte) {
	for i := 0; i < 0x100; i++ {
		c, ok := config.Map[byte(i)]
		if !ok {
			// no matching symbol
			c = 0xff
		}
		res[i] = c
	}
	return
}

// NewEncoder creates new Encoder from BCD configuration.
// If the configuration is invalid NewEncoder will panic.
func NewEncoder(config *BCD) *Encoder {
	if !checkBCD(config) {
		panic("BCD table is incorrect")
	}
	return &Encoder{
		hash:   newHashEnc(config),
		filler: config.Filler,
		swap:   config.SwapNibble}
}

func (enc *Encoder) packNibs(nib1, nib2 byte) byte {
	if enc.swap {
		return (nib2 << 4) + nib1&0xf
	} else {
		return (nib1 << 4) + nib2&0xf
	}
}

// encode w[0] and w[1]
func (enc *Encoder) packWord(w []byte) (byte, error) {
	nib1, nib2 := enc.hash[w[0]], enc.hash[w[1]]
	if nib1 > 0xf || nib2 > 0xf {
		return 0, ErrBadInput
	}
	return enc.packNibs(nib1, nib2), nil
}

// encode w[0] and filler
func (enc *Encoder) packLastByte(b byte) (byte, error) {
	nib1, nib2 := enc.hash[b], enc.filler
	if nib1 > 0xf {
		return 0, ErrBadInput
	}
	return enc.packNibs(nib1, nib2), nil
}

// EncodedLen returns amount of space needed to store
// bytes after encoding data of length x.
func EncodedLen(x int) int {
	return (x + 1) / 2
}

// Encode get input bytes from src and encodes them into
// BCD data. Number of encoded bytes and possible
// error is returned.
func (enc *Encoder) Encode(dst, src []byte) (int, error) {
	dst = dst[:0]
	var err error
	var b byte

	for {
		switch len(src) {
		case 0:
			return len(dst), nil
		case 1:
			b, err = enc.packLastByte(src[0])
			dst = append(dst, b)
			return len(dst), err
		default:
			if b, err = enc.packWord(src[:2]); err != nil {
				return 0, err
			}
			dst = append(dst, b)
			src = src[2:]
		}
	}
}
