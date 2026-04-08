package window

import "strings"

// ParseMatcherString converts a compact matcher DSL into a Matcher.
//
// Supported keys: title, title_exact, title_regex, class, exe, active.
// Unknown keys are treated as title contains expressions for compatibility.
func ParseMatcherString(raw string) Matcher {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Matcher{}
	}

	out := Matcher{}
	for _, part := range strings.Split(raw, ",") {
		token := strings.TrimSpace(part)
		if token == "" {
			continue
		}
		key, value, ok := strings.Cut(token, ":")
		if !ok {
			out.TitleContains = token
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		switch key {
		case "title":
			out.TitleContains = value
		case "title_exact":
			out.TitleExact = value
		case "title_regex":
			out.TitleRegex = value
		case "class":
			out.ClassName = value
		case "exe":
			out.ExeName = value
		case "active":
			out.ActiveOnly = strings.EqualFold(value, "true") || value == "1" || strings.EqualFold(value, "yes")
		default:
			out.TitleContains = strings.TrimSpace(token)
		}
	}
	return out
}
