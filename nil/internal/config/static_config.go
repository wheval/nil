package config

import (
	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/internal/db"
)

var _ ConfigAccessor = &StaticConfig{}

type StaticConfig struct {
	config map[string][]byte
}

func NewStaticConfig(validators []ValidatorInfo) (*StaticConfig, error) {
	config := make(map[string][]byte)

	paramValidators := &ParamValidators{List: validators}
	data, err := paramValidators.MarshalSSZ()
	if err != nil {
		return nil, err
	}

	config[paramValidators.Name()] = data
	return &StaticConfig{config: config}, nil
}

func (c *StaticConfig) GetParamData(name string) ([]byte, error) {
	return c.config[name], nil
}

func (c *StaticConfig) SetParamData(name string, data []byte) error {
	c.config[name] = data
	return nil
}

func (c *StaticConfig) Commit(db.RwTx, common.Hash) (common.Hash, error) {
	return common.EmptyHash, nil
}
