package types

import (
	"fmt"
	"time"
)

type ProverResultType uint8

const (
	_ ProverResultType = iota
	PartialProof
	CommitmentState
	PartialProofChallenges
	AssignmentTableDescription
	ThetaPower
	AggregatedThetaPowers
	PreprocessedCommonData
	AggregatedChallenges
	CombinedQPolynomial
	AggregatedFRIProof
	ProofOfWork
	ConsistencyCheckChallenges
	LPCConsistencyCheckProof
	FinalProof
	BlockProof
	AggregatedProof
)

type TaskOutputArtifacts map[ProverResultType]string

type TaskResultData []byte

// TaskResult represents the result of a task provided via RPC by the executor with id = TaskResult.Sender.
type TaskResult struct {
	TaskId          TaskId              `json:"taskId"`
	IsSuccess       bool                `json:"isSuccess"`
	ErrorText       string              `json:"errorText,omitempty"`
	Sender          TaskExecutorId      `json:"sender"`
	OutputArtifacts TaskOutputArtifacts `json:"dataAddresses,omitempty"`
	Data            TaskResultData      `json:"binaryData,omitempty"`
}

func NewSuccessProviderTaskResult(
	taskId TaskId,
	proofProviderId TaskExecutorId,
	outputArtifacts TaskOutputArtifacts,
	binaryData TaskResultData,
) *TaskResult {
	return &TaskResult{
		TaskId:          taskId,
		IsSuccess:       true,
		Sender:          proofProviderId,
		OutputArtifacts: outputArtifacts,
		Data:            binaryData,
	}
}

func NewFailureProviderTaskResult(
	taskId TaskId,
	proofProviderId TaskExecutorId,
	err error,
) *TaskResult {
	return &TaskResult{
		TaskId:    taskId,
		IsSuccess: false,
		Sender:    proofProviderId,
		ErrorText: fmt.Sprintf("failed to proof block: %v", err),
	}
}

func NewSuccessProverTaskResult(
	taskId TaskId,
	sender TaskExecutorId,
	outputArtifacts TaskOutputArtifacts,
	binaryData TaskResultData,
) *TaskResult {
	return &TaskResult{
		TaskId:          taskId,
		IsSuccess:       true,
		Sender:          sender,
		OutputArtifacts: outputArtifacts,
		Data:            binaryData,
	}
}

func NewFailureProverTaskResult(
	taskId TaskId,
	sender TaskExecutorId,
	err error,
) *TaskResult {
	return &TaskResult{
		TaskId:    taskId,
		Sender:    sender,
		IsSuccess: false,
		ErrorText: fmt.Sprintf("failed to generate proof: %v", err),
	}
}

// TaskResultDetails represents the result of a task, extending TaskResult with additional task-specific metadata.
type TaskResultDetails struct {
	TaskResult
	TaskType      TaskType      `json:"type"`
	CircuitType   CircuitType   `json:"circuitType"`
	ExecutionTime time.Duration `json:"executionTime"`
}

func NewTaskResultDetails(result *TaskResult, taskEntry *TaskEntry, currentTime time.Time) *TaskResultDetails {
	return &TaskResultDetails{
		TaskResult:    *result,
		TaskType:      taskEntry.Task.TaskType,
		CircuitType:   taskEntry.Task.CircuitType,
		ExecutionTime: *taskEntry.ExecutionTime(currentTime),
	}
}
