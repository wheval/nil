package pb

import (
	"fmt"

	"github.com/NilFoundation/nil/nil/internal/network"
	rpctypes "github.com/NilFoundation/nil/nil/services/rpc/types"
	"gopkg.in/yaml.v3"
)

func (bc *BootstrapConfigResponse) PackProtoMessage(config *rpctypes.BootstrapConfig, err error) error {
	if err != nil {
		bc.Result = &BootstrapConfigResponse_Error{Error: new(Error).PackProtoMessage(err)}
		return nil
	}

	zeroStateConfigYaml, err := yaml.Marshal(config.ZeroStateConfig)
	if err != nil {
		return err
	}

	dhtBootstrapPeers, err := config.DhtBootstrapPeers.Strings()
	if err != nil {
		return err
	}

	bootstrapPeers, err := config.BootstrapPeers.Strings()
	if err != nil {
		return err
	}

	bc.Result = &BootstrapConfigResponse_Data{
		Data: &BootstrapConfig{
			NShards:             config.NShards,
			ZeroStateConfigYaml: string(zeroStateConfigYaml),
			DhtBootstrapPeers:   dhtBootstrapPeers,
			BootstrapPeers:      bootstrapPeers,
		},
	}

	return nil
}

func (bc *BootstrapConfigResponse) UnpackProtoMessage() (*rpctypes.BootstrapConfig, error) {
	switch res := bc.GetResult().(type) {
	case *BootstrapConfigResponse_Data:
		config := &rpctypes.BootstrapConfig{
			NShards: res.Data.GetNShards(),
		}
		if err := yaml.Unmarshal([]byte(res.Data.GetZeroStateConfigYaml()), &config.ZeroStateConfig); err != nil {
			return nil, err
		}
		config.DhtBootstrapPeers = make(network.AddrInfoSlice, 0)
		if err := config.DhtBootstrapPeers.FromStrings(res.Data.GetDhtBootstrapPeers()); err != nil {
			return nil, err
		}

		config.BootstrapPeers = make(network.AddrInfoSlice, 0)
		if err := config.BootstrapPeers.FromStrings(res.Data.GetBootstrapPeers()); err != nil {
			return nil, err
		}

		return config, nil

	case *BootstrapConfigResponse_Error:
		return nil, res.Error.UnpackProtoMessage()
	}
	return nil, fmt.Errorf("unexpected response type: %T", bc.GetResult())
}
