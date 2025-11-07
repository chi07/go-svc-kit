package pagex_test

import (
	"reflect"
	"testing"

	"github.com/chi07/go-svc-kit/pagex"
	"github.com/chi07/go-svc-kit/responsex"
)

func TestClampLimit(t *testing.T) {
	tests := []struct {
		name string
		in   int64
		want int64
	}{
		{"below min -> default", pagex.MinLimit - 1, pagex.DefaultLimit},
		{"zero -> default", 0, pagex.DefaultLimit},
		{"equal min", pagex.MinLimit, pagex.MinLimit},
		{"within range", 25, 25},
		{"above max -> max", pagex.MaxLimit + 1, pagex.MaxLimit},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := pagex.ClampLimit(tc.in); got != tc.want {
				t.Fatalf("ClampLimit(%d) = %d; want %d", tc.in, got, tc.want)
			}
		})
	}
}

func TestPageToOffset(t *testing.T) {
	tests := []struct {
		name string
		page int64
		lim  int64
		want int64
	}{
		{"page 1 -> 0", 1, 10, 0},
		{"page 2 -> lim", 2, 10, 10},
		{"page 5 -> 40", 5, 10, 40},
		{"limit<=0 -> 0", 3, 0, 0},
		{"page<=1 -> 0", 0, 10, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := pagex.PageToOffset(tc.page, tc.lim); got != tc.want {
				t.Fatalf("PageToOffset(%d,%d) = %d; want %d", tc.page, tc.lim, got, tc.want)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	lim, off := pagex.Normalize(-5, -10)
	if lim != 0 || off != 0 {
		t.Fatalf("Normalize(-5,-10) = (%d,%d); want (0,0)", lim, off)
	}
	lim2, off2 := pagex.Normalize(10, 20)
	if lim2 != 10 || off2 != 20 {
		t.Fatalf("Normalize(10,20) = (%d,%d); want (10,20)", lim2, off2)
	}
}

func TestFromTotal(t *testing.T) {
	t.Run("total<=0 -> single page, page=1, no next", func(t *testing.T) {
		pi := pagex.FromTotal(10, 0, 0)
		if pi.TotalPages != 1 || pi.CurrentPage != 1 || pi.HasNext || pi.Total != 0 {
			t.Fatalf("unexpected: %+v", pi)
		}
	})
	t.Run("normal pagination", func(t *testing.T) {
		pi := pagex.FromTotal(10, 20, 95) // page=3
		if pi.TotalPages != 10 || pi.CurrentPage != 3 || !pi.HasPrevious || pi.HasNext == false {
			t.Fatalf("unexpected: %+v", pi)
		}
	})
	t.Run("offset beyond total -> clamp current to totalPages", func(t *testing.T) {
		pi := pagex.FromTotal(10, 1000, 50) // totalPages=5
		if pi.TotalPages != 5 || pi.CurrentPage != 5 {
			t.Fatalf("expected current=5,totalPages=5; got %+v", pi)
		}
	})
	t.Run("limit<=0 -> defaulted", func(t *testing.T) {
		pi := pagex.FromTotal(0, 0, 25)
		if pi.Limit != pagex.DefaultLimit {
			t.Fatalf("limit should default; got %+v", pi)
		}
	})
}

func TestFromLookahead(t *testing.T) {
	t.Run("rowsLen>limit -> HasNext true", func(t *testing.T) {
		pi := pagex.FromLookahead(10, 20, 11)
		if !pi.HasNext || !pi.HasPrevious || pi.CurrentPage != 3 {
			t.Fatalf("unexpected: %+v", pi)
		}
	})
	t.Run("limit<=0 -> defaulted", func(t *testing.T) {
		pi := pagex.FromLookahead(0, 0, 1)
		if pi.Limit != pagex.DefaultLimit {
			t.Fatalf("limit should default; got %+v", pi)
		}
	})
}

func TestPageInfo_Compute(t *testing.T) {
	t.Run("total<=0 branch", func(t *testing.T) {
		p := pagex.PageInfo{Limit: 0, Offset: 5, Total: 0}
		p.Compute()
		if p.Limit != pagex.DefaultLimit || p.TotalPages != 1 || p.CurrentPage != 1 || !p.HasPrevious || p.HasNext {
			t.Fatalf("unexpected: %+v", p)
		}
	})
	t.Run("normal branch", func(t *testing.T) {
		p := pagex.PageInfo{Limit: 10, Offset: 30, Total: 35}
		p.Compute()
		if p.TotalPages != 4 || p.CurrentPage != 4 || !p.HasPrevious || p.HasNext {
			t.Fatalf("unexpected: %+v", p)
		}
	})
	t.Run("clamp current page to bounds", func(t *testing.T) {
		p := pagex.PageInfo{Limit: 10, Offset: 10, Total: 5} // totalPages=1
		p.Compute()
		if p.CurrentPage != 1 {
			t.Fatalf("expected current=1; got %+v", p)
		}
	})
}

func TestFromPageLimitTotal(t *testing.T) {
	pi := pagex.FromPageLimitTotal(3, 10, 95) // offset=20
	if pi.Offset != 20 || pi.CurrentPage != 3 || pi.TotalPages != 10 {
		t.Fatalf("unexpected: %+v", pi)
	}
	// limit clamped
	pi2 := pagex.FromPageLimitTotal(1, pagex.MaxLimit+999, 5)
	if pi2.Limit != pagex.MaxLimit {
		t.Fatalf("limit should be clamped to MaxLimit; got %+v", pi2)
	}
}

func TestFromPageLimitLookahead(t *testing.T) {
	pi := pagex.FromPageLimitLookahead(2, 10, 15) // offset=10
	if pi.Offset != 10 || !pi.HasNext {
		t.Fatalf("unexpected: %+v", pi)
	}
}

func TestTrimLookahead(t *testing.T) {
	rows := []int{1, 2, 3, 4}
	trimmed, hasNext := pagex.TrimLookahead(rows, 3)
	if !hasNext || len(trimmed) != 3 || trimmed[2] != 3 {
		t.Fatalf("unexpected: hasNext=%v, rows=%v", hasNext, trimmed)
	}
	trimmed2, hasNext2 := pagex.TrimLookahead(rows, 5)
	if hasNext2 || len(trimmed2) != 4 {
		t.Fatalf("unexpected: hasNext=%v, rows=%v", hasNext2, trimmed2)
	}
}

func TestSlicePage(t *testing.T) {
	rows := []int{1, 2, 3, 4, 5}

	// middle slice
	got := pagex.SlicePage(rows, 1, 2) // expect [2,3]
	want := []int{2, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SlicePage middle: want %v, got %v", want, got)
	}
	// end clamp
	got2 := pagex.SlicePage(rows, 3, 10) // expect [4,5]
	want2 := []int{4, 5}
	if !reflect.DeepEqual(got2, want2) {
		t.Fatalf("SlicePage end clamp: want %v, got %v", want2, got2)
	}
	// offset >= len -> nil
	got3 := pagex.SlicePage(rows, 10, 2)
	if got3 != nil {
		t.Fatalf("SlicePage beyond end: want nil, got %v", got3)
	}
	// returned slice is a clone (modifying result must not change original)
	got4 := pagex.SlicePage(rows, 1, 2)
	got4[0] = 999
	if rows[1] != 2 {
		t.Fatalf("expected underlying rows unaffected by mutation, rows=%v", rows)
	}
}

func TestToPaginator(t *testing.T) {
	p := pagex.PageInfo{
		Limit: 7, Offset: 14, Total: 100, TotalPages: 15, CurrentPage: 3, HasNext: true, HasPrevious: true,
	}
	pg := p.ToPaginator()
	want := &responsex.Paginator{
		Limit: 7, Offset: 14, Total: 100, TotalPages: 15, CurrentPage: 3, HasNext: true, HasPrevious: true,
	}
	if !reflect.DeepEqual(pg, want) {
		t.Fatalf("ToPaginator mismatch: want %+v, got %+v", want, pg)
	}
}
