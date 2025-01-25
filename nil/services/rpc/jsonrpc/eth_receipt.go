package jsonrpc

import (
	"context"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/internal/types"
)

func (api *APIImplRo) GetInTransactionReceipt(ctx context.Context, hash common.Hash) (*RPCReceipt, error) {
	info, err := api.rawapi.GetInTransactionReceipt(ctx, types.ShardIdFromHash(hash), hash)
	if err != nil {
		return nil, err
	}
	return NewRPCReceipt(info)
}
