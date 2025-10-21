// pkg/mw/accesslog.go

package mw

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

func AccessLogger(log zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		reqID, _ := c.Locals("request_id").(string)
		evt := log.Info()
		if c.Response().StatusCode() >= 500 {
			evt = log.Error()
		} else if c.Response().StatusCode() >= 400 {
			evt = log.Warn()
		}

		evt.
			Str("event", "http_access").
			Str("request_id", reqID).
			Str("method", c.Method()).
			Str("path", c.OriginalURL()).
			Int("status", c.Response().StatusCode()).
			Dur("latency", time.Since(start)).
			Int("resp_bytes", len(c.Response().Body())).
			Str("ip", c.IP()).
			Str("user_agent", string(c.Request().Header.UserAgent())).
			Err(err).
			Msg("access")
		return err
	}
}
