package rawapi

import (
	"context"
	"errors"
	"time"
)

func (api *localShardApiRw) DoPanicOnShard(ctx context.Context) (uint64, error) {
	// FIXME: move to devApi
	if !api.roApi.enableDevApi {
		return 0, errors.New("dev api is not enabled")
	}
	go func() {
		time.Sleep(10 * time.Second)
		panic("RPC request for panic on shard")
	}()
	return 0, nil
}
