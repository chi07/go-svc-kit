package mwx_test

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chi07/go-svc-kit/mwx"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

func newAppWithLogger(buf *bytes.Buffer) (*fiber.App, zerolog.Logger) {
	log := zerolog.New(buf).With().Timestamp().Logger()
	app := fiber.New()
	app.Use(mwx.AccessLogger(log))
	return app, log
}

func TestAccessLogger_InfoOn2xx(t *testing.T) {
	var buf bytes.Buffer
	app, _ := newAppWithLogger(&buf)

	app.Get("/ok", func(c *fiber.Ctx) error {
		c.Locals("request_id", "rid-200")
		return c.SendString("hello")
	})

	req := httptest.NewRequest("GET", "/ok", nil)
	_, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}

	log := buf.String()
	wantSubs := []string{
		`"level":"info"`,
		`"event":"http_access"`,
		`"method":"GET"`,
		`"path":"/ok"`,
		`"status":200`,
		`"request_id":"rid-200"`,
		`"message":"access"`,
	}
	for _, sub := range wantSubs {
		if !strings.Contains(log, sub) {
			t.Fatalf("missing %q in log: %s", sub, log)
		}
	}
	if strings.Contains(log, `"error":`) {
		t.Fatalf("did not expect error field in info log: %s", log)
	}
}

func TestAccessLogger_WarnOn4xx_IncludesError(t *testing.T) {
	var buf bytes.Buffer
	app, _ := newAppWithLogger(&buf)

	app.Get("/bad", func(c *fiber.Ctx) error {
		c.Locals("request_id", "rid-400")
		// Explicitly set 400 so middleware sees it
		_ = c.SendStatus(400)
		return fiber.ErrBadRequest
	})

	req := httptest.NewRequest("GET", "/bad", nil)
	_, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}

	log := buf.String()
	wantSubs := []string{
		`"level":"warn"`,
		`"status":400`,
		`"path":"/bad"`,
		`"request_id":"rid-400"`,
		`"message":"access"`,
		`"error":"Bad Request"`,
	}
	for _, sub := range wantSubs {
		if !strings.Contains(log, sub) {
			t.Fatalf("missing %q in log: %s", sub, log)
		}
	}
}

func TestAccessLogger_ErrorOn5xx_IncludesError(t *testing.T) {
	var buf bytes.Buffer
	app, _ := newAppWithLogger(&buf)

	app.Get("/boom", func(c *fiber.Ctx) error {
		c.Locals("request_id", "rid-500")
		// Explicitly set 500 so middleware sees it
		_ = c.SendStatus(500)
		return fiber.ErrInternalServerError
	})

	req := httptest.NewRequest("GET", "/boom", nil)
	_, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}

	log := buf.String()
	wantSubs := []string{
		`"level":"error"`,
		`"status":500`,
		`"path":"/boom"`,
		`"request_id":"rid-500"`,
		`"message":"access"`,
		`"error":"Internal Server Error"`,
	}
	for _, sub := range wantSubs {
		if !strings.Contains(log, sub) {
			t.Fatalf("missing %q in log: %s", sub, log)
		}
	}
}
