package server

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

func Listen(app *fiber.App, ipVersion, port string, log zerolog.Logger) error {
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

	go func() {
		if e := app.Listener(ln); e != nil {
			log.Error().Err(e).Msg("fiber_listener_error")
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	return app.Shutdown()
}
