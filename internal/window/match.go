package window

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Matcher describes window filtering criteria.
type Matcher struct {
	TitleExact    string
	TitleContains string
	TitleRegex    string
	ClassName     string
	ExeName       string
	ActiveOnly    bool
}

type compiledMatcher struct {
	matcher Matcher
	titleRE *regexp.Regexp
}

func compileMatcher(m Matcher) (compiledMatcher, error) {
	cm := compiledMatcher{matcher: m}
	if strings.TrimSpace(m.TitleRegex) != "" {
		re, err := regexp.Compile(m.TitleRegex)
		if err != nil {
			return compiledMatcher{}, fmt.Errorf("invalid title regex %q: %w", m.TitleRegex, err)
		}
		cm.titleRE = re
	}
	return cm, nil
}

func (m Matcher) matchTitle(info Info, re *regexp.Regexp) bool {
	switch {
	case m.TitleExact != "":
		return strings.EqualFold(info.Title, m.TitleExact)
	case m.TitleContains != "":
		return strings.Contains(strings.ToLower(info.Title), strings.ToLower(m.TitleContains))
	case m.TitleRegex != "":
		return re != nil && re.MatchString(info.Title)
	default:
		return true
	}
}

func (m Matcher) matches(info Info, re *regexp.Regexp) bool {
	if m.ActiveOnly && !info.Active {
		return false
	}
	if !m.matchTitle(info, re) {
		return false
	}
	if m.ClassName != "" && !strings.EqualFold(info.Class, m.ClassName) {
		return false
	}
	if m.ExeName != "" && !strings.EqualFold(info.Exe, m.ExeName) {
		return false
	}
	return true
}

// Filter returns windows matching the criteria. Enumeration order is preserved.
func Filter(windows []Info, m Matcher) ([]Info, error) {
	cm, err := compileMatcher(m)
	if err != nil {
		return nil, err
	}
	out := make([]Info, 0, len(windows))
	for _, win := range windows {
		if cm.matcher.matches(win, cm.titleRE) {
			out = append(out, win)
		}
	}
	return out, nil
}

// MatchesWithTimeout performs a single-window match and allows callers to enforce a timeout budget.
func MatchesWithTimeout(info Info, m Matcher, timeout time.Duration) (bool, error) {
	if timeout <= 0 {
		matches, err := Filter([]Info{info}, m)
		return len(matches) == 1, err
	}

	type result struct {
		ok  bool
		err error
	}
	ch := make(chan result, 1)
	go func() {
		matches, err := Filter([]Info{info}, m)
		ch <- result{ok: len(matches) == 1, err: err}
	}()

	select {
	case res := <-ch:
		return res.ok, res.err
	case <-time.After(timeout):
		return false, fmt.Errorf("window match timed out after %s", timeout)
	}
}
