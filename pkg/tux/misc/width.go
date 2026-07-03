package misc

import "github.com/rivo/uniseg"

func RuneWidth(r rune) int {
	if r == 0 {
		return 0
	}
	width := uniseg.StringWidth(string(r))
	if width < 1 {
		return 1
	}
	if width > 2 {
		return 2
	}
	return width
}

func StringWidth(s string) int {
	return uniseg.StringWidth(s)
}

// PrefixRunes returns the first n runes of string s.
// If n is greater than the number of runes in s, returns s.
// If n is less than or equal to 0, returns empty string.
func PrefixRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if n > len(runes) {
		n = len(runes)
	}
	return string(runes[:n])
}
