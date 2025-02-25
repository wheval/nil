package jsonrpc

import (
	"context"

	"github.com/NilFoundation/nil/nil/internal/types"
)

func (api *APIImplRo) GetShardIdList(ctx context.Context) ([]types.ShardId, error) {
	return api.rawapi.GetShardIdList(ctx)
}

func (api *APIImplRo) GetNumShards(ctx context.Context) (uint64, error) {
	return api.rawapi.GetNumShards(ctx)
}
