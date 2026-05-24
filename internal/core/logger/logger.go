package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	LogLevelDebug = 0
	LogLevelInfo  = 1
	LogLevelWarn  = 2
	LogLevelError = 3
)

func Init(level int, format string) {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	var out io.Writer = os.Stdout
	if format == "console" {
		out = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339Nano}
	}

	log.Logger = zerolog.New(out).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(mapLevel(level))
}

func mapLevel(level int) zerolog.Level {
	switch level {
	case LogLevelDebug:
		return zerolog.DebugLevel
	case LogLevelInfo:
		return zerolog.InfoLevel
	case LogLevelWarn:
		return zerolog.WarnLevel
	case LogLevelError:
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}
