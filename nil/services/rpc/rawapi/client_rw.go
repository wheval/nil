package rawapi

import (
	"context"
	"github.com/NilFoundation/nil/nil/common/check"
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

func newShardApiClientNetworkRw(shardId types.ShardId, networkManager network.Manager) *shardApiClientRw {
	client, err := newShardApiClientNetwork[*shardApiClientRw](
		shardId,
		apiNameRw,
		networkManager,
		reflect.TypeFor[ShardApiRw](),
		reflect.TypeFor[NetworkTransportProtocolRw]())
	check.PanicIfErr(err)
	return client
}

func newShardApiClientDirectEmulatorRw(shardApi ShardApiRw) *shardApiClientRw {
	client, err := newShardApiClientDirectEmulator[ShardApiRw, *shardApiClientRw](
		apiNameRw, shardApi, reflect.TypeFor[ShardApiRw](), reflect.TypeFor[NetworkTransportProtocolRw]())
	check.PanicIfErr(err)
	return client
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
