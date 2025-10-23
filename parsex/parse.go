package parsex

import (
	"fmt"
	"strconv"
	"strings"
)

func Int64(s string, def int64) int64 {
	if s == "" {
		return def
	}
	i, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil || i < 0 {
		return def
	}
	return i
}

func CSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func Bool(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "t", "true", "y", "yes":
		return true
	default:
		return false
	}
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
