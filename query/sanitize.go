package query

import (
	"strings"
)

func SanitizeCols(allowed map[string]struct{}, cols []string) []string {
	if len(cols) == 0 {
		return cols
	}
	out := make([]string, 0, len(cols))
	for _, c := range cols {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		if _, ok := allowed[c]; ok {
			out = append(out, c)
		}
	}
	return out
}
