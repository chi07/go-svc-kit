package mwx

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

func AccessLogger(log zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		statusCode := c.Response().StatusCode()
		reqID, _ := c.Locals("request_id").(string)
		evt := log.Info()
		if statusCode >= 500 {
			evt = log.Error()
		} else if statusCode >= 400 {
			evt = log.Warn()
		}

		evt.
			Str("event", "http_access").
			Str("request_id", reqID).
			Str("method", c.Method()).
			Str("path", c.OriginalURL()).
			Int("status", statusCode).
			Dur("latency", time.Since(start)).
			Int("resp_bytes", c.Response().Header.ContentLength()).
			Str("ip", c.IP()).
			Str("user_agent", string(c.Request().Header.UserAgent())).
			Err(err).
			Msg("access")
		return err
	}
}
