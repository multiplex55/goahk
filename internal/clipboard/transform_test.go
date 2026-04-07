package clipboard

import "testing"

func TestAppendText_Edges(t *testing.T) {
	cases := []struct {
		name    string
		current string
		suffix  string
		want    string
	}{
		{name: "unicode", current: "naïve", suffix: "🙂", want: "naïve🙂"},
		{name: "empty suffix", current: "abc", suffix: "", want: "abc"},
		{name: "empty current", current: "", suffix: "x", want: "x"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := AppendText(tc.current, tc.suffix); got != tc.want {
				t.Fatalf("AppendText() = %q want %q", got, tc.want)
			}
		})
	}
}

func TestPrependText_Edges(t *testing.T) {
	cases := []struct {
		name    string
		prefix  string
		current string
		want    string
	}{
		{name: "unicode", prefix: "🙂", current: "世界", want: "🙂世界"},
		{name: "empty prefix", prefix: "", current: "abc", want: "abc"},
		{name: "empty current", prefix: "x", current: "", want: "x"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := PrependText(tc.prefix, tc.current); got != tc.want {
				t.Fatalf("PrependText() = %q want %q", got, tc.want)
			}
		})
	}
}

func TestUTF16BoundaryNormalization(t *testing.T) {
	encoded := EncodeUTF16("hi🙂")
	decoded, err := DecodeUTF16(encoded)
	if err != nil {
		t.Fatalf("DecodeUTF16() err = %v", err)
	}
	if decoded != "hi🙂" {
		t.Fatalf("DecodeUTF16() = %q", decoded)
	}
	if NormalizeReadText("ok\x00\x00") != "ok" {
		t.Fatal("NormalizeReadText did not trim NUL terminators")
	}
}
