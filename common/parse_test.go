package common

import (
	"bufio"
	"bytes"
	"testing"
	"testing/iotest"
	"unicode"

	"github.com/yerden/go-util/assert"
)

func TestScanWordsWithQuotes(t *testing.T) {
	a := assert.New(t)
	s := "k1=\"val\" k2=\"val2\""

	b := bytes.NewBufferString(s)
	scanner := bufio.NewScanner(iotest.OneByteReader(b))
	scanner.Split(SplitWithQuotesFunc(unicode.IsSpace,
		func(r rune) bool {
			return unicode.In(r, unicode.Quotation_Mark)
		}))

	a.True(scanner.Scan())
	a.Equal(scanner.Text(), "k1=\"val\"")
	a.True(scanner.Scan())
	a.Equal(scanner.Text(), "k2=\"val2\"")
	a.NotTrue(scanner.Scan())
}

func TestScanWordsWithQuotes_SomeSpaces(t *testing.T) {
	a := assert.New(t)
	s := " k1=\"val\"  k2=\"val2\"    "

	b := bytes.NewBufferString(s)
	scanner := bufio.NewScanner(iotest.OneByteReader(b))
	scanner.Split(SplitWithQuotesFunc(unicode.IsSpace,
		func(r rune) bool {
			return unicode.In(r, unicode.Quotation_Mark)
		}))

	a.True(scanner.Scan())
	a.Equal(scanner.Text(), "k1=\"val\"")
	a.True(scanner.Scan())
	a.Equal(scanner.Text(), "k2=\"val2\"")
	a.NotTrue(scanner.Scan())
}

func TestScanWordsWithQuotes_QuotedSpaces(t *testing.T) {
	a := assert.New(t)
	s := " k1=\"val bye\"  k2=\"val2 hello\"    \"another token\"  "

	b := bytes.NewBufferString(s)
	scanner := bufio.NewScanner(iotest.OneByteReader(b))
	scanner.Split(SplitWithDoubleQuotes)

	a.True(scanner.Scan())
	a.Equal(scanner.Text(), "k1=\"val bye\"")
	a.True(scanner.Scan())
	a.Equal(scanner.Text(), "k2=\"val2 hello\"")
	a.True(scanner.Scan())
	a.Equal(scanner.Text(), "\"another token\"")
	a.NotTrue(scanner.Scan())
}

func TestScanWordsWithQuotes_QuotedSpaces1(t *testing.T) {
	a := assert.New(t)
	s := " k1='val bye'  k2=\"val2 hello\"    \"another token\"  "

	b := bytes.NewBufferString(s)
	scanner := bufio.NewScanner(iotest.OneByteReader(b))
	scanner.Split(SplitWithQuotesFunc(unicode.IsSpace,
		func(r rune) bool {
			return unicode.In(r, unicode.Quotation_Mark)
		}))

	a.True(scanner.Scan())
	a.Equal(scanner.Text(), "k1='val bye'")
	a.True(scanner.Scan())
	a.Equal(scanner.Text(), "k2=\"val2 hello\"")
	a.True(scanner.Scan())
	a.Equal(scanner.Text(), "\"another token\"")
	a.NotTrue(scanner.Scan())
}

func testSplitter(t *testing.T, sample string, expToks []string) {
	a := assert.New(t)

	splitter := &Splitter{
		IsSpace: unicode.IsSpace,
		IsQuote: func(r rune) (rune, bool) {
			if r == '<' {
				return '>', true
			}
			if r == '"' {
				return '"', true
			}
			return ' ', false
		},
		AllowOpenQuote: false}

	b := bytes.NewBufferString(sample)
	s := bufio.NewScanner(b)
	s.Split(splitter.SplitFunc())

	for _, tok := range expToks {
		a.True(s.Scan())
		a.Equal(s.Text(), tok)
	}
	a.NotTrue(s.Scan())
	a.Nil(s.Err())
}

func TestSplitter(t *testing.T) {

	testSplitter(t,
		"   \"good bye\" hello <how hello> do <hell is real>   ",
		[]string{
			"\"good bye\"",
			"hello",
			"<how hello>",
			"do",
			"<hell is real>",
		})

	testSplitter(t,
		"\"good bye\"",
		[]string{
			"\"good bye\"",
		})
}

func TestSplitter1(t *testing.T) {
	a := assert.New(t)

	splitter := &Splitter{
		IsSpace: unicode.IsSpace,
		IsQuote: func(r rune) (rune, bool) {
			if r == '<' {
				return '>', true
			}
			if r == '"' {
				return '"', true
			}
			return ' ', false
		},
		AllowOpenQuote: false}

	sample := "   \"good bye\" hello <how hello> do <hell is real>   "

	b := bytes.NewBufferString(sample)
	s := bufio.NewScanner(b)
	s.Split(splitter.SplitFunc())

	a.True(s.Scan())
	a.Equal(s.Text(), "\"good bye\"")
	a.True(s.Scan())
	a.Equal(s.Text(), "hello")
	a.True(s.Scan())
	a.Equal(s.Text(), "<how hello>")
	a.True(s.Scan())
	a.Equal(s.Text(), "do")
	a.True(s.Scan())
	a.Equal(s.Text(), "<hell is real>")
	a.NotTrue(s.Scan())
}
