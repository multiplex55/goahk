package window

// MatchTitleExact returns a Matcher that matches title case-insensitively by full string.
func MatchTitleExact(title string) Matcher {
	return Matcher{TitleExact: title}
}

// MatchTitleContains returns a Matcher that matches title case-insensitively by substring.
func MatchTitleContains(title string) Matcher {
	return Matcher{TitleContains: title}
}

// MatchClass returns a Matcher that matches class name case-insensitively.
func MatchClass(className string) Matcher {
	return Matcher{ClassName: className}
}

// MatchExe returns a Matcher that matches executable basename case-insensitively.
func MatchExe(exe string) Matcher {
	return Matcher{ExeName: exe}
}
