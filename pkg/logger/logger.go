package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func InitLogger(env string, logLevel string) {
	// Parse log level
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// In development, write human-readable formatted logs.
	// In production/staging, write raw structured JSON.
	if env == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	} else {
		// Production JSON logging
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	log.Info().Msgf("Logger initialized with level: %s", level.String())
}

// Log wrappers for ease of use
func Info(message string, args ...interface{}) {
	log.Info().Msgf(message, args...)
}

func Error(err error, message string, args ...interface{}) {
	log.Error().Err(err).Msgf(message, args...)
}

func Warn(message string, args ...interface{}) {
	log.Warn().Msgf(message, args...)
}

func Debug(message string, args ...interface{}) {
	log.Debug().Msgf(message, args...)
}

func Fatal(err error, message string, args ...interface{}) {
	log.Fatal().Err(err).Msgf(message, args...)
}
