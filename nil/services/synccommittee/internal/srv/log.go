package srv

import "github.com/rs/zerolog"

func WorkerLogger(logger zerolog.Logger, worker Worker) zerolog.Logger {
	return logger.With().Str("worker", worker.Name()).Logger()
}
