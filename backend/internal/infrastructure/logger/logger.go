package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Logger zerolog.Logger

func Init(env string) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if env == "production" {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "2006-01-02 15:04:05"})
		Logger = log.With().Timestamp().Logger()
	}
}

func Info(args ...interface{}) {
	Logger.Info().Msgf("%v", args)
}

func Infof(format string, args ...interface{}) {
	Logger.Info().Msgf(format, args...)
}

func Error(args ...interface{}) {
	Logger.Error().Msgf("%v", args)
}

func Errorf(format string, args ...interface{}) {
	Logger.Error().Msgf(format, args...)
}

func Debug(args ...interface{}) {
	Logger.Debug().Msgf("%v", args)
}

func Debugf(format string, args ...interface{}) {
	Logger.Debug().Msgf(format, args...)
}

func Warn(args ...interface{}) {
	Logger.Warn().Msgf("%v", args)
}

func Warnf(format string, args ...interface{}) {
	Logger.Warn().Msgf(format, args...)
}

func Fatal(args ...interface{}) {
	Logger.Fatal().Msgf("%v", args)
}

func Fatalf(format string, args ...interface{}) {
	Logger.Fatal().Msgf(format, args...)
}

func WithField(key string, value interface{}) *zerolog.Event {
	return Logger.Info().Str(key, formatValue(value))
}

func WithFields(fields map[string]interface{}) *zerolog.Event {
	event := Logger.Info()
	for k, v := range fields {
		event = event.Str(k, formatValue(v))
	}
	return event
}

func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case error:
		return val.Error()
	default:
		return ""
	}
}
