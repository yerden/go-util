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
		_, ok := s.IsQuote(r)
		return ok || s.IsSpace(r)
	}

	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// skip whitespace
		data, n := skip(data, s.IsSpace)
		advance += n

		// finished here
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		// start of a token
		n = 0 // length of token

		for {
			k := bytes.IndexFunc(data[n:], isSpaceOrQuote)
			if k < 0 {
				goto UNTERMINATED
			}

			// rune 'r' is either space or quote
			r, wid := utf8.DecodeRune(data[n+k:])

			// if quote, then look for closing quote
			if q, ok := s.IsQuote(r); ok {
				n += k + wid
				k = bytes.IndexRune(data[n:], r)
				if k < 0 {
					goto UNTERMINATED
				}
				n += k + wid
			}

			if s.IsSpace(r) {
				if !quoted {
					return advance + n + k + 1, data[:n+k], nil
				}
				n += k + wid
				continue
			}
		}

	UNTERMINATED:
		if !atEOF {
			// unterminated token, need more data
			return advance, nil, nil
		} else if !quoted || s.AllowOpenQuote {
			// unterminated token, no more data
			return advance + len(data), data, nil
		} else {
			// unterminated quote is not allowed
			return advance + len(data), data, ErrOpenQuote
		}

	}
}
