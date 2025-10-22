package query

import (
	"strconv"
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

type SortField struct {
	Field string
	Desc  bool
}

func ParseBool(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "t", "true", "yes", "y":
		return true
	default:
		return false
	}
}

func ParseInt(raw string, def int) int {
	if raw == "" {
		return def
	}
	i, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || i < 0 {
		return def
	}
	return i
}

func ParseCSV(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func ParseSort(raw string) []SortField {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	items := strings.Split(raw, ",")
	out := make([]SortField, 0, len(items))
	for _, it := range items {
		it = strings.TrimSpace(it)
		if it == "" {
			continue
		}
		field := it
		desc := false
		if i := strings.IndexByte(it, ':'); i >= 0 {
			field = strings.TrimSpace(it[:i])
			dir := strings.ToLower(strings.TrimSpace(it[i+1:]))
			desc = dir == "desc" || dir == "descending" || dir == "d"
		}
		if field != "" {
			out = append(out, SortField{Field: field, Desc: desc})
		}
	}
	if len(out) == 0 {
		return nil
	}

	return out
}
