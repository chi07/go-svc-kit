package httpx_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/chi07/go-svc-kit/httpx"
)

func TestContextRequestID(t *testing.T) {
	ctx := context.Background()
	if got := httpx.RequestIDFromContext(ctx); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}

	ctx = httpx.ContextWithRequestID(ctx, "rid-123")
	if got := httpx.RequestIDFromContext(ctx); got != "rid-123" {
		t.Fatalf("want rid-123, got %q", got)
	}

	ctx2 := httpx.ContextWithRequestID(ctx, "")
	if got := httpx.RequestIDFromContext(ctx2); got != "rid-123" {
		t.Fatalf("empty override should keep old value, got %q", got)
	}
}

func TestNewBaseResty_DefaultTimeout(t *testing.T) {
	client := httpx.NewBaseResty(zerolog.Nop(), false, 0)
	if client.GetClient().Timeout != 3200*time.Millisecond {
		t.Fatalf("expected default timeout 3200ms, got %v", client.GetClient().Timeout)
	}
}

func TestNewBaseResty_BeforeRequest_AddsRequestIDHeader(t *testing.T) {
	var seenRID atomic.Value

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenRID.Store(r.Header.Get("X-Request-ID"))
		w.WriteHeader(200)
	}))
	defer s.Close()

	log := zerolog.Nop()
	c := httpx.NewBaseResty(log, false, 2*time.Second)

	ctx := httpx.ContextWithRequestID(context.Background(), "rid-xyz")
	resp, err := c.R().SetContext(ctx).Get(s.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode() != 200 {
		t.Fatalf("unexpected status: %d", resp.StatusCode())
	}

	rid, _ := seenRID.Load().(string)
	if rid != "rid-xyz" {
		t.Fatalf("X-Request-ID not propagated, got %q", rid)
	}
}

func TestNewBaseResty_RetryOn5xxOnly(t *testing.T) {
	var hits int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&hits, 1)
		if n == 1 {
			http.Error(w, "boom", http.StatusInternalServerError) // 500
			return
		}
		w.WriteHeader(200)
	}))
	defer s.Close()

	c := httpx.NewBaseResty(zerolog.Nop(), false, 2*time.Second)

	resp, err := c.R().Get(s.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode() != 200 {
		t.Fatalf("expected 200 after retry, got %d", resp.StatusCode())
	}
	// Resty increments Attempt; on success after one retry it should be 2
	if resp.Request.Attempt != 2 {
		t.Fatalf("expected Attempt=2, got %d", resp.Request.Attempt)
	}
}

func TestNewBaseResty_NoRetryOn4xx(t *testing.T) {
	var hits int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		http.Error(w, "bad", http.StatusBadRequest) // 400
	}))
	defer s.Close()

	c := httpx.NewBaseResty(zerolog.Nop(), false, 2*time.Second)

	resp, err := c.R().Get(s.URL)
	if err != nil {
	}
	if resp.Request.Attempt != 1 {
		t.Fatalf("expected no retry on 4xx, Attempt=1, got %d", resp.Request.Attempt)
	}
}

func TestNewBaseResty_NoRetryOnContextCanceled(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		w.WriteHeader(200)
	}))
	defer s.Close()

	c := httpx.NewBaseResty(zerolog.Nop(), false, 2*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.R().SetContext(ctx).Get(s.URL)
	if err == nil {
		t.Fatalf("expected context canceled error")
	}

	if !strings.Contains(err.Error(), "context canceled") &&
		!strings.Contains(err.Error(), "context canceled") { // keep simple
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewBaseResty_DebugMode_DoesNotBreakFlow(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer s.Close()

	log := zerolog.New(nil)
	c := httpx.NewBaseResty(log, true, time.Second)

	resp, err := c.R().Get(s.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode() != 200 {
		t.Fatalf("unexpected status: %d", resp.StatusCode())
	}
}
