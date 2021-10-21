package colorize

import (
	"bytes"
	"math/rand"
	"testing"
)

func TestColorizer(t *testing.T) {
	c := NewColorizer(bytes.NewBuffer([]byte("foo bar")))
	if _, err := c.Next(); err != nil {
		t.Error("unexpected error", err)
	}

	// Verify success on huge input.
	c = NewColorizer(bytes.NewBuffer(make([]byte, 16*1024)))
	if _, err := c.Next(); err != nil {
		t.Error("unexpected error", err)
	}
}

func TestTokenizer(t *testing.T) {
	tests := []struct {
		input  string
		tokens []string
		words  []bool
	}{
		{"", []string{""}, []bool{false}},
		{
			"foo bar",
			[]string{"foo", " ", "bar", ""},
			[]bool{true, false, true, false},
		},
		{
			`foo "bar"`,
			[]string{"foo", " ", `"`, "bar", `"`, ""},
			[]bool{true, false, false, true, false, false},
		},
		{
			`"escaped quote \""`,
			[]string{`"`, `escaped quote \"`, `"`, ""},
			[]bool{false, true, false, false},
		},
		{
			"foo   bar",
			[]string{"foo", "   ", "bar", ""},
			[]bool{true, false, true, false},
		},
	}

	for _, tt := range tests {
		if len(tt.tokens) != len(tt.words) {
			t.Errorf("mismatch got=%v want=%v", len(tt.tokens), len(tt.words))
		}

		tz := newTokenizer([]byte(tt.input))
		for i, wantTkn := range tt.tokens {
			gotTkn, gotWord := tz.chunk()
			if string(gotTkn) != string(wantTkn) {
				t.Errorf("mismatch i=%d got=%q want=%q", i, string(gotTkn), string(wantTkn))
			} else if gotWord != tt.words[i] {
				t.Errorf("mismatch i=%d got=%t want=%t", i, gotWord, tt.words[i])
			}
		}
	}
}

func TestColor(t *testing.T) {
	min := color(nil)
	max := color(nil)
	for i := 0; i < 10000; i++ {
		var buf [16]byte
		rand.Read(buf[:])
		c := color(buf[:])
		if min > c {
			min = c
		}
		if max < c {
			max = c
		}
	}

	if min != minColor {
		t.Errorf("mismatch got=%v want=%v", min, minColor)
	}
	if max != maxColor {
		t.Errorf("mismatch got=%v want=%v", max, maxColor)
	}
}
