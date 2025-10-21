package logx

import (
	"os"

	"github.com/rs/zerolog"
)

func New(env string) zerolog.Logger {
	if env == "prod" || env == "staging" {
		return zerolog.New(os.Stdout).With().Timestamp().Logger()
	}
	// for dev: make beautiful console output
	cw := zerolog.NewConsoleWriter()
	return zerolog.New(cw).With().Timestamp().Logger()
}
