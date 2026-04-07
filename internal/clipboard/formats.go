package clipboard

import (
	"errors"
	"strings"
	"unicode/utf16"
)

var ErrInvalidUTF16Data = errors.New("clipboard: invalid UTF-16 data")

// NormalizeReadText trims terminating NUL runes that often appear at the
// UTF-16/UTF-8 API boundary when clipboard payloads are decoded.
func NormalizeReadText(s string) string {
	return strings.TrimRight(s, "\x00")
}

// NormalizeWriteText ensures text is in a consistent form before being written.
func NormalizeWriteText(s string) string {
	return strings.TrimRight(s, "\x00")
}

// DecodeUTF16 decodes UTF-16 code units while honoring an optional terminating
// NUL code unit from native clipboard memory blocks.
func DecodeUTF16(units []uint16) (string, error) {
	if len(units) == 0 {
		return "", nil
	}
	if units[len(units)-1] == 0 {
		units = units[:len(units)-1]
	}
	if len(units) == 0 {
		return "", nil
	}

	buf := make([]rune, 0, len(units))
	for i := 0; i < len(units); i++ {
		u := units[i]
		if utf16.IsSurrogate(rune(u)) {
			if i+1 >= len(units) {
				return "", ErrInvalidUTF16Data
			}
			r := utf16.DecodeRune(rune(u), rune(units[i+1]))
			if r == '\uFFFD' {
				return "", ErrInvalidUTF16Data
			}
			buf = append(buf, r)
			i++
			continue
		}
		buf = append(buf, rune(u))
	}
	return string(buf), nil
}

// EncodeUTF16 converts UTF-8 text into UTF-16 code units terminated by a NUL.
func EncodeUTF16(s string) []uint16 {
	s = NormalizeWriteText(s)
	return append(utf16.Encode([]rune(s)), 0)
}

func AppendText(current, suffix string) string {
	return NormalizeWriteText(NormalizeReadText(current) + NormalizeWriteText(suffix))
}

func PrependText(prefix, current string) string {
	return NormalizeWriteText(NormalizeWriteText(prefix) + NormalizeReadText(current))
}
