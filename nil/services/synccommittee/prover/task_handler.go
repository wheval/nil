package prover

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/NilFoundation/nil/nil/client"
	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/services/rpc/transport"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/api"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/log"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/storage"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/types"
	"github.com/NilFoundation/nil/nil/services/synccommittee/prover/tracer"
	"github.com/rs/zerolog"
)

type taskHandler struct {
	resultStorage storage.TaskResultStorage
	timer         common.Timer
	logger        zerolog.Logger
	config        taskHandlerConfig
	client        client.Client
}

func newTaskHandler(
	resultStorage storage.TaskResultStorage,
	timer common.Timer,
	logger zerolog.Logger,
	config taskHandlerConfig,
) api.TaskHandler {
	return &taskHandler{
		resultStorage: resultStorage,
		timer:         timer,
		logger:        logger,
		config:        config,
		client:        NewRPCClient(config.NilRpcEndpoint, logging.NewLogger("client")),
	}
}

type taskHandlerConfig struct {
	NilRpcEndpoint      string
	ProofProducerBinary string
	OutDir              string
}

func newTaskHandlerConfig(nilRpcEndpoint string) taskHandlerConfig {
	return taskHandlerConfig{
		NilRpcEndpoint:      nilRpcEndpoint,
		ProofProducerBinary: "proof-producer-multi-threaded",
		OutDir:              os.TempDir(), // TODO: replace with shared folder
	}
}

type commandDescription struct {
	runCommands           []*exec.Cmd
	expectedResult        types.TaskResultAddresses
	binaryExpectedResults types.TaskResultData
}

func circuitTypeToArg(ct types.CircuitType) string {
	switch ct {
	case types.None:
		return "none"
	case types.CircuitBytecode:
		return "bytecode"
	case types.CircuitMPT:
		return "mpt"
	case types.CircuitReadWrite:
		return "rw"
	case types.CircuitZKEVM:
		return "zkevm"
	case types.CircuitCopy:
		return "copy"
	default:
		panic("Unknown circuit type")
	}
}

func circuitIdx(ct types.CircuitType) uint8 {
	return uint8(ct)
}

func collectDependencyFiles(task *types.Task, dependencyType types.TaskType, resultType types.ProverResultType) ([]string, error) {
	depFiles := []string{}
	for _, res := range task.DependencyResults {
		if res.TaskType == dependencyType {
			path, ok := res.DataAddresses[resultType]
			if !ok {
				return depFiles, errors.New("Inconsistent task " + task.Id.String() +
					", dependencyType " + dependencyType.String() + " has no expected result " + resultType.String())
			}
			depFiles = append(depFiles, path)
		}
	}
	return depFiles, nil
}

func insufficientTaskInputMsg(task *types.Task, dependencyType string, expected int, actual int) string {
	return fmt.Sprintf("Insufficient input for task %v with type %v on %v dependency: expected %d, actual %d",
		task.Id.String(),
		task.TaskType.String(),
		dependencyType,
		expected,
		actual,
	)
}

func (h *taskHandler) generateTraces(ctx context.Context, task *types.Task) (string, error) {
	traceFile := filepath.Join(h.config.OutDir, fmt.Sprintf("trace.%v.%v", task.ShardId, task.BlockHash))

	h.logger.Info().Msgf("Tracer arguments: trace --nil-endpoint %v %v %v %v", h.config.NilRpcEndpoint, traceFile, task.ShardId, task.BlockHash.String())

	tracerConfig := tracer.TraceConfig{
		MarshalMode:  tracer.MarshalModeBinary,
		ShardID:      task.ShardId,
		BaseFileName: traceFile,
		BlockIDs:     []transport.BlockReference{transport.HashBlockReference(task.BlockHash)},
	}
	err := tracer.GenerateTrace(ctx, h.client, &tracerConfig)
	return traceFile, err
}

func makePartialProofRunCommand(task *types.Task, traceFile string, config taskHandlerConfig) (*exec.Cmd, types.TaskResultAddresses, error) {
	resultFiles := make(types.TaskResultAddresses)
	proofProducerBinary := config.ProofProducerBinary
	stage := []string{"--stage", "fast-generate-partial-proof"}
	filePostfix := fmt.Sprintf(".%v.%v.%v", circuitIdx(task.CircuitType), task.ShardId, task.BlockHash.String())
	circuitName := []string{"--circuit-name", circuitTypeToArg(task.CircuitType)}

	assignmentDescFile := filepath.Join(config.OutDir, "assignment_table_description"+filePostfix)
	assignmentDescArg := []string{"--assignment-description-file", assignmentDescFile}
	resultFiles[types.AssignmentTableDescription] = assignmentDescFile

	traceArg := []string{"--trace", traceFile}

	partialProofFile := filepath.Join(config.OutDir, "partial_proof"+filePostfix)
	partialProofArg := []string{"--proof", partialProofFile}
	resultFiles[types.PartialProof] = partialProofFile

	challengeFile := filepath.Join(config.OutDir, "challenge"+filePostfix)
	challengeArg := []string{"--challenge-file", challengeFile}
	resultFiles[types.PartialProofChallenges] = challengeFile

	thetaPowerFile := filepath.Join(config.OutDir, "theta_power"+filePostfix)
	thetaPowerArg := []string{"--theta-power-file", thetaPowerFile}
	resultFiles[types.ThetaPower] = thetaPowerFile

	commonDataFile := filepath.Join(config.OutDir, "preprocessed_common_data"+filePostfix)
	commonDataArg := []string{"--common-data", commonDataFile}
	resultFiles[types.PreprocessedCommonData] = commonDataFile

	commitmentStateFile := filepath.Join(config.OutDir, "commitment_state"+filePostfix)
	commitmentStateArg := []string{"--updated-commitment-state-file", commitmentStateFile}
	resultFiles[types.CommitmentState] = commitmentStateFile

	allArgs := slices.Concat(stage, circuitName, assignmentDescArg, traceArg, partialProofArg, challengeArg, thetaPowerArg, commitmentStateArg, commonDataArg)
	cmd := exec.Command(proofProducerBinary, allArgs...)
	return cmd, resultFiles, cmd.Err
}

func (h *taskHandler) makePartialProofCommand(ctx context.Context, task *types.Task) (commandDescription, error) {
	// First we run tracer right away, since it's a part of prover binary, so no need for separate command
	traceFile, err := h.generateTraces(ctx, task)
	if err != nil {
		return commandDescription{}, err
	}

	// Now we are ready to generate the partial proof
	proofProducerCmd, resultFiles, err := makePartialProofRunCommand(task, traceFile, h.config)
	if err != nil {
		return commandDescription{}, err
	}
	resultCommandSet := commandDescription{
		runCommands:    []*exec.Cmd{proofProducerCmd},
		expectedResult: resultFiles,
	}
	return resultCommandSet, err
}

func (h *taskHandler) makeAggregateChallengesCommand(task *types.Task) (commandDescription, error) {
	binary := h.config.ProofProducerBinary
	stage := []string{"--stage", "generate-aggregated-challenge"}
	inputFiles, err := collectDependencyFiles(task, types.PartialProve, types.PartialProofChallenges)
	if err != nil {
		return commandDescription{}, err
	}
	if len(inputFiles) != int(types.CircuitAmount) {
		err = errors.New(insufficientTaskInputMsg(task, "PartialProofChallenges", int(types.CircuitAmount), len(inputFiles)))
		return commandDescription{}, err
	}
	inputs := append([]string{"--input-challenge-files"}, inputFiles...)
	outFile := filepath.Join(h.config.OutDir,
		fmt.Sprintf("aggregated_challenges.%v.%v", task.ShardId, task.BlockHash.String()))
	outArg := []string{"--aggregated-challenge-file", outFile}
	allArgs := slices.Concat(stage, inputs, outArg)
	cmd := exec.Command(binary, allArgs...)
	return commandDescription{
		runCommands:    []*exec.Cmd{cmd},
		expectedResult: types.TaskResultAddresses{types.AggregatedChallenges: outFile},
	}, cmd.Err
}

func (h *taskHandler) makeCombinedQCommand(task *types.Task) (commandDescription, error) {
	binary := h.config.ProofProducerBinary
	stage := []string{"--stage", "compute-combined-Q"}
	commitmentStateFile, err := collectDependencyFiles(task, types.PartialProve, types.CommitmentState)
	if err != nil {
		return commandDescription{}, err
	}
	if len(commitmentStateFile) != 1 {
		err = errors.New(insufficientTaskInputMsg(task, "CommitmentState", 1, len(commitmentStateFile)))
		return commandDescription{}, err
	}
	commitmentState := []string{"--commitment-state-file", commitmentStateFile[0]}

	aggChallengesFile, err := collectDependencyFiles(task, types.AggregatedChallenge, types.AggregatedChallenges)
	if err != nil {
		return commandDescription{}, err
	}
	if len(aggChallengesFile) != 1 {
		err = errors.New(insufficientTaskInputMsg(task, "AggregatedChallenges", 1, len(aggChallengesFile)))
		return commandDescription{}, err
	}
	aggregateChallenges := []string{"--aggregated-challenge-file", aggChallengesFile[0]}

	startingPower := []string{"--combined-Q-starting-power=0"} // TODO: compute it properly from dependencies
	outFile := filepath.Join(h.config.OutDir,
		fmt.Sprintf("combined_Q.%v.%v.%v", circuitIdx(task.CircuitType), task.ShardId, task.BlockHash.String()))
	outArg := []string{"--combined-Q-polynomial-file", outFile}

	allArgs := slices.Concat(stage, commitmentState, aggregateChallenges, startingPower, outArg)
	cmd := exec.Command(binary, allArgs...)
	return commandDescription{
		runCommands:    []*exec.Cmd{cmd},
		expectedResult: types.TaskResultAddresses{types.CombinedQPolynomial: outFile},
	}, cmd.Err
}

func (h *taskHandler) makeAggregateFRICommand(task *types.Task) (commandDescription, error) {
	binary := h.config.ProofProducerBinary
	stage := []string{"--stage", "aggregated-FRI"}
	assignmentTableFile, err := collectDependencyFiles(task, types.PartialProve, types.AssignmentTableDescription)
	if err != nil {
		return commandDescription{}, err
	}
	if len(assignmentTableFile) != int(types.CircuitAmount) {
		err = errors.New(insufficientTaskInputMsg(task, "AssignmentTableDescription", int(types.CircuitAmount), len(assignmentTableFile)))
		return commandDescription{}, err
	}
	assignmentTable := []string{"--assignment-description-file", assignmentTableFile[0]}

	aggChallengeFile, err := collectDependencyFiles(task, types.AggregatedChallenge, types.AggregatedChallenges)
	if err != nil {
		return commandDescription{}, err
	}
	if len(aggChallengeFile) != 1 {
		err = errors.New(insufficientTaskInputMsg(task, "AggregatedChallenges", 1, len(aggChallengeFile)))
		return commandDescription{}, err
	}
	aggregatedChallenge := []string{"--aggregated-challenge-file", aggChallengeFile[0]}

	combinedQFiles, err := collectDependencyFiles(task, types.CombinedQ, types.CombinedQPolynomial)
	if err != nil {
		return commandDescription{}, err
	}
	if len(combinedQFiles) != int(types.CircuitAmount) {
		err = errors.New(insufficientTaskInputMsg(task, "CombinedQPolynomial", int(types.CircuitAmount), len(combinedQFiles)))
		return commandDescription{}, err
	}
	combinedQ := append([]string{"--input-combined-Q-polynomial-files"}, combinedQFiles...)

	resFiles := make(types.TaskResultAddresses)
	filePostfix := fmt.Sprintf(".%v.%v", task.ShardId, task.BlockHash.String())
	resFiles[types.AggregatedFRIProof] = filepath.Join(h.config.OutDir, "aggregated_FRI_proof"+filePostfix)
	resFiles[types.ProofOfWork] = filepath.Join(h.config.OutDir, "POW"+filePostfix)
	resFiles[types.ConsistencyCheckChallenges] = filepath.Join(h.config.OutDir, "challenges"+filePostfix)

	aggFRI := []string{"--proof", resFiles[types.AggregatedFRIProof]}
	POW := []string{"--proof-of-work-file", resFiles[types.ProofOfWork]}
	consistencyChallenges := []string{"--consistency-checks-challenges-file", resFiles[types.ConsistencyCheckChallenges]}
	allArgs := slices.Concat(stage, assignmentTable, aggregatedChallenge, combinedQ, aggFRI, POW, consistencyChallenges)
	cmd := exec.Command(binary, allArgs...)
	return commandDescription{runCommands: []*exec.Cmd{cmd}, expectedResult: resFiles}, cmd.Err
}

func (h *taskHandler) makeConsistencyCheckCommand(task *types.Task) (commandDescription, error) {
	binary := h.config.ProofProducerBinary
	stage := []string{"--stage", "consistency-checks"}
	commitmentStateFile, err := collectDependencyFiles(task, types.PartialProve, types.CommitmentState)
	if err != nil {
		return commandDescription{}, err
	}
	if len(commitmentStateFile) != 1 {
		err = errors.New(insufficientTaskInputMsg(task, "CommitmentState", 1, len(commitmentStateFile)))
		return commandDescription{}, err
	}
	commitmentState := []string{"--commitment-state-file", commitmentStateFile[0]}
	combinedQFile, err := collectDependencyFiles(task, types.CombinedQ, types.CombinedQPolynomial)
	if err != nil {
		return commandDescription{}, err
	}
	if len(combinedQFile) != 1 {
		err = errors.New(insufficientTaskInputMsg(task, "CombinedQPolynomial", 1, len(combinedQFile)))
		return commandDescription{}, err
	}
	combinedQ := []string{"--combined-Q-polynomial-file", combinedQFile[0]}
	consistencyChallengeFiles, err := collectDependencyFiles(task, types.AggregatedFRI, types.ConsistencyCheckChallenges)
	if err != nil {
		return commandDescription{}, err
	}
	if len(consistencyChallengeFiles) != 1 {
		err = errors.New(insufficientTaskInputMsg(task, "ConsistencyCheckChallenges", 1, len(consistencyChallengeFiles)))
		return commandDescription{}, err
	}
	consistencyChallenges := []string{"--consistency-checks-challenges-file", consistencyChallengeFiles[0]}

	outFile := filepath.Join(h.config.OutDir,
		fmt.Sprintf("LPC_consistency_check_proof.%v.%v.%v", circuitIdx(task.CircuitType), task.ShardId, task.BlockHash.String()))
	outArg := []string{"--proof", outFile}

	allArgs := slices.Concat(stage, commitmentState, combinedQ, consistencyChallenges, outArg)
	cmd := exec.Command(binary, allArgs...)
	return commandDescription{
		runCommands:    []*exec.Cmd{cmd},
		expectedResult: types.TaskResultAddresses{types.LPCConsistencyCheckProof: outFile},
	}, cmd.Err
}

func (h *taskHandler) makeMergeProofCommand(task *types.Task) (commandDescription, error) {
	binary := h.config.ProofProducerBinary
	stage := []string{"--stage", "merge-proofs"}
	partialProofFiles, err := collectDependencyFiles(task, types.PartialProve, types.PartialProof)
	if err != nil {
		return commandDescription{}, err
	}
	if len(partialProofFiles) != int(types.CircuitAmount) {
		err = errors.New(insufficientTaskInputMsg(task, "PartialProof", int(types.CircuitAmount), len(partialProofFiles)))
		return commandDescription{}, err
	}
	partialProofs := append([]string{"--partial-proof"}, partialProofFiles...)

	LPCCheckFiles, err := collectDependencyFiles(task, types.FRIConsistencyChecks, types.LPCConsistencyCheckProof)
	if err != nil {
		return commandDescription{}, err
	}
	if len(LPCCheckFiles) != int(types.CircuitAmount) {
		err = errors.New(insufficientTaskInputMsg(task, "LPCConsistencyCheckProof", int(types.CircuitAmount), len(LPCCheckFiles)))
		return commandDescription{}, err
	}
	LPCChecks := append([]string{"--initial-proof"}, LPCCheckFiles...)

	aggFRIFile, err := collectDependencyFiles(task, types.AggregatedFRI, types.AggregatedFRIProof)
	if err != nil {
		return commandDescription{}, err
	}
	if len(aggFRIFile) != 1 {
		err = errors.New(insufficientTaskInputMsg(task, "AggregatedFRIProof", int(types.CircuitAmount), len(aggFRIFile)))
		return commandDescription{}, err
	}
	aggFRI := []string{"--aggregated-FRI-proof", aggFRIFile[0]}

	outFile := filepath.Join(h.config.OutDir,
		fmt.Sprintf("final-proof.%v.%v", task.ShardId, task.BlockHash.String()))
	outArg := []string{"--proof", outFile}

	allArgs := slices.Concat(stage, partialProofs, LPCChecks, aggFRI, outArg)
	cmd := exec.Command(binary, allArgs...)
	return commandDescription{
		runCommands:    []*exec.Cmd{cmd},
		expectedResult: types.TaskResultAddresses{types.FinalProof: outFile},
	}, cmd.Err
}

func (h *taskHandler) makeAggregateProofCommand(task *types.Task) (commandDescription, error) {
	binary := "echo" // TODO: enable aggregate proof command once it will be implemented
	stage := []string{"--stage", "aggregate-proofs"}
	blockProofFiles, err := collectDependencyFiles(task, types.MergeProof, types.FinalProof)
	if err != nil {
		return commandDescription{}, err
	}
	blockProofs := append([]string{"--block-proof"}, blockProofFiles...)

	outFile := filepath.Join(h.config.OutDir,
		fmt.Sprintf("aggregated-proof.%v.%v", task.ShardId, task.BlockHash.String()))
	outArg := []string{"--proof", outFile}

	allArgs := slices.Concat(stage, blockProofs, outArg)
	var aggregatedProof []byte
	for _, dependency := range task.DependencyResults {
		aggregatedProof = append(aggregatedProof, dependency.Data...)
	}
	cmd := exec.Command(binary, allArgs...)
	return commandDescription{
		runCommands:           []*exec.Cmd{cmd},
		expectedResult:        types.TaskResultAddresses{types.AggregatedProof: outFile},
		binaryExpectedResults: aggregatedProof,
	}, cmd.Err
}

func (h *taskHandler) makeCommandForTask(ctx context.Context, task *types.Task) (commandDescription, error) {
	switch task.TaskType {
	case types.PartialProve:
		return h.makePartialProofCommand(ctx, task)
	case types.AggregatedChallenge:
		return h.makeAggregateChallengesCommand(task)
	case types.CombinedQ:
		return h.makeCombinedQCommand(task)
	case types.AggregatedFRI:
		return h.makeAggregateFRICommand(task)
	case types.FRIConsistencyChecks:
		return h.makeConsistencyCheckCommand(task)
	case types.MergeProof:
		return h.makeMergeProofCommand(task)
	case types.AggregateProofs:
		return h.makeAggregateProofCommand(task)
	case types.ProofBlock:
		err := errors.New("ProofBlock task type is not supposed to be encountered in prover task handler for task " + task.Id.String() +
			" type " + task.TaskType.String())
		return commandDescription{}, err
	case types.TaskTypeNone:
		err := fmt.Errorf("task with id=%s has unspecified type", task.Id.String())
		return commandDescription{}, err
	default:
		err := errors.New("Unknown type for task " + task.Id.String() +
			" type " + task.TaskType.String())
		return commandDescription{}, err
	}
}

func (h *taskHandler) Handle(ctx context.Context, executorId types.TaskExecutorId, task *types.Task) error {
	if task.TaskType == types.ProofBlock {
		err := types.UnexpectedTaskType(task)
		taskResult := types.NewFailureProverTaskResult(task.Id, executorId, fmt.Errorf("failed to create command for task: %w", err))
		log.NewTaskEvent(h.logger, zerolog.ErrorLevel, task).Err(err).Msg("failed to create command for task")
		return h.resultStorage.Put(ctx, taskResult)
	}
	desc, err := h.makeCommandForTask(ctx, task)
	if err != nil {
		taskResult := types.NewFailureProverTaskResult(task.Id, executorId, fmt.Errorf("failed to create command for task: %w", err))
		log.NewTaskEvent(h.logger, zerolog.ErrorLevel, task).
			Err(err).
			Msg("failed to create command for task")
		return h.resultStorage.Put(ctx, taskResult)
	}
	startTime := h.timer.NowTime()
	log.NewTaskEvent(h.logger, zerolog.InfoLevel, task).Msg("Starting task execution")
	for _, cmd := range desc.runCommands {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		cmdString := strings.Join(cmd.Args, " ")
		h.logger.Info().Msgf("Run command %v\n", cmdString)
		err := cmd.Run()
		h.logger.Trace().Msgf("Task execution stdout:\n%v\n", stdout.String())
		if err != nil {
			taskResult := types.NewFailureProverTaskResult(task.Id, executorId, fmt.Errorf("task execution failed: %w", err))
			timeSpent := h.timer.NowTime().Sub(startTime)
			log.NewTaskEvent(h.logger, zerolog.ErrorLevel, task).
				Str("commandText", cmdString).
				Dur(logging.FieldTaskExecTime, timeSpent).
				Msgf("Task execution failed, stderr:\n%s\n", stderr.String())
			return h.resultStorage.Put(ctx, taskResult)
		}
	}

	executionTime := h.timer.NowTime().Sub(startTime)
	log.NewTaskEvent(h.logger, zerolog.InfoLevel, task).
		Dur(logging.FieldTaskExecTime, executionTime).
		Msg("Task execution completed successfully")

	taskResult := types.NewSuccessProverTaskResult(task.Id, executorId, desc.expectedResult, desc.binaryExpectedResults)
	return h.resultStorage.Put(ctx, taskResult)
}
