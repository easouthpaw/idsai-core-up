package strutil

// TruncateUTF8 truncates s to at most maxRunes Unicode code points.
// Unlike s[:n], this never splits a multi-byte character.
func TruncateUTF8(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes])
}
