package mpt

import (
	"github.com/NilFoundation/nil/nil/internal/db"
	"github.com/NilFoundation/nil/nil/internal/types"
)

type Getter interface {
	Get(key []byte) ([]byte, error)
}

type Setter interface {
	Set(key, value []byte) error
}

type DbGetter struct {
	tx        db.RoTx
	shardId   types.ShardId
	tableName db.ShardedTableName
}

func NewDbGetter(tx db.RoTx, shardId types.ShardId, tableName db.ShardedTableName) *DbGetter {
	return &DbGetter{tx, shardId, tableName}
}

func (g *DbGetter) Get(key []byte) ([]byte, error) {
	return g.tx.GetFromShard(g.shardId, g.tableName, key)
}

type DbSetter struct {
	tx        db.RwTx
	shardId   types.ShardId
	tableName db.ShardedTableName
}

func NewDbSetter(tx db.RwTx, shardId types.ShardId, tableName db.ShardedTableName) *DbSetter {
	return &DbSetter{tx, shardId, tableName}
}

func (s *DbSetter) Set(key, value []byte) error {
	return s.tx.PutToShard(s.shardId, s.tableName, key, value)
}

type InMemHolder map[string][]byte

func NewInMemHolder() InMemHolder {
	return make(map[string][]byte)
}

func NewMPTFromMap(m InMemHolder) *MerklePatriciaTrie {
	return NewMPT(&m, NewReader(m))
}

func NewInMemMPT() *MerklePatriciaTrie {
	return NewMPTFromMap(NewInMemHolder())
}

func (s InMemHolder) Get(key []byte) ([]byte, error) {
	return s[string(key)], nil
}

func (s *InMemHolder) Set(key, value []byte) error {
	(*s)[string(key)] = value
	return nil
}
