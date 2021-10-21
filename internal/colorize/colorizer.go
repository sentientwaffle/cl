// Package colorize colors terminal log streams.
package colorize

import (
	"bufio"
	"encoding/binary"
	"hash/fnv"
	"io"
	"strconv"
)

type Colorizer struct {
	// Use bufio.Reader instead of bufio.Scanner so that lines that are too
	// long to be buffered are handled without an error.
	r   *bufio.Reader
	out []byte
}

func NewColorizer(r io.Reader) Colorizer {
	const scanBufSize = 1024 * 16
	b := bufio.NewReaderSize(r, scanBufSize)
	return Colorizer{b, make([]byte, scanBufSize)}
}

const (
	setForegroundPrefix = "\u001b[38;5;"
	setForegroundSuffix = "m"
	clearForeground     = "\u001b[39m"
)

func (c Colorizer) Next() ([]byte, error) {
	buf, _, err := c.r.ReadLine()
	if err != nil {
		return nil, io.EOF
	}
	c.out = c.out[:0]
	tz := newTokenizer(buf)
	for {
		chunk, isWord := tz.chunk()
		if len(chunk) == 0 {
			break
		} else if !isWord {
			c.out = append(c.out, chunk...)
			continue
		}
		c.out = append(c.out, setForegroundPrefix...)
		c.out = append(c.out, strconv.Itoa(int(color(chunk)))...)
		c.out = append(c.out, setForegroundSuffix...)
		c.out = append(c.out, chunk...)
		c.out = append(c.out, clearForeground...)
	}
	c.out = append(c.out, '\n')
	return c.out, nil
}

type tokenizer struct {
	buf   []byte
	inStr bool
}

func newTokenizer(buf []byte) *tokenizer {
	return &tokenizer{buf, false}
}

func (t *tokenizer) chunk() (chunk []byte, isWord bool) {
	switch {
	case len(t.buf) == 0:
		return nil, false
	case t.buf[0] == '"':
		t.inStr = !t.inStr
		chunk = t.buf[:1]
	case t.inStr:
		isWord = true
		var esc bool
		for i, ch := range t.buf {
			if esc {
				esc = !esc
			} else if ch == '\\' {
				esc = true
			} else if ch == '"' {
				chunk = t.buf[:i]
				break
			}
		}
	case isWordByte(t.buf[0]):
		chunk = takeWhile(t.buf, isWordByte)
		isWord = true
	default:
		chunk = takeWhile(t.buf, func(b byte) bool {
			return b != '"' && !isWordByte(b)
		})
	}
	t.buf = t.buf[len(chunk):]
	return
}

func takeWhile(buf []byte, test func(byte) bool) []byte {
	for i, b := range buf {
		if !test(b) {
			return buf[:i]
		}
	}
	return buf
}

var nonWordBytes = []byte{
	'\t', '\n', '\r', ' ',
	'{', '}', '[', ']',
	',', ':', '"',
}

func isWordByte(b byte) bool {
	for _, wb := range nonWordBytes {
		if wb == b {
			return false
		}
	}
	return true
}

const (
	minColor = 19  // inclusive
	maxColor = 230 // inclusive
)

func color(b []byte) uint8 {
	h := hash(b)
	// This isn't uniformly distributed, but close enough.
	return uint8(minColor + h%(1+maxColor-minColor))
}

func hash(b []byte) uint64 {
	h := fnv.New64()
	h.Write(b)
	var out [8]byte
	h.Sum(out[:0])
	return binary.BigEndian.Uint64(out[:])
}
