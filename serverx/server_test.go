package server_test

import (
	"bytes"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"github.com/chi07/go-svc-kit/serverx"
)

func newLogger() zerolog.Logger {
	var buf bytes.Buffer
	return zerolog.New(&buf).With().Timestamp().Logger()
}

func TestListen_IPv4_OK(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })

	log := newLogger()

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Listen(app, "v4", "0", log)
	}()

	time.Sleep(100 * time.Millisecond)

	if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
		t.Fatalf("failed to signal self: %v", err)
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Listen returned error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for Listen to return after SIGTERM")
	}
}

func TestListen_InvalidPort_ReturnsError(t *testing.T) {
	app := fiber.New()
	log := newLogger()

	if err := server.Listen(app, "v4", "not-a-port", log); err == nil {
		t.Fatal("expected error for invalid port, got nil")
	}
}

func TestListen_IPv6_Smoke(t *testing.T) {
	app := fiber.New()
	log := newLogger()

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Listen(app, "ipv6", "0", log)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			t.Logf("IPv6 not available or bind failed quickly: %v (acceptable)", err)
			return
		}
	default:
		time.Sleep(100 * time.Millisecond)
		if err := syscall.Kill(os.Getpid(), syscall.SIGINT); err != nil {
			t.Fatalf("failed to signal self: %v", err)
		}
		select {
		case err := <-errCh:
			if err != nil {
				t.Fatalf("Listen (ipv6) returned error: %v", err)
			}
		case <-time.After(3 * time.Second):
			t.Fatal("timeout waiting for IPv6 Listen to return after SIGINT")
		}
	}
}
