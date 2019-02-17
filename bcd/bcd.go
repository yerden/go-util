/*
Package bcd provides functions to encode byte arrays
to BCD (Binary-Coded Decimal) encoding and back.
*/
package bcd

import (
	"fmt"
)

// BCD is the configuration for Binary-Coded Decimal
// encoding.
type BCD struct {
	// Map of symbols to encode and decode routines.
	// Example:
	//    key 'a' -> value 0x9
	Map map[byte]byte

	// If true nibbles (4-bit part of a byte) will
	// be swapped, meaning bits 0123 will encode
	// first digit and bits 4567 will encode the
	// second.
	SwapNibble bool

	// Filler nibble is used if the input has odd
	// number of bytes. Then the output's final nibble
	// will contain the specified nibble.
	Filler byte
}

var (
	// Standard 8-4-2-1 decimal-only encoding.
	StdEncoding = &BCD{
		Map: map[byte]byte{
			'0': 0x0, '1': 0x1, '2': 0x2, '3': 0x3,
			'4': 0x4, '5': 0x5, '6': 0x6, '7': 0x7,
			'8': 0x8, '9': 0x9,
		},
		SwapNibble: false,
		Filler:     0xf}

	// Excess-3 or Stibitz encoding.
	Excess3Encoding = &BCD{
		Map: map[byte]byte{
			'0': 0x3, '1': 0x4, '2': 0x5, '3': 0x6,
			'4': 0x7, '5': 0x8, '6': 0x9, '7': 0xa,
			'8': 0xb, '9': 0xc,
		},
		SwapNibble: false,
		Filler:     0x0}

	// TBCD (Telephony BCD) as in 3GPP TS 29.002.
	TBCDEncoding = &BCD{
		Map: map[byte]byte{
			'0': 0x0, '1': 0x1, '2': 0x2, '3': 0x3,
			'4': 0x4, '5': 0x5, '6': 0x6, '7': 0x7,
			'8': 0x8, '9': 0x9, '*': 0xa, '#': 0xb,
			'a': 0xc, 'b': 0xd, 'c': 0xe,
		},
		SwapNibble: true,
		Filler:     0xf}
)

var (
	ErrBadInput = fmt.Errorf("non-encodable data")
	ErrBadBCD   = fmt.Errorf("Bad BCD data")
)

// PlainEncode is a simple but not efficient implementation
// of BCD encoding. It is preserved for historical and
// educational purposes. The user might want to use Encoder
// instead.
func (enc *BCD) PlainEncode(in, out []byte) ([]byte, error) {
	if !checkBCD(enc) {
		panic("BCD table is incorrect")
	}
	out = append(out[:0], make([]byte, 1+len(in)/2)...)
	out = out[:0]
	var b, b1, b2 byte

	for {
		switch len(in) {
		case 0:
			return out, nil
		case 1:
			b1, in = in[0], in[1:]
			b2 = enc.Filler
		default:
			b1, b2, in = in[0], in[1], in[2:]
			if b, ok := enc.Map[b2]; !ok {
				return nil, ErrBadInput
			} else {
				b2 = b
			}
		}
		if b, ok := enc.Map[b1]; !ok {
			return nil, ErrBadInput
		} else {
			b1 = b
		}

		if enc.SwapNibble {
			b = (b2 << 4) + b1&0xf
		} else {
			b = (b1 << 4) + b2&0xf
		}
		out = append(out, b)
	}
}

func reverse(m map[byte]byte) map[byte]byte {
	res := make(map[byte]byte)
	for k, v := range m {
		res[v] = k
	}
	return res
}

// PlainDecode is a simple but not efficient implementation
// of BCD decoding. It is preserved for historical and
// educational purposes. The user might want to use Decoder
// instead.
func (enc *BCD) PlainDecode(in, out []byte) ([]byte, error) {
	if !checkBCD(enc) {
		panic("BCD table is incorrect")
	}
	out = append(out[:0], make([]byte, 2*len(in))...)
	out = out[:0]
	hash := reverse(enc.Map)
	var b1, b2 byte

	for i, b := range in {
		if enc.SwapNibble {
			b1, b2 = b&0xf, b>>4
			// if b2 == enc.Filler
		} else {
			b2, b1 = b&0xf, b>>4
		}

		// first byte
		if b1 == enc.Filler {
			// unexpected end of data
			return nil, ErrBadBCD
		}
		if b, ok := hash[b1]; !ok {
			return nil, ErrBadBCD
		} else {
			b1 = b
		}

		// second byte
		if b2 == enc.Filler {
			if i == len(in)-1 {
				out = append(out, b1)
				return out, nil
			} else {
				return nil, ErrBadBCD
			}
		} else if b, ok := hash[b2]; !ok {
			return nil, ErrBadBCD
		} else {
			b2 = b
		}
		out = append(out, []byte{b1, b2}...)
	}

	return out, nil
}
