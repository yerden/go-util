package common

import (
	"bufio"
	"bytes"
	"errors"
	"unicode"
	"unicode/utf8"
)

var (
	ErrUnprintable = errors.New("unprintable char")
	ErrOpenQuote   = errors.New("no closing quote")

	SplitWithQuotes bufio.SplitFunc = SplitWithQuotesFunc(unicode.IsSpace,
		func(r rune) bool {
			return unicode.In(r, unicode.Quotation_Mark)
		})
	SplitWithDoubleQuotes bufio.SplitFunc = SplitWithQuotesFunc(unicode.IsSpace,
		func(r rune) bool { return r == '"' })
)

func skip(data []byte, f func(rune) bool) (ret []byte, advance int) {
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
		data, n := skip(data, isSpace)
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

type Splitter struct {
	// True if rune is a white space.
	IsSpace func(rune) bool

	// True if rune is quote. Any rune embraced by the one of these
	// pairs is considered a part of a token even if IsSpace returns
	// true.  A pairs must not contradict white space and another
	// pair.
	//
	// If true, return closing quote rune.
	IsQuote func(rune) (rune, bool)

	// If true, final token is allowed not to contain closing quote.
	// If false, ErrOpenQuote error will be returned if no closing
	// quote found.
	AllowOpenQuote bool
}

func (s *Splitter) SplitFunc() bufio.SplitFunc {
	isSpaceOrQuote = func(r rune) bool {
		return s.IsSpace(r) || s.IsQuote(r)
	}

	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// skip whitespace
		data, n := skip(data, s.IsSpace)
		advance += n

		// finished here
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		n = 0 // offset
		unquote := rune(utf8.RuneError)

		n = bytes.IndexFunc(data, isSpaceOrQuote)
		if n
		; n >= 0 {
			return advance + n, data[:n], nil
		}

		switch {
		case n < 0:
			if !atEOF {
				return advance, nil, nil
			}
			fallthrough
		case s.IsSpace(data[n]):
		}
		case s.IsQuote(data[n]):
			// opening quote
			return advance + n, data[:n], nil
		}

		for {
			r, wid := utf8.DecodeRune(data[n:])

			// finished another token
			if s.IsSpace(r) && unquote == utf8.RuneError {
				return advance + n, data[:n], nil
			}

			if unquote == r && unquote != utf8.RuneError {
				// close quote
				unquote = utf8.RuneError
			} else if q, ok := s.IsQuote(r); ok {
				// open quote
				unquote = q
			}

			// case r != utf8.RuneError:
			// break
			// case wid == 0:
			// if !atEOF {
			// return advance, nil, nil
			// }
			// return advance + n, data[:n], nil
			// case wid == 1:
			// return 0, nil, ErrUnprintable
			// default:
			// }
			n += wid
		}
		// unreachable
		return 0, nil, nil
	}
}
