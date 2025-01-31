package ibft

import (
	"github.com/rs/zerolog"
)

type ibftLogger struct {
	logger zerolog.Logger
}

func (l *ibftLogger) Info(msg string, args ...any) {
	l.logger.Info().Fields(args).Msg(msg)
}

func (l *ibftLogger) Debug(msg string, args ...any) {
	l.logger.Debug().Fields(args).Msg(msg)
}

func (l *ibftLogger) Error(msg string, args ...any) {
	l.logger.Error().Fields(args).Msg(msg)
}
