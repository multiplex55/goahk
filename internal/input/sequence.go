package input

import (
	"fmt"
	"strings"
)

type Sequence struct {
	Tokens []Token
}

type Token struct {
	Raw  string
	Keys []string
}

func ParseSequence(raw string) (Sequence, error) {
	tokens, err := TokenizeSequence(raw)
	if err != nil {
		return Sequence{}, err
	}
	return Sequence{Tokens: tokens}, nil
}

func TokenizeSequence(raw string) ([]Token, error) {
	parts, err := splitSequence(raw)
	if err != nil {
		return nil, err
	}
	out := make([]Token, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}
		keys := strings.Split(part, "+")
		norm := make([]string, 0, len(keys))
		for _, key := range keys {
			key = strings.TrimSpace(key)
			if key == "" {
				return nil, fmt.Errorf("invalid key token %q", part)
			}
			norm = append(norm, strings.ToLower(key))
		}
		out = append(out, Token{Raw: part, Keys: norm})
	}
	return out, nil
}

func splitSequence(raw string) ([]string, error) {
	var out []string
	for i := 0; i < len(raw); {
		for i < len(raw) && (raw[i] == ' ' || raw[i] == '\t' || raw[i] == '\n') {
			i++
		}
		if i >= len(raw) {
			break
		}
		if raw[i] == '{' {
			j := i + 1
			for j < len(raw) && raw[j] != '}' {
				j++
			}
			if j >= len(raw) {
				return nil, fmt.Errorf("unterminated brace token")
			}
			tok := strings.TrimSpace(raw[i+1 : j])
			if tok == "" {
				return nil, fmt.Errorf("empty brace token")
			}
			out = append(out, tok)
			i = j + 1
			continue
		}
		j := i
		for j < len(raw) && raw[j] != ' ' && raw[j] != '\t' && raw[j] != '\n' {
			j++
		}
		out = append(out, raw[i:j])
		i = j
	}
	return out, nil
}
