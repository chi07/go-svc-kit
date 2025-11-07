package mwx_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/chi07/go-svc-kit/httpx"
	"github.com/chi07/go-svc-kit/mwx"
)

func newApp() *fiber.App {
	app := fiber.New()
	app.Use(mwx.RequestID())
	app.Get("/check", func(c *fiber.Ctx) error {
		loc, _ := c.Locals("request_id").(string)
		ctxRID := httpx.RequestIDFromContext(c.UserContext())
		return c.JSON(map[string]string{
			"header": c.GetRespHeader("X-Request-ID"),
			"locals": loc,
			"ctx":    ctxRID,
		})
	})
	return app
}

func TestRequestID_GeneratesWhenMissing(t *testing.T) {
	app := newApp()

	req := httptest.NewRequest("GET", "/check", nil)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body error: %v", err)
	}

	rh := body["header"]
	rl := body["locals"]
	rc := body["ctx"]

	if rh == "" || rl == "" || rc == "" {
		t.Fatalf("expected all request id fields to be non-empty, got header=%q locals=%q ctx=%q", rh, rl, rc)
	}

	if _, err := uuid.Parse(rh); err != nil {
		t.Fatalf("generated X-Request-ID is not a valid UUID: %q (%v)", rh, err)
	}

	if !(rh == rl && rl == rc) {
		t.Fatalf("request id mismatch: header=%q locals=%q ctx=%q", rh, rl, rc)
	}
}

func TestRequestID_UsesProvidedHeader(t *testing.T) {
	app := newApp()

	const given = "req-12345-fixed"
	req := httptest.NewRequest("GET", "/check", nil)
	req.Header.Set("X-Request-ID", given)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body error: %v", err)
	}

	rh := body["header"]
	rl := body["locals"]
	rc := body["ctx"]

	if rl != given || rc != given {
		t.Fatalf("expected locals and ctx to be %q, got locals=%q ctx=%q", given, rl, rc)
	}

	if rh != "" && rh != given {
		t.Fatalf("if response X-Request-ID is set, it must equal %q; got %q", given, rh)
	}
}
