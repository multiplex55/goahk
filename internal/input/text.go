package input

import (
	"fmt"
	"strconv"
)

func DecodeEscapes(raw string) (string, error) {
	decoded, err := strconv.Unquote("\"" + raw + "\"")
	if err != nil {
		return "", fmt.Errorf("decode escapes: %w", err)
	}
	return decoded, nil
}

func ChunkByRune(text string, chunkSize int) []string {
	if chunkSize <= 0 || text == "" {
		if text == "" {
			return nil
		}
		return []string{text}
	}
	runes := []rune(text)
	out := make([]string, 0, (len(runes)+chunkSize-1)/chunkSize)
	for start := 0; start < len(runes); start += chunkSize {
		end := start + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		out = append(out, string(runes[start:end]))
	}
	return out
}
