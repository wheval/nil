package ibft

import (
	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/common/check"
	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/internal/config"
)

func (i *backendIBFT) calcProposer(height, round uint64) (*config.ValidatorInfo, error) {
	var hash *common.Hash
	key := mainBlockMapKey(height, round)
	hashAny, ok := i.mainBlockMap.Load(key)
	if !ok {
		// This can happen when IsProposer is called to determine if the node is the proposer instead of validating the message
		// In this case we want to use the latest config
		hash = nil
	} else {
		hashRaw, ok := hashAny.(common.Hash)
		check.PanicIfNotf(ok, "Failed to convert main block hash to common.Hash")
		hash = &hashRaw
	}

	validators, err := i.getValidators(i.ctx, hash)
	if err != nil {
		i.logger.Error().
			Err(err).
			Uint64(logging.FieldRound, round).
			Uint64(logging.FieldHeight, height).
			Msg("Failed to get validators")
		return nil, err
	}

	index := (height + round) % uint64(len(validators))
	return &validators[index], nil
}
