package rpc

import (
	"context"

	"github.com/NilFoundation/nil/nil/client"
	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/api"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/types"
)

type taskRequestRpcClient struct {
	client client.RawClient
}

func NewTaskRequestRpcClient(apiEndpoint string, logger logging.Logger) api.TaskRequestHandler {
	return &taskRequestRpcClient{
		client: NewRetryClient(apiEndpoint, logger),
	}
}

func (r *taskRequestRpcClient) GetTask(ctx context.Context, request *api.TaskRequest) (*types.Task, error) {
	return doRPCCall[*api.TaskRequest, *types.Task](
		ctx,
		r.client,
		api.TaskRequestHandlerGetTask,
		request,
	)
}

func (r *taskRequestRpcClient) CheckIfTaskExists(ctx context.Context, request *api.TaskCheckRequest) (bool, error) {
	return doRPCCall[*api.TaskCheckRequest, bool](
		ctx,
		r.client,
		api.TaskRequestHandlerCheckIfTaskExists,
		request,
	)
}

func (r *taskRequestRpcClient) SetTaskResult(ctx context.Context, result *types.TaskResult) error {
	_, err := doRPCCall[*types.TaskResult, any](
		ctx,
		r.client,
		api.TaskRequestHandlerSetTaskResult,
		result,
	)
	return err
}
