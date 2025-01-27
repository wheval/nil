package signer

import (
	"context"
	"errors"
	"fmt"

	"github.com/NilFoundation/nil/nil/internal/config"
	"github.com/NilFoundation/nil/nil/internal/types"
)

var errBlockVerify = errors.New("failed to verify block")

type BlockVerifier struct {
	shardId    types.ShardId
	validators []config.ValidatorInfo
}

func NewBlockVerifier(shardId types.ShardId, validators []config.ValidatorInfo) *BlockVerifier {
	return &BlockVerifier{
		shardId:    shardId,
		validators: validators,
	}
}

func (b *BlockVerifier) VerifyBlock(ctx context.Context, block *types.Block) error {
	accessor, err := config.NewStaticConfig(b.validators)
	if err != nil {
		return err
	}

	validators, err := config.GetParamValidators(accessor)
	if err != nil {
		return fmt.Errorf("%w: failed to get validators set: %w", errBlockVerify, err)
	}

	// TODO: for now check that block is signed by one known validator
	for _, v := range validators.List {
		if err = block.VerifySignature(v.PublicKey[:], b.shardId); err == nil {
			return nil
		}
	}
	return fmt.Errorf("%w: failed to verify signature: %w", errBlockVerify, err)
}
