package pagex

import (
	"slices"

	"github.com/chi07/go-svc-kit/responsex"
)

const (
	DefaultLimit int64 = 10
	MinLimit     int64 = 1
	MaxLimit     int64 = 200
)

func clampNonNeg(x int64) int64 {
	if x < 0 {
		return 0
	}
	return x
}

func ceilDiv(a, b int64) int64 {
	if b <= 0 {
		return 0
	}
	return (a + b - 1) / b
}

func ClampLimit(limit int64) int64 {
	switch {
	case limit < MinLimit:
		return DefaultLimit
	case limit > MaxLimit:
		return MaxLimit
	default:
		return limit
	}
}

type PageInfo struct {
	Limit       int64 `json:"limit"`
	Offset      int64 `json:"offset"`
	Total       int64 `json:"total"`
	TotalPages  int64 `json:"totalPages"`
	CurrentPage int64 `json:"currentPage"`
	HasNext     bool  `json:"hasNext"`
	HasPrevious bool  `json:"hasPrevious"`
}

func PageToOffset(page, limit int64) int64 {
	if limit <= 0 || page <= 1 {
		return 0
	}
	return (page - 1) * limit
}

func Normalize(limit, offset int64) (int64, int64) {
	return clampNonNeg(limit), clampNonNeg(offset)
}

func FromTotal(limit, offset, total int64) PageInfo {
	limit, offset = Normalize(limit, offset)
	if limit <= 0 {
		limit = DefaultLimit
	}
	pi := PageInfo{
		Limit:  limit,
		Offset: offset,
		Total:  total,
	}

	if total <= 0 {
		pi.TotalPages = 1
		pi.CurrentPage = 1
		pi.HasPrevious = offset > 0
		pi.HasNext = false
		return pi
	}

	pi.TotalPages = ceilDiv(total, limit)
	pi.CurrentPage = (offset / limit) + 1
	if pi.CurrentPage < 1 {
		pi.CurrentPage = 1
	} else if pi.CurrentPage > pi.TotalPages {
		pi.CurrentPage = pi.TotalPages
	}

	pi.HasPrevious = offset > 0
	pi.HasNext = (offset + limit) < total
	return pi
}

func FromLookahead(limit, offset, rowsLen int64) PageInfo {
	limit, offset = Normalize(limit, offset)
	if limit <= 0 {
		limit = DefaultLimit
	}
	pi := PageInfo{
		Limit:       limit,
		Offset:      offset,
		HasPrevious: offset > 0,
		Total:       0,
		TotalPages:  1,
	}

	pi.CurrentPage = (offset / limit) + 1
	pi.HasNext = rowsLen > limit
	return pi
}

func (p *PageInfo) Compute() {
	if p.Limit <= 0 {
		p.Limit = DefaultLimit
	}
	p.Offset = clampNonNeg(p.Offset)

	if p.Total <= 0 {
		p.TotalPages = 1
		p.CurrentPage = 1
		p.HasPrevious = p.Offset > 0
		p.HasNext = false
		return
	}

	p.TotalPages = ceilDiv(p.Total, p.Limit)
	p.CurrentPage = (p.Offset / p.Limit) + 1
	if p.CurrentPage < 1 {
		p.CurrentPage = 1
	} else if p.CurrentPage > p.TotalPages {
		p.CurrentPage = p.TotalPages
	}

	p.HasPrevious = p.Offset > 0
	p.HasNext = (p.Offset + p.Limit) < p.Total
}

func FromPageLimitTotal(page, limit, total int64) PageInfo {
	limit = ClampLimit(limit)
	offset := PageToOffset(page, limit)
	return FromTotal(limit, offset, total)
}

func FromPageLimitLookahead(page, limit, fetchedRows int64) PageInfo {
	limit = ClampLimit(limit)
	offset := PageToOffset(page, limit)
	return FromLookahead(limit, offset, fetchedRows)
}

func TrimLookahead[T any](rows []T, limit int64) ([]T, bool) {
	if limit <= 0 {
		return rows, false
	}
	if int64(len(rows)) > limit {
		return rows[:limit], true
	}
	return rows, false
}

func SlicePage[T any](rows []T, offset, limit int64) []T {
	offset = clampNonNeg(offset)
	limit = ClampLimit(limit)
	if offset >= int64(len(rows)) {
		return nil
	}
	end := offset + limit
	if end > int64(len(rows)) {
		end = int64(len(rows))
	}
	return slices.Clone(rows[offset:end])
}

func (p PageInfo) ToPaginator() *responsex.Paginator {
	return &responsex.Paginator{
		Limit:       p.Limit,
		Offset:      p.Offset,
		Total:       p.Total,
		TotalPages:  p.TotalPages,
		CurrentPage: p.CurrentPage,
		HasNext:     p.HasNext,
		HasPrevious: p.HasPrevious,
	}
}
