package rawapi

import (
	"context"
	"reflect"

	"github.com/NilFoundation/nil/nil/internal/network"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/txnpool"
)

type shardApiClientRw struct {
	shardApiRequestPerformer
}

var (
	_ ShardApiRw                     = (*shardApiClientRw)(nil)
	_ shardApiRequestPerformerSetter = (*shardApiClientRw)(nil)
)

func newShardApiClientNetworkRw(shardId types.ShardId, networkManager network.Manager) (*shardApiClientRw, error) {
	return newShardApiClientNetwork[*shardApiClientRw](
		shardId,
		apiNameRw,
		networkManager,
		reflect.TypeFor[ShardApiRw](),
		reflect.TypeFor[NetworkTransportProtocolRw]())
}

func newShardApiClientDirectEmulatorRw(shardApi ShardApiRw) (*shardApiClientRw, error) {
	return newShardApiClientDirectEmulator[ShardApiRw, *shardApiClientRw](
		apiNameRw, shardApi, reflect.TypeFor[ShardApiRw](), reflect.TypeFor[NetworkTransportProtocolRw]())
}

func (api *shardApiClientRw) setShardApiRequestPerformer(requestPerformer shardApiRequestPerformer) {
	api.shardApiRequestPerformer = requestPerformer
}

func (api *shardApiClientRw) SendTransaction(ctx context.Context, transaction []byte) (txnpool.DiscardReason, error) {
	return sendRequestAndGetResponseWithCallerMethodName[txnpool.DiscardReason](
		ctx, api, "SendTransaction", transaction)
}

func (api *shardApiClientRw) DoPanicOnShard(ctx context.Context) (uint64, error) {
	return sendRequestAndGetResponseWithCallerMethodName[uint64](ctx, api, "DoPanicOnShard")
}
