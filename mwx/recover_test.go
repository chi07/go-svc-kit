package mwx_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"github.com/chi07/go-svc-kit/mwx"
)

func newAppWithRecover(buf *bytes.Buffer) *fiber.App {
	log := zerolog.New(buf).With().Timestamp().Logger()
	app := fiber.New()
	app.Use(mwx.Recover(log))
	return app
}

func TestRecover_PanickingHandler_Returns500AndLogs(t *testing.T) {
	var buf bytes.Buffer
	app := newAppWithRecover(&buf)

	app.Get("/panic", func(c *fiber.Ctx) error {
		// Simulate RequestID middleware having run earlier
		c.Locals("request_id", "rid-xyz")
		panic("boom")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}

	if resp.StatusCode != 500 {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	errField, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error object in body, got: %#v", body["error"])
	}
	if msg, _ := errField["message"].(string); msg != "internal serverx error" {
		t.Fatalf("unexpected error message: %v", errField["message"])
	}
	if rid, _ := errField["requestId"].(string); rid != "rid-xyz" {
		t.Fatalf("unexpected requestId in body: %v", errField["requestId"])
	}

	logs := buf.String()
	// Ensure we logged at error level with expected fields
	wantSubs := []string{
		`"level":"error"`,
		`"message":"panic_recovered"`,
		`"panic":"boom"`,
		`"request_id":"rid-xyz"`,
		`"method":"GET"`,
		`"path":"/panic"`,
		// stack is logged as bytes; presence of "stack" key is enough
		`"stack":`,
	}
	for _, sub := range wantSubs {
		if !strings.Contains(logs, sub) {
			t.Fatalf("missing %q in logs: %s", sub, logs)
		}
	}
}

func TestRecover_NoPanic_PassesThrough(t *testing.T) {
	var buf bytes.Buffer
	app := newAppWithRecover(&buf)

	app.Get("/ok", func(c *fiber.Ctx) error {
		return c.SendString("fine")
	})

	req := httptest.NewRequest("GET", "/ok", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	logs := buf.String()
	if strings.Contains(logs, `"message":"panic_recovered"`) {
		t.Fatalf("did not expect panic_recovered in logs: %s", logs)
	}
}
