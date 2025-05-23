package collate

import (
	"fmt"

	"github.com/NilFoundation/nil/nil/internal/types"
)

type ShardTopology interface {
	// returns list of neighbor shard ids
	GetNeighbors(id types.ShardId, nShards uint32, includeSelf bool) []types.ShardId

	// returns whether we need to propagate transaction from `from` shard to `dest` shard via `cur` shard
	ShouldPropagateTxn(from types.ShardId, cur types.ShardId, dest types.ShardId) bool
}

type NeighbouringShardTopology struct{}

var neighbouringShardTopology ShardTopology = new(NeighbouringShardTopology)

func (*NeighbouringShardTopology) GetNeighbors(id types.ShardId, nShards uint32, includeSelf bool) []types.ShardId {
	var leftId, rightId types.ShardId
	if id > 0 {
		leftId = id - 1
	} else {
		leftId = types.ShardId(nShards - 1)
	}
	if id < types.ShardId(nShards)-1 {
		rightId = id + 1
	} else {
		rightId = 0
	}
	res := []types.ShardId{leftId}
	if rightId != leftId {
		res = append(res, rightId)
	}
	if includeSelf {
		res = append(res, id)
	}
	return res
}

func (*NeighbouringShardTopology) ShouldPropagateTxn(from types.ShardId, cur types.ShardId, dest types.ShardId) bool {
	// check that from->cur and cur->dest are in the same direction
	return (from < cur) == (cur < dest)
}

type TrivialShardTopology struct{}

var trivialShardTopology ShardTopology = new(TrivialShardTopology)

func (*TrivialShardTopology) GetNeighbors(id types.ShardId, nShards uint32, includeSelf bool) []types.ShardId {
	nb := make([]types.ShardId, 0, nShards)
	for i := range types.ShardId(nShards) {
		if i != id || includeSelf {
			nb = append(nb, i)
		}
	}
	return nb
}

func (*TrivialShardTopology) ShouldPropagateTxn(from types.ShardId, cur types.ShardId, dest types.ShardId) bool {
	return false
}

const (
	TrivialShardTopologyId      = "TrivialShardTopology"
	NeighbouringShardTopologyId = "NeighbouringShardTopology"
)

func GetShardTopologyById(id string) ShardTopology {
	switch id {
	case TrivialShardTopologyId:
		return trivialShardTopology
	case NeighbouringShardTopologyId:
		return neighbouringShardTopology
	}
	panic(fmt.Errorf("unknown shard topology id: %v", id))
}
