package logging

import "github.com/rs/zerolog"

type CHLogger struct {
	l zerolog.Logger
}

func NewCHLogger(component, table string) CHLogger {
	return CHLogger{
		NewLogger(component).With().
			Bool("store_to_clickhouse", true).
			Str("database", "metrics").
			Str("table", table).
			Logger(),
	}
}

func (l CHLogger) Disable() CHLogger {
	l.l = zerolog.Nop()
	return l
}

func (l CHLogger) Log() *zerolog.Event {
	// cheap trick to ensure these logs are always written
	// despite of zerolog global log level
	return l.l.WithLevel(zerolog.PanicLevel) //nolint:zerologlint
}
