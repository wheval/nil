package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/NilFoundation/nil/nil/common/check"
	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/internal/cobrax"
	"github.com/NilFoundation/nil/nil/internal/db"
	"github.com/NilFoundation/nil/nil/internal/profiling"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/rpc/transport"
	"github.com/NilFoundation/nil/nil/services/synccommittee/prover"
	"github.com/NilFoundation/nil/nil/services/synccommittee/prover/tracer"
	"github.com/spf13/cobra"
)

const appTitle = "=nil; Prover"

func main() {
	check.PanicIfNotCancelledErr(execute())
}

type CommonConfig = prover.Config

type RunConfig struct {
	*CommonConfig `yaml:",inline"`

	DbPath string `yaml:"dbPath"`
}

type PrintConfig struct {
	BaseFileName string
	MarshalMode  string
}

func execute() error {
	rootCmd := &cobra.Command{
		Use:   os.Args[0],
		Short: "Run nil prover node",
	}

	runConfig, err := loadRunConfig()
	if err != nil {
		return err
	}

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the prover service",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(runConfig)
		},
	}
	commonCfg := runConfig.CommonConfig
	addCommonFlags(rootCmd, commonCfg)
	runCmd.Flags().StringVar(&runConfig.DbPath, "db-path", runConfig.DbPath, "path to database")

	traceConfig := tracer.TraceConfig{}
	var marshalModePlaceholder string
	generateTraceCmd := &cobra.Command{
		Use:   "trace [base_file_name] [shard_id] [block_ids...]",
		Short: "Collect traces for a block, dump into file",
		Args:  cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			traceConfig.BaseFileName = args[0]
			var err error
			traceConfig.MarshalMode, err = tracer.MarshalModeFromString(marshalModePlaceholder)
			if err != nil {
				return err
			}
			traceConfig.BlockIDs = make([]tracer.BlockId, len(args)-2)
			for i, blockArg := range args[2:] {
				traceConfig.BlockIDs[i].ShardId, err = types.ParseShardIdFromString(args[1])
				if err != nil {
					return err
				}
				traceConfig.BlockIDs[i].Id, err = transport.AsBlockReference(blockArg)
				if err != nil {
					return err
				}
			}
			client := prover.NewRPCClient(commonCfg.NilRpcEndpoint, logging.NewLogger("client"))
			return tracer.CollectTracesToFile(context.Background(), client, &traceConfig)
		},
	}
	addMarshalModeFlag(generateTraceCmd, &marshalModePlaceholder)

	var printConfig PrintConfig
	printTraceCmd := &cobra.Command{
		Use:   "print [file_name]",
		Short: "Read serialized traces from files, print them into console",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			printConfig.BaseFileName = args[0]
			return readTrace(&printConfig)
		},
	}
	addMarshalModeFlag(printTraceCmd, &printConfig.MarshalMode)

	versionCmd := cobrax.VersionCmd(appTitle)
	rootCmd.AddCommand(runCmd, generateTraceCmd, printTraceCmd, versionCmd)
	return rootCmd.Execute()
}

func loadRunConfig() (*RunConfig, error) {
	cfg := &RunConfig{
		CommonConfig: prover.NewDefaultConfig(),
		DbPath:       "prover.db",
	}

	if err := cobrax.LoadConfigFromFile(cobrax.GetConfigNameFromArgs(), cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func addCommonFlags(cmd *cobra.Command, cfg *CommonConfig) {
	cobrax.AddConfigFlag(cmd.PersistentFlags())
	cmd.PersistentFlags().StringVar(
		&cfg.ProofProviderRpcEndpoint,
		"proof-provider-endpoint",
		cfg.ProofProviderRpcEndpoint,
		"proof provider rpc endpoint")
	cmd.PersistentFlags().StringVar(&cfg.NilRpcEndpoint, "nil-endpoint", cfg.NilRpcEndpoint, "nil rpc endpoint")
	logLevel := cmd.PersistentFlags().String("log-level", "info", "log level: trace|debug|info|warn|error|fatal|panic")

	cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		logging.SetupGlobalLogger(*logLevel)
	}
}

func addMarshalModeFlag(cmd *cobra.Command, placeholder *string) {
	cmd.Flags().StringVar(
		placeholder,
		"marshal-mode",
		tracer.MarshalModeBinary.String(),
		"marshal modes (bin,json) for trace files separated by ','")
}

func run(cfg *RunConfig) error {
	profiling.Start(profiling.DefaultPort)

	serviceConfig := prover.Config{
		NilRpcEndpoint:           cfg.NilRpcEndpoint,
		ProofProviderRpcEndpoint: cfg.ProofProviderRpcEndpoint,
	}

	database, err := db.NewBadgerDb(cfg.DbPath)
	if err != nil {
		return fmt.Errorf("failed to create new BadgerDB: %w", err)
	}
	defer database.Close()

	service, err := prover.New(serviceConfig, database)
	if err != nil {
		return fmt.Errorf("failed to create prover service: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	err = service.Run(ctx)
	if err != nil {
		return fmt.Errorf("service exited with error: %w", err)
	}

	return nil
}

func readTrace(cfg *PrintConfig) error {
	mode, err := tracer.MarshalModeFromString(cfg.MarshalMode)
	if err != nil {
		return err
	}

	blockTraces, err := tracer.DeserializeFromFile(cfg.BaseFileName, mode)
	if err != nil {
		return err
	}
	fmt.Printf("%+v", blockTraces)
	return nil
}
