// pkg/mwx/requestid.go

package mw

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/chi07/go-svc-kit/httpx"
)

func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
			c.Set("X-Request-ID", id)
		}
		c.Locals("request_id", id)

		ctx := httpx.ContextWithRequestID(c.Context(), id)
		c.SetUserContext(ctx)

		return c.Next()
	}
}
