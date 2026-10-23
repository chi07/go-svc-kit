package server

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

// DefaultShutdownTimeout bounds graceful shutdown on SIGINT/SIGTERM.
const DefaultShutdownTimeout = 10 * time.Second

// Listen serves app on the given port and blocks until SIGINT/SIGTERM,
// then shuts down gracefully within DefaultShutdownTimeout.
func Listen(app *fiber.App, ipVersion, port string, log zerolog.Logger) error {
	return ListenWithTimeout(app, ipVersion, port, log, DefaultShutdownTimeout)
}

// ListenWithTimeout is Listen with a caller-controlled graceful shutdown
// deadline.
func ListenWithTimeout(app *fiber.App, ipVersion, port string, log zerolog.Logger, shutdownTimeout time.Duration) error {
	host := "0.0.0.0"
	netw := "tcp"
	if ipVersion == "v6" || ipVersion == "ipv6" {
		host = "::"
		netw = "tcp6"
	}
	addr := net.JoinHostPort(host, port)

	ln, err := net.Listen(netw, addr)
	if err != nil {
		return err
	}
	log.Info().Str("addr", addr).Str("net", netw).Msg("server_listen")

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- app.Listener(ln)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case e := <-serveErr:
		return e // listener died before any signal
	case <-quit:
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return app.ShutdownWithContext(ctx)
}
