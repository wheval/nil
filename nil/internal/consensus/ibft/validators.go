package ibft

import (
	"github.com/NilFoundation/nil/nil/internal/config"
)

func (i *backendIBFT) calcProposer(height, round uint64) config.ValidatorInfo {
	index := (height + round) % uint64(len(i.validators))
	return i.validators[index]
}
