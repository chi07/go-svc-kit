package query

import (
	"fmt"
	"strings"
)

func ParseCSVFields(input string) []string {
	if input == "" {
		return nil
	}
	parts := strings.Split(input, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

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

func ParseSort(sort string, allowed map[string]struct{}, def string) []string {
	if sort == "" {
		return []string{def}
	}
	parts := strings.Split(sort, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		dir, col := "ASC", p
		if strings.HasPrefix(p, "-") {
			dir = "DESC"
			col = p[1:]
		}
		if _, ok := allowed[col]; ok {
			out = append(out, fmt.Sprintf("%s %s", col, dir))
		}
	}
	if len(out) == 0 {
		out = []string{def}
	}
	return out
}
