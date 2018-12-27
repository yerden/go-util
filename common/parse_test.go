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
