package mwx_test

import (
	"bytes"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/chi07/go-svc-kit/mwx"
)

func TestCORS_SpecificOrigins_EchoesMatchingOrigin(t *testing.T) {
	app := fiber.New()
	app.Use(mwx.CORS("https://a.com, https://b.com"))
	app.Get("/x", func(c *fiber.Ctx) error { return c.SendString("ok") })

	// Matching origin
	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("Origin", "https://a.com")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if ao := resp.Header.Get("Access-Control-Allow-Origin"); ao != "https://a.com" {
		t.Fatalf("expected ACAO to echo https://a.com, got %q", ao)
	}

	// Non-matching origin should result in no ACAO
	req2 := httptest.NewRequest("GET", "/x", nil)
	req2.Header.Set("Origin", "https://c.com")
	resp2, err := app.Test(req2, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if ao := resp2.Header.Get("Access-Control-Allow-Origin"); ao != "" {
		t.Fatalf("expected empty ACAO for disallowed origin, got %q", ao)
	}
}

func TestGzip_CompressesWhenAccepted(t *testing.T) {
	app := fiber.New()
	app.Use(mwx.Gzip())
	app.Get("/big", func(c *fiber.Ctx) error {
		var b bytes.Buffer
		for i := 0; i < 5000; i++ { // big enough to trigger compression
			b.WriteString("0123456789")
		}
		return c.Send(b.Bytes())
	})

	// With gzip accepted
	req := httptest.NewRequest("GET", "/big", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if ce := resp.Header.Get("Content-Encoding"); ce != "gzip" {
		t.Fatalf("expected Content-Encoding=gzip, got %q", ce)
	}
	zipped, _ := io.ReadAll(resp.Body)
	if len(zipped) == 0 {
		t.Fatalf("expected non-empty compressed body")
	}

	// Without gzip accepted -> no compression
	req2 := httptest.NewRequest("GET", "/big", nil)
	resp2, err := app.Test(req2, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if ce := resp2.Header.Get("Content-Encoding"); ce != "" {
		t.Fatalf("expected no Content-Encoding, got %q", ce)
	}
	plain, _ := io.ReadAll(resp2.Body)
	if len(plain) == 0 {
		t.Fatalf("expected non-empty uncompressed body")
	}
}
