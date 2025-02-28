package reset

import (
	"context"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/rs/zerolog"
)

type StateResetter interface {
	// ResetProgressPartial resets Sync Committee's block processing progress
	// to a point preceding main shard block with the specified hash.
	ResetProgressPartial(ctx context.Context, firstMainHashToPurge common.Hash) error

	// ResetProgressNotProved resets Sync Committee's progress for all not yet proven blocks.
	ResetProgressNotProved(ctx context.Context) error
}

func NewStateResetter(logger zerolog.Logger, resetters ...StateResetter) StateResetter {
	return &compositeStateResetter{
		resetters: resetters,
		logger:    logger,
	}
}

type compositeStateResetter struct {
	resetters []StateResetter
	logger    zerolog.Logger
}

func (r *compositeStateResetter) ResetProgressPartial(ctx context.Context, firstMainHashToPurge common.Hash) error {
	r.logger.Info().
		Stringer(logging.FieldBlockMainChainHash, firstMainHashToPurge).
		Msg("Started partial progress reset")

	for _, resetter := range r.resetters {
		if err := resetter.ResetProgressPartial(ctx, firstMainHashToPurge); err != nil {
			return err
		}
	}

	r.logger.Info().
		Stringer(logging.FieldBlockMainChainHash, firstMainHashToPurge).
		Msg("Finished partial progress reset")

	return nil
}

func (r *compositeStateResetter) ResetProgressNotProved(ctx context.Context) error {
	r.logger.Info().Msg("Started not proven progress reset")

	for _, resetter := range r.resetters {
		if err := resetter.ResetProgressNotProved(ctx); err != nil {
			return err
		}
	}

	r.logger.Info().Msg("Finished not proven progress reset")
	return nil
}
