package docs_test

import "testing"

func TestMustContainAllAcceptsMultipleMatches(t *testing.T) {
	t.Parallel()
	mustContainAll(t, "alpha beta gamma", "alpha", "gamma")
}

func TestMustContainFindsSubstring(t *testing.T) {
	t.Parallel()
	mustContain(t, "viewer build command", "build")
}
