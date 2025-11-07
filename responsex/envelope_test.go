package responsex_test

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/chi07/go-svc-kit/responsex"
)

func TestNewEnvelopeAndPaginator(t *testing.T) {
	p := responsex.NewPaginator(3, 20, 95)
	env := responsex.NewEnvelope([]int{1, 2, 3}, p)

	if env.Paginator == nil {
		t.Fatalf("expected paginator to be set")
	}
	if env.Paginator.CurrentPage != 3 || env.Paginator.Limit != 20 || env.Paginator.Total != 95 {
		t.Fatalf("unexpected paginator: %+v", env.Paginator)
	}
	if len(env.Data) != 3 || env.Data[0] != 1 || env.Data[2] != 3 {
		t.Fatalf("unexpected data: %#v", env.Data)
	}
}

func TestNewErrorEnvelope(t *testing.T) {
	env := responsex.NewErrorEnvelope(map[string]any{"msg": "boom"})
	// Data is not tagged with omitempty, so it should be the zero value (nil slice).
	b, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	s := string(b)
	if !strings.Contains(s, `"error":{"msg":"boom"}`) {
		t.Fatalf("expected error payload, got: %s", s)
	}
	if !strings.Contains(s, `"data":null`) {
		t.Fatalf("expected data to be null (zero slice), got: %s", s)
	}
}

func TestFiberWriteJSON_WithPaginator(t *testing.T) {
	app := fiber.New()
	app.Get("/ok", func(c *fiber.Ctx) error {
		p := &responsex.Paginator{
			Limit: 10, Offset: 20, Total: 77, TotalPages: 8, CurrentPage: 3, HasNext: true, HasPrevious: true,
		}
		return responsex.FiberWriteJSON(c, 200, []string{"a", "b"}, p)
	})

	req := httptest.NewRequest("GET", "/ok", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("fiber app.Test error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	// data
	data, ok := body["data"].([]any)
	if !ok || len(data) != 2 || data[0].(string) != "a" || data[1].(string) != "b" {
		t.Fatalf("unexpected data: %#v", body["data"])
	}
	// paginator present with expected fields
	pg, ok := body["paginator"].(map[string]any)
	if !ok {
		t.Fatalf("paginator missing")
	}
	if int(pg["limit"].(float64)) != 10 || int(pg["offset"].(float64)) != 20 ||
		int(pg["total"].(float64)) != 77 || int(pg["totalPages"].(float64)) != 8 ||
		int(pg["currentPage"].(float64)) != 3 || pg["hasNext"] != true || pg["hasPrevious"] != true {
		t.Fatalf("unexpected paginator: %#v", pg)
	}
	// error should be absent
	if _, exists := body["error"]; exists {
		t.Fatalf("did not expect error field, got: %#v", body["error"])
	}
}

func TestFiberWriteJSON_WithoutPaginator_OmitsPaginatorField(t *testing.T) {
	app := fiber.New()
	app.Get("/nop", func(c *fiber.Ctx) error {
		return responsex.FiberWriteJSON(c, 201, []int{1, 2}, nil)
	})
	req := httptest.NewRequest("GET", "/nop", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("fiber app.Test error: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if _, ok := body["paginator"]; ok {
		t.Fatalf("paginator should be omitted when nil, got: %#v", body["paginator"])
	}
	if _, ok := body["error"]; ok {
		t.Fatalf("unexpected error field: %#v", body["error"])
	}
}

func TestFiberWriteError(t *testing.T) {
	app := fiber.New()
	app.Get("/err", func(c *fiber.Ctx) error {
		return responsex.FiberWriteError(c, 400, map[string]any{"code": "BAD", "msg": "bad request"})
	})
	req := httptest.NewRequest("GET", "/err", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("fiber app.Test error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	// error present
	er, ok := body["error"].(map[string]any)
	if !ok || er["code"] != "BAD" || er["msg"] != "bad request" {
		t.Fatalf("unexpected error body: %#v", body["error"])
	}
	// data should be null (zero value of []T)
	if v, ok := body["data"]; !ok || v != nil {
		t.Fatalf("expected data to be null, got: %#v", v)
	}
	// paginator should be omitted
	if _, ok := body["paginator"]; ok {
		t.Fatalf("unexpected paginator field: %#v", body["paginator"])
	}
}
