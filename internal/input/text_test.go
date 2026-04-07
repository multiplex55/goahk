package input

import (
	"reflect"
	"testing"
)

func TestDecodeEscapes(t *testing.T) {
	got, err := DecodeEscapes(`line1\n🙂\u0021`)
	if err != nil {
		t.Fatalf("decode err: %v", err)
	}
	if got != "line1\n🙂!" {
		t.Fatalf("decoded=%q", got)
	}
}

func TestChunkByRune_UnicodeSafe(t *testing.T) {
	chunks := ChunkByRune("A🙂BC界", 2)
	want := []string{"A🙂", "BC", "界"}
	if !reflect.DeepEqual(chunks, want) {
		t.Fatalf("chunks=%v want=%v", chunks, want)
	}
}
