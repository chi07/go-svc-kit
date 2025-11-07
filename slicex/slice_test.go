package slice_test

import (
	"reflect"
	"testing"

	"github.com/chi07/go-svc-kit/slicex"
)

func TestUniqueStrings(t *testing.T) {
	nilSlice := []string(nil)

	tests := []struct {
		name string
		in   []string
		out  []string
		// extra checks
		wantNil    bool
		wantNonNil bool
	}{
		{
			name:    "nil input -> nil output",
			in:      nilSlice,
			out:     nil,
			wantNil: true,
		},
		{
			name:       "empty slice -> empty slice (non-nil)",
			in:         []string{},
			out:        []string{},
			wantNonNil: true,
		},
		{
			name: "all unique already",
			in:   []string{"a", "b", "c"},
			out:  []string{"a", "b", "c"},
		},
		{
			name: "deduplicate preserves first-seen order",
			in:   []string{"b", "a", "b", "c", "a", "b"},
			out:  []string{"b", "a", "c"},
		},
		{
			name: "skip empty strings",
			in:   []string{"", "a", "", "a", "b", "", "b", ""},
			out:  []string{"a", "b"},
		},
		{
			name: "case-sensitive",
			in:   []string{"A", "a", "A", "a"},
			out:  []string{"A", "a"},
		},
		{
			name: "unicode handling",
			in:   []string{"văn", "văn", "Văn", "😀", "😀", "😃"},
			out:  []string{"văn", "Văn", "😀", "😃"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := slice.UniqueStrings(tc.in)

			if tc.wantNil && got != nil {
				t.Fatalf("expected nil slice, got %#v", got)
			}
			if tc.wantNonNil && got == nil {
				t.Fatalf("expected non-nil empty slice, got nil")
			}

			if !reflect.DeepEqual(got, tc.out) {
				t.Fatalf("want %v, got %v", tc.out, got)
			}
		})
	}
}

func TestContainsStr(t *testing.T) {
	tests := []struct {
		name string
		ss   []string
		want string
		out  bool
	}{
		{
			name: "found in middle",
			ss:   []string{"a", "b", "c"},
			want: "b",
			out:  true,
		},
		{
			name: "not found",
			ss:   []string{"a", "b", "c"},
			want: "x",
			out:  false,
		},
		{
			name: "empty slice",
			ss:   []string{},
			want: "a",
			out:  false,
		},
		{
			name: "nil slice",
			ss:   nil,
			want: "a",
			out:  false,
		},
		{
			name: "empty string present",
			ss:   []string{"", "a"},
			want: "",
			out:  true,
		},
		{
			name: "case-sensitive",
			ss:   []string{"A"},
			want: "a",
			out:  false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := slice.ContainsStr(tc.ss, tc.want)
			if got != tc.out {
				t.Fatalf("ContainsStr(%v,%q) = %v; want %v", tc.ss, tc.want, got, tc.out)
			}
		})
	}
}
