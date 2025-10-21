package httpreq

import (
	"strconv"
	"strings"
)

func SplitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func StringsCSV(s string) []string { return SplitCSV(s) }

func IntFrom(s string, def int) int {
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err == nil && v > 0 {
		return v
	}
	return def
}

func BoolFrom(s string, def bool) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "t", "true", "yes", "y":
		return true
	case "0", "f", "false", "no", "n":
		return false
	default:
		return def
	}
}
