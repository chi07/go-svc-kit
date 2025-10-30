package responsex

import (
	"github.com/gofiber/fiber/v2"
)

type DataEnvelope[T any] struct {
	Data      []T        `json:"data"`
	Paginator *Paginator `json:"paginator,omitempty"`
	Error     any        `json:"error,omitempty"`
}

type Paginator struct {
	Limit       int64 `json:"limit"`
	Offset      int64 `json:"offset"`
	Total       int64 `json:"total"`
	TotalPages  int64 `json:"totalPages"`
	CurrentPage int64 `json:"currentPage"`
	HasNext     bool  `json:"hasNext"`
	HasPrevious bool  `json:"hasPrevious"`
}

func NewEnvelope[T any](data []T, p *Paginator) DataEnvelope[T] {
	return DataEnvelope[T]{Data: data, Paginator: p}
}

func NewErrorEnvelope(err any) DataEnvelope[any] {
	return DataEnvelope[any]{Error: err}
}

func NewPaginator(page, limit, total int64) *Paginator {
	return &Paginator{CurrentPage: page, Limit: limit, Total: total}
}

func FiberWriteJSON[T any](c *fiber.Ctx, status int, data []T, p *Paginator) error {
	return c.Status(status).JSON(NewEnvelope(data, p))
}

func FiberWriteError(c *fiber.Ctx, status int, err any) error {
	return c.Status(status).JSON(NewErrorEnvelope(err))
}
