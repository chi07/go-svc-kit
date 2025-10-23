package pagex

type PageInfo struct {
	Limit       int64 `json:"limit"`
	Offset      int64 `json:"offset"`
	Total       int64 `json:"total"`
	TotalPages  int64 `json:"totalPages"`
	CurrentPage int64 `json:"currentPage"`
	HasNext     bool  `json:"hasNext"`
	HasPrevious bool  `json:"hasPrevious"`
}

func (p *PageInfo) Compute() {
	if p.Limit <= 0 {
		p.Limit = 10
	}
	if p.Total <= 0 {
		p.TotalPages = 1
		p.CurrentPage = 1
		p.HasNext = false
		p.HasPrevious = p.Offset > 0
		return
	}
	p.TotalPages = (p.Total + p.Limit - 1) / p.Limit
	p.CurrentPage = (p.Offset / p.Limit) + 1
	if p.CurrentPage > p.TotalPages {
		p.CurrentPage = p.TotalPages
	}
	if p.CurrentPage < 1 {
		p.CurrentPage = 1
	}
	p.HasPrevious = p.CurrentPage > 1
	p.HasNext = p.CurrentPage < p.TotalPages
}

func PageToOffset(page, limit int64) int64 {
	if limit <= 0 || page <= 1 {
		return 0
	}
	return (page - 1) * limit
}
