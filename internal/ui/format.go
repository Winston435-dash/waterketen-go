package ui

import (
	"strconv"
	"strings"
)

// Thousands formats counts with Dutch thousand separators.
func Thousands(n int) string {
	s := strconv.Itoa(n)
	sign := ""
	if strings.HasPrefix(s, "-") {
		sign, s = "-", s[1:]
	}
	var out strings.Builder
	out.WriteString(sign)
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out.WriteByte('.')
		}
		out.WriteRune(r)
	}
	return out.String()
}
