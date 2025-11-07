package fieldx_test

import (
	"reflect"
	"testing"

	"github.com/chi07/go-svc-kit/fieldx"
)

func TestMakeSet(t *testing.T) {
	t.Run("empty returns nil", func(t *testing.T) {
		got := fieldx.MakeSet(nil)
		if got != nil {
			t.Fatalf("expected nil, got %#v", got)
		}
		got = fieldx.MakeSet([]string{})
		if got != nil {
			t.Fatalf("expected nil for empty slice, got %#v", got)
		}
	})

	t.Run("non-empty builds set with exact keys", func(t *testing.T) {
		in := []string{"A", "b", "A"} // duplicate should be fine in map
		got := fieldx.MakeSet(in)
		if got == nil {
			t.Fatalf("got nil")
		}
		if len(got) != 2 {
			t.Fatalf("expected 2 keys, got %d", len(got))
		}
		if _, ok := got["A"]; !ok {
			t.Fatalf("missing key A")
		}
		if _, ok := got["b"]; !ok {
			t.Fatalf("missing key b")
		}
	})
}

func TestMakeSetLower(t *testing.T) {
	in := []string{"A", "b", "C", "c"}
	got := fieldx.MakeSetLower(in)
	wantKeys := []string{"a", "b", "c"}
	if len(got) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(got))
	}
	for _, k := range wantKeys {
		if _, ok := got[k]; !ok {
			t.Fatalf("missing lowered key %q", k)
		}
	}
}

func TestFilterAllowedFields_DefaultWhenSelectedEmpty(t *testing.T) {
	allowed := fieldx.MakeSet([]string{"id", "name"})
	def := []string{"id", "name", "email"}

	got := fieldx.FilterAllowedFields(nil, allowed, def)
	if !reflect.DeepEqual(got, def) {
		t.Fatalf("expected default %v, got %v", def, got)
	}

	got = fieldx.FilterAllowedFields([]string{}, allowed, def)
	if !reflect.DeepEqual(got, def) {
		t.Fatalf("expected default for empty selected %v, got %v", def, got)
	}
}

func TestFilterAllowedFields_FiltersAndDedups_PreservesOrder(t *testing.T) {
	allowed := fieldx.MakeSet([]string{"id", "name", "email", "age"})
	selected := []string{"email", "name", "email", "unknown", "id", "id"}

	got := fieldx.FilterAllowedFields(selected, allowed, []string{"DEFAULT"})
	want := []string{"email", "name", "id"} // order as first-seen after filtering, no duplicates

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestFilterAllowedFields_AllFilteredOutWhenNotAllowed(t *testing.T) {
	var allowed map[string]struct{} // nil map: lookups are safe and always false
	selected := []string{"a", "b"}
	got := fieldx.FilterAllowedFields(selected, allowed, []string{"def"})
	if len(got) != 0 {
		t.Fatalf("expected empty when none allowed, got %v", got)
	}
}

func TestFilterAllowedFieldsNorm_DefaultWhenSelectedEmpty(t *testing.T) {
	allowedNorm := fieldx.MakeSetLower([]string{"ID", "Name"})
	def := []string{"id", "name"}

	got := fieldx.FilterAllowedFieldsNorm(nil, allowedNorm, def)
	if !reflect.DeepEqual(got, def) {
		t.Fatalf("expected default %v, got %v", def, got)
	}
}

func TestFilterAllowedFieldsNorm_NormalizesCaseTrimAndDedups(t *testing.T) {
	allowedNorm := fieldx.MakeSetLower([]string{"ID", "Name", "Email"})
	selected := []string{"  ID ", "name", "EMAIL", "email", "NaMe", "unknown"}

	got := fieldx.FilterAllowedFieldsNorm(selected, allowedNorm, []string{"def"})
	want := []string{"id", "name", "email"} // normalized to lower, dedup, preserve first-seen order

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestFilterAllowedFieldsWithAlias_DefaultWhenSelectedEmpty(t *testing.T) {
	alias := map[string]string{"id": "id", "name": "full_name"}
	allowedAlias := fieldx.MakeSet([]string{"id", "name"})
	def := []string{"id"}

	got := fieldx.FilterAllowedFieldsWithAlias(nil, alias, allowedAlias, def)
	if !reflect.DeepEqual(got, def) {
		t.Fatalf("expected default %v, got %v", def, got)
	}
}

func TestFilterAllowedFieldsWithAlias_FiltersMapsAndDedups(t *testing.T) {
	alias := map[string]string{
		"id":    "id",
		"name":  "full_name",
		"email": "email_address",
	}
	allowedAlias := fieldx.MakeSet([]string{"id", "name", "email"}) // keys must be normalized (lower) by caller
	selected := []string{"Name", " EMAIL ", "id", "name", "unknown", "email"}

	got := fieldx.FilterAllowedFieldsWithAlias(selected, alias, allowedAlias, []string{"def"})
	want := []string{"full_name", "email_address", "id"} // order after normalization and aliasing, dedup by resulting col

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestFilterAllowedFieldsWithAlias_SkipsWhenNotInAllowedAlias(t *testing.T) {
	alias := map[string]string{
		"id":   "id",
		"name": "full_name",
	}
	// only "id" allowed; "name" exists in alias but not allowedAliasSet
	allowedAlias := fieldx.MakeSet([]string{"id"})
	selected := []string{"name", "id", "NAME"}

	got := fieldx.FilterAllowedFieldsWithAlias(selected, alias, allowedAlias, nil)
	want := []string{"id"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}
