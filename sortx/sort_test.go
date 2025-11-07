package sortx_test

import (
	"reflect"
	"testing"

	"github.com/chi07/go-svc-kit/sortx"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []sortx.SortField
	}{
		{
			name: "empty string -> nil",
			raw:  "",
			want: nil,
		},
		{
			name: "only spaces -> nil",
			raw:  "    ",
			want: nil,
		},
		{
			name: "single field asc (implicit)",
			raw:  "name",
			want: []sortx.SortField{{Field: "name", Desc: false}},
		},
		{
			name: "single field asc (explicit asc ignored)",
			raw:  "name:asc",
			want: []sortx.SortField{{Field: "name", Desc: false}},
		},
		{
			name: "single field desc (desc)",
			raw:  "name:desc",
			want: []sortx.SortField{{Field: "name", Desc: true}},
		},
		{
			name: "single field desc (d)",
			raw:  "name:d",
			want: []sortx.SortField{{Field: "name", Desc: true}},
		},
		{
			name: "single field desc (descending)",
			raw:  "name:descending",
			want: []sortx.SortField{{Field: "name", Desc: true}},
		},
		{
			name: "unknown direction -> treat as asc",
			raw:  "name:zzz",
			want: []sortx.SortField{{Field: "name", Desc: false}},
		},
		{
			name: "name with trailing colon -> asc",
			raw:  "name:",
			want: []sortx.SortField{{Field: "name", Desc: false}},
		},
		{
			name: "name with whitespace dir -> asc",
			raw:  "name:   ",
			want: []sortx.SortField{{Field: "name", Desc: false}},
		},
		{
			name: "multiple fields with mix and spaces",
			raw:  "  email:desc ,  name:asc ,  id  ",
			want: []sortx.SortField{
				{Field: "email", Desc: true},
				{Field: "name", Desc: false},
				{Field: "id", Desc: false},
			},
		},
		{
			name: "extra commas and empties are skipped",
			raw:  ",,, a:desc , , b ,, :desc , , c:asc , ,",
			want: []sortx.SortField{
				{Field: "a", Desc: true},
				{Field: "b", Desc: false},
				{Field: "c", Desc: false},
			},
		},
		{
			name: "case-insensitive direction",
			raw:  "x:DeSc, y:D, z:DESCENDING",
			want: []sortx.SortField{
				{Field: "x", Desc: true},
				{Field: "y", Desc: true},
				{Field: "z", Desc: true},
			},
		},
		{
			name: "unicode field names preserved",
			raw:  "tên:desc, tuổi:asc",
			want: []sortx.SortField{
				{Field: "tên", Desc: true},
				{Field: "tuổi", Desc: false},
			},
		},
		{
			name: "empty field before colon is ignored",
			raw:  ":desc,  :d ,  :descending , valid:desc",
			want: []sortx.SortField{
				{Field: "valid", Desc: true},
			},
		},
		{
			name: "order preserved and duplicates allowed",
			raw:  "a:desc,a, b:desc, a:descending",
			want: []sortx.SortField{
				{Field: "a", Desc: true},
				{Field: "a", Desc: false},
				{Field: "b", Desc: true},
				{Field: "a", Desc: true},
			},
		},
		{
			name: "trim spaces around field and dir",
			raw:  "  field  :   desc  ",
			want: []sortx.SortField{{Field: "field", Desc: true}},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := sortx.Parse(tc.raw)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Parse(%q) = %#v; want %#v", tc.raw, got, tc.want)
			}
		})
	}
}
