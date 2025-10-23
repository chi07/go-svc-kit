package sortx

import "strings"

type SortField struct {
	Field string
	Desc  bool
}

func Parse(raw string) []SortField {
	raw = strings.TrimSpace(raw)
	if raw == "" {
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
		if i := strings.IndexRune(it, ':'); i >= 0 {
			field = strings.TrimSpace(it[:i])
			dir := strings.ToLower(strings.TrimSpace(it[i+1:]))
			desc = dir == "desc" || dir == "d" || dir == "descending"
		}
		if field != "" {
			out = append(out, SortField{Field: field, Desc: desc})
		}
	}
	return out
}
