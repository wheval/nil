package rawapi

import (
	"github.com/NilFoundation/nil/nil/common/assert"
	"github.com/NilFoundation/nil/nil/internal/db"
	"github.com/NilFoundation/nil/nil/internal/network"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/txnpool"
)

type nodeApiBuilder struct {
	nodeApi *nodeApiOverShardApis
}

func NodeApiBuilder() *nodeApiBuilder {
	return &nodeApiBuilder{
		nodeApi: &nodeApiOverShardApis{
			apisRo:  make(map[types.ShardId]ShardApiRo),
			apisRw:  make(map[types.ShardId]ShardApiRw),
			allApis: make([]shardApiBase, 0),
		},
	}
}

func (nb *nodeApiBuilder) BuildAndReset() NodeApi {
	rv := *nb.nodeApi
	nb.nodeApi = &nodeApiOverShardApis{}
	for _, api := range rv.allApis {
		api.setNodeApi(&rv)
	}
	return &rv
}

func (nb *nodeApiBuilder) WithLocalShardApiRo(
	shardId types.ShardId,
	db db.ReadOnlyDB,
	txnpool txnpool.Pool,
	enableDevApi bool,
) error {
	var localShardApi ShardApiRo = newLocalShardApiRo(shardId, db, txnpool, enableDevApi)
	if assert.Enable {
		var err error
		localShardApi, err = newShardApiClientDirectEmulatorRo(localShardApi)
		if err != nil {
			return err
		}
	}
	nb.nodeApi.apisRo[shardId] = localShardApi
	nb.nodeApi.allApis = append(nb.nodeApi.allApis, localShardApi)
	return nil
}

func (nb *nodeApiBuilder) WithLocalShardApiRw(
	shardId types.ShardId,
	db db.ReadOnlyDB,
	txnpool txnpool.Pool,
	enableDevApi bool,
) error {
	var localShardApi ShardApiRw = newLocalShardApiRw(newLocalShardApiRo(shardId, db, txnpool, enableDevApi))
	if assert.Enable {
		var err error
		localShardApi, err = newShardApiClientDirectEmulatorRw(localShardApi)
		if err != nil {
			return err
		}
	}
	nb.nodeApi.apisRw[shardId] = localShardApi
	nb.nodeApi.allApis = append(nb.nodeApi.allApis, localShardApi)
	return nil
}

func (nb *nodeApiBuilder) WithNetworkShardApiClientRo(shardId types.ShardId, networkManager network.Manager) error {
	networkShardApiClient, err := newShardApiClientNetworkRo(shardId, networkManager)
	if err != nil {
		return err
	}
	nb.nodeApi.apisRo[shardId] = networkShardApiClient
	nb.nodeApi.allApis = append(nb.nodeApi.allApis, networkShardApiClient)
	return nil
}

func (nb *nodeApiBuilder) WithNetworkShardApiClientRw(shardId types.ShardId, networkManager network.Manager) error {
	networkShardApiClient, err := newShardApiClientNetworkRw(shardId, networkManager)
	if err != nil {
		return err
	}
	nb.nodeApi.apisRw[shardId] = networkShardApiClient
	nb.nodeApi.allApis = append(nb.nodeApi.allApis, networkShardApiClient)
	return nil
}
