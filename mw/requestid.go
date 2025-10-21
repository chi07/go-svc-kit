// pkg/mw/requestid.go

package mw

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
			c.Set("X-Request-ID", id)
		}
		c.Locals("request_id", id)
		return c.Next()
	}
}
