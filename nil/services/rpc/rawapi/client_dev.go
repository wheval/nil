package rawapi

import (
	"context"

	"github.com/NilFoundation/nil/nil/common/check"
	"github.com/NilFoundation/nil/nil/internal/network"
	"github.com/NilFoundation/nil/nil/internal/types"
)

type shardApiClientDev struct {
	shardApiRequestPerformer
}

var _ ShardApiDev = (*shardApiClientDev)(nil)

func constructShardApiClientDev(performer shardApiRequestPerformer) *shardApiClientDev {
	return &shardApiClientDev{
		shardApiRequestPerformer: performer,
	}
}

func newShardApiClientNetworkDev(shardId types.ShardId, networkManager network.Manager) *shardApiClientDev {
	client, err := newShardApiClientNetwork[shardApiClientDev, ShardApiDev, NetworkTransportProtocolDev](
		constructShardApiClientDev, shardId, apiNameDev, networkManager)
	check.PanicIfErr(err)
	return client
}

func (api *shardApiClientDev) DoPanicOnShard(ctx context.Context) (uint64, error) {
	return sendRequestAndGetResponseWithCallerMethodName[uint64](
		ctx, api.shardApiRequestPerformer, "DoPanicOnShard")
}
