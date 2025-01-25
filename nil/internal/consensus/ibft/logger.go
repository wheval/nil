package ibft

import (
	"github.com/NilFoundation/nil/nil/common/check"
	"github.com/rs/zerolog"
)

type ibftLogger struct {
	logger zerolog.Logger
}

func logArgs(event *zerolog.Event, msg string, args ...any) {
	// go-ibft uses logger with following format: Log("message", "key1", value1, "key2", value2, ...)
	if len(args)%2 == 0 {
		for i := 0; i < len(args); i += 2 {
			key, ok := args[i].(string)
			check.PanicIfNotf(ok, "key must be a string: %s", args[i])
			event = event.Any(key, args[i+1])
		}
		event.Msg(msg)
	} else {
		event.Msgf(msg, args...)
	}
}

func (l *ibftLogger) Info(msg string, args ...any) {
	logArgs(l.logger.Info(), msg, args...)
}

func (l *ibftLogger) Debug(msg string, args ...any) {
	logArgs(l.logger.Debug(), msg, args...)
}

func (l *ibftLogger) Error(msg string, args ...any) {
	logArgs(l.logger.Error(), msg, args...)
}
