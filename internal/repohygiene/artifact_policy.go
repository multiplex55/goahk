package repohygiene

import "strings"

var blockedBinaryExtensions = []string{".exe", ".dll", ".so", ".dylib"}

const releaseArtifactPrefix = "dist/releases/"

func blockedBinaryFiles(files []string, allowReleaseArtifacts bool) []string {
	var blocked []string
	for _, file := range files {
		if file == "" {
			continue
		}
		lower := strings.ToLower(file)
		if !hasBlockedBinaryExtension(lower) {
			continue
		}
		if allowReleaseArtifacts && strings.HasPrefix(lower, releaseArtifactPrefix) {
			continue
		}
		blocked = append(blocked, file)
	}
	return blocked
}

func hasBlockedBinaryExtension(file string) bool {
	for _, ext := range blockedBinaryExtensions {
		if strings.HasSuffix(file, ext) {
			return true
		}
	}
	return false
}
