package common

import (
	"bufio"
	"errors"
	"unicode"
	"unicode/utf8"
)

var (
	ErrUnprintable = errors.New("unprintable char")

	SplitWithQuotes bufio.SplitFunc = SplitWithQuotesFunc(unicode.IsSpace,
		func(r rune) bool {
			return unicode.In(r, unicode.Quotation_Mark)
		})
	SplitWithDoubleQuotes bufio.SplitFunc = SplitWithQuotesFunc(unicode.IsSpace,
		func(r rune) bool { return r == '"' })
)

func skipFunc(data []byte, f func(rune) bool) (ret []byte, advance int) {
	for {
		ret = data[advance:]
		if r, wid := utf8.DecodeRune(ret); f(r) {
			advance += wid
		} else {
			break
		}
	}
	return
}

// Split given data into words separated by whitespace and
// don't split the words if they're under quotes.
// Whitespace and quote are defined by isSpace and isQuote functions.
func SplitWithQuotesFunc(isSpace, isQuote func(rune) bool) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// skip whitespace
		data, n := skipFunc(data, isSpace)
		advance += n

		// finished here
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		n = 0 // offset
		var quotemark rune = utf8.RuneError
		for {
			r, wid := utf8.DecodeRune(data[n:])
			switch {
			case isSpace(r):
				if quotemark == utf8.RuneError {
					// quote is closed
					return advance + n, data[:n], nil
				}
			case isQuote(r):
				if r == quotemark {
					// close quote
					quotemark = utf8.RuneError
				} else {
					// open quote
					quotemark = r
				}
			case r != utf8.RuneError:
				break
			case wid == 0:
				if !atEOF {
					return advance, nil, nil
				}
				return advance + n, data[:n], nil
			case wid == 1:
				return 0, nil, ErrUnprintable
			default:
			}
			n += wid
		}

		// unreachable
		return 0, nil, nil
	}
}
