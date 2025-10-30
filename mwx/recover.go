// pkg/mwx/recover.go

package mw

import (
	"runtime/debug"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

func Recover(log zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				reqID, _ := c.Locals("request_id").(string)
				log.Error().
					Interface("panic", r).
					Str("request_id", reqID).
					Str("method", c.Method()).
					Str("path", c.OriginalURL()).
					Str("ip", c.IP()).
					Bytes("stack", debug.Stack()).
					Msg("panic_recovered")
				_ = c.Status(500).JSON(fiber.Map{"error": fiber.Map{"message": "internal serverx error", "requestId": reqID}})
			}
		}()
		return c.Next()
	}
}
