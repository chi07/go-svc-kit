package field

import "strings"

func MakeSet(fields []string) map[string]struct{} {
	if len(fields) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		set[f] = struct{}{}
	}
	return set
}

func FilterAllowedFields(selected []string, allowedSet map[string]struct{}, allDefault []string) []string {
	if len(selected) == 0 {
		return allDefault
	}
	out := make([]string, 0, len(selected))
	seen := make(map[string]struct{}, len(selected))
	for _, f := range selected {
		if _, ok := allowedSet[f]; ok {
			if _, dup := seen[f]; !dup {
				seen[f] = struct{}{}
				out = append(out, f)
			}
		}
	}
	return out
}

func FilterAllowedFieldsNorm(selected []string, allowedSetNorm map[string]struct{}, allDefault []string) []string {
	if len(selected) == 0 {
		return allDefault
	}
	out := make([]string, 0, len(selected))
	seen := make(map[string]struct{}, len(selected))
	for _, f := range selected {
		f = strings.TrimSpace(strings.ToLower(f))
		if _, ok := allowedSetNorm[f]; ok {
			if _, dup := seen[f]; !dup {
				seen[f] = struct{}{}
				out = append(out, f)
			}
		}
	}
	return out
}

func MakeSetLower(fields []string) map[string]struct{} {
	set := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		set[strings.ToLower(f)] = struct{}{}
	}
	return set
}

func FilterAllowedFieldsWithAlias(selected []string, alias map[string]string, allowedAliasSet map[string]struct{}, allDefault []string) []string {
	if len(selected) == 0 {
		return allDefault
	}
	out := make([]string, 0, len(selected))
	seen := make(map[string]struct{}, len(selected))
	for _, f := range selected {
		key := strings.TrimSpace(strings.ToLower(f))
		if _, ok := allowedAliasSet[key]; !ok {
			continue
		}
		col := alias[key]
		if _, dup := seen[col]; !dup {
			seen[col] = struct{}{}
			out = append(out, col)
		}
	}
	return out
}
