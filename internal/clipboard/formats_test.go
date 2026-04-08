package clipboard

import (
	"errors"
	"reflect"
	"testing"
)

func TestEncodeUTF16_AppendsSingleNULTerminator(t *testing.T) {
	t.Parallel()

	got := EncodeUTF16("hello\x00\x00")
	want := []uint16{'h', 'e', 'l', 'l', 'o', 0}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("EncodeUTF16() = %#v, want %#v", got, want)
	}
}

func TestDecodeUTF16_EmptyAndTerminatorOnly(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		units []uint16
		want  string
	}{
		{name: "empty", units: nil, want: ""},
		{name: "single nul", units: []uint16{0}, want: ""},
		{name: "double nul", units: []uint16{0, 0}, want: "\x00"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := DecodeUTF16(tc.units)
			if err != nil {
				t.Fatalf("DecodeUTF16(%#v) err = %v", tc.units, err)
			}
			if got != tc.want {
				t.Fatalf("DecodeUTF16(%#v) = %q, want %q", tc.units, got, tc.want)
			}
		})
	}
}

func TestDecodeUTF16_InvalidSurrogateFails(t *testing.T) {
	t.Parallel()

	_, err := DecodeUTF16([]uint16{0xD83D})
	if !errors.Is(err, ErrInvalidUTF16Data) {
		t.Fatalf("DecodeUTF16() err = %v, want %v", err, ErrInvalidUTF16Data)
	}
}

func TestUTF16RoundTripAndNormalizationBoundaries(t *testing.T) {
	t.Parallel()

	text := "prefix🙂suffix\x00\x00"
	encoded := EncodeUTF16(text)
	decoded, err := DecodeUTF16(encoded)
	if err != nil {
		t.Fatalf("DecodeUTF16() err = %v", err)
	}
	if decoded != "prefix🙂suffix" {
		t.Fatalf("DecodeUTF16() = %q, want %q", decoded, "prefix🙂suffix")
	}

	if got := NormalizeReadText("value\x00\x00"); got != "value" {
		t.Fatalf("NormalizeReadText() = %q, want value", got)
	}
	if got := NormalizeWriteText("value\x00\x00"); got != "value" {
		t.Fatalf("NormalizeWriteText() = %q, want value", got)
	}
}
