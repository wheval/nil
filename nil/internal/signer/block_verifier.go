package signer

import (
	"context"
	"errors"
	"fmt"

	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/internal/config"
	"github.com/NilFoundation/nil/nil/internal/db"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/rs/zerolog"
)

var errBlockVerify = errors.New("failed to verify block")

type BlockVerifier struct {
	shardId types.ShardId
	db      db.DB
}

func NewBlockVerifier(shardId types.ShardId, db db.DB) *BlockVerifier {
	return &BlockVerifier{
		shardId: shardId,
		db:      db,
	}
}

func (b *BlockVerifier) VerifyBlock(ctx context.Context, logger zerolog.Logger, block *types.Block) error {
	tx, err := b.db.CreateRoTx(ctx)
	if err != nil {
		return fmt.Errorf("%w: failed to create read-only transaction: %w", errBlockVerify, err)
	}
	defer tx.Rollback()

	var accessor config.ConfigAccessor
	_, err = db.ReadBlock(tx, types.MainShardId, block.MainChainHash)
	if err != nil {
		// It is possible that the needed main chain block has not arrived yet, or that this one is some byzantine block.
		// Because right now the config is actually constant, we can use whatever version we like in this case,
		// so we use the latest accessible config.
		// TODO(@isergeyam): create some subscription mechanism that will handle this correctly.
		if errors.Is(err, db.ErrKeyNotFound) {
			logger.Warn().
				Str(logging.FieldBlockHash, block.MainChainHash.String()).
				Msg("Main chain block not found, using the latest accessible config")
			accessor, err = config.NewConfigAccessorTx(ctx, tx, nil)
		}
	} else {
		accessor, err = config.NewConfigAccessorTx(ctx, tx, &block.MainChainHash)
	}
	if err != nil {
		return fmt.Errorf("%w: failed to create config accessor: %w", errBlockVerify, err)
	}

	validatorsList, err := config.GetParamValidators(accessor)
	if err != nil {
		return fmt.Errorf("%w: failed to get validators set: %w", errBlockVerify, err)
	}

	if b.shardId == 0 {
		b.shardId = 0
	}

	// TODO: for now check that block is signed by one known validator
	for _, v := range validatorsList.Validators[b.shardId].List {
		if err = block.VerifySignature(v.PublicKey[:], b.shardId); err == nil {
			return nil
		}
	}
	return fmt.Errorf("%w: failed to verify signature: %w", errBlockVerify, err)
}
