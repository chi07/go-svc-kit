package parsex_test

import (
	"reflect"
	"testing"

	"github.com/chi07/go-svc-kit/queryx"
)

func TestSanitizeCols(t *testing.T) {
	allowed := map[string]struct{}{"id": {}, "name": {}, "email": {}}

	tests := []struct {
		name string
		in   []string
		want []string
		// nil/empty expectations
		wantNil    bool
		wantNonNil bool
	}{
		{"nil input -> nil", nil, nil, true, false},
		{"empty slice -> empty (non-nil)", []string{}, []string{}, false, true},
		{"trim and filter allowed", []string{" id ", "name", "x", "email", "  "}, []string{"id", "name", "email"}, false, true},
		{"all disallowed -> empty", []string{"x", "y"}, []string{}, false, true},
		{"preserve order", []string{"email", "id"}, []string{"email", "id"}, false, true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := queryx.SanitizeCols(allowed, tc.in)
			if tc.wantNil && got != nil {
				t.Fatalf("expected nil, got %#v", got)
			}
			if tc.wantNonNil && got == nil {
				t.Fatalf("expected non-nil slice, got nil")
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestSplitCSV(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    []string
		wantNil bool
	}{
		{"empty -> nil", "", nil, true},
		{"simple", "a,b,c", []string{"a", "b", "c"}, false},
		{"trim spaces", " a ,  b, c  ", []string{"a", "b", "c"}, false},
		{"skip empties", ",,a,,b,,", []string{"a", "b"}, false},
		{"unicode", "văn, hóa , 😀", []string{"văn", "hóa", "😀"}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := queryx.SplitCSV(tc.in)
			if tc.wantNil && got != nil {
				t.Fatalf("expected nil, got %#v", got)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestStringsCSV(t *testing.T) {
	got := queryx.StringsCSV(" a, b ,, c ")
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("StringsCSV mismatch: want %v, got %v", want, got)
	}
}

func TestIntFrom(t *testing.T) {
	tests := []struct {
		name string
		in   string
		def  int
		want int
	}{
		{"valid positive", "42", 7, 42},
		{"trim spaces", "  5  ", 1, 5},
		{"zero -> def", "0", 9, 9},
		{"negative -> def", "-3", 8, 8},
		{"invalid -> def", "x", 11, 11},
		{"empty -> def", "", 2, 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := queryx.IntFrom(tc.in, tc.def); got != tc.want {
				t.Fatalf("IntFrom(%q,%d) = %d; want %d", tc.in, tc.def, got, tc.want)
			}
		})
	}
}

func TestBoolFrom(t *testing.T) {
	tests := []struct {
		name string
		in   string
		def  bool
		want bool
	}{
		{"true 1", "1", false, true},
		{"true t", "t", false, true},
		{"true true", "true", false, true},
		{"true yes", "yes", false, true},
		{"true y", "y", false, true},

		{"false 0", "0", true, false},
		{"false f", "f", true, false},
		{"false false", "false", true, false},
		{"false no", "no", true, false},
		{"false n", "n", true, false},

		{"default when unknown", "maybe", true, true},
		{"default when empty", "", false, false},
		{"case-insensitive", "TrUe", false, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := queryx.BoolFrom(tc.in, tc.def); got != tc.want {
				t.Fatalf("BoolFrom(%q,%v) = %v; want %v", tc.in, tc.def, got, tc.want)
			}
		})
	}
}

func TestBoolOrNil(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want *bool // nil means expect nil
	}{
		{"empty -> nil", "", nil},
		{"true 1", "1", ptr(true)},
		{"true yes", "yes", ptr(true)},
		{"true TRUE", "TRUE", ptr(true)},
		{"false 0", "0", ptr(false)},
		{"false n", "n", ptr(false)},
		{"unknown -> nil", "perhaps", nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := queryx.BoolOrNil(tc.in)
			if tc.want == nil && got != nil {
				t.Fatalf("want nil, got %v", *got)
			}
			if tc.want != nil {
				if got == nil {
					t.Fatalf("want %v, got nil", *tc.want)
				}
				if *got != *tc.want {
					t.Fatalf("want %v, got %v", *tc.want, *got)
				}
			}
		})
	}
}

func ptr[B any](b B) *B { return &b }
