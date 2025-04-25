package main

import (
	"context"
	"errors"
	"os"
	"time"

	rpc_client "github.com/NilFoundation/nil/nil/client/rpc"
	"github.com/NilFoundation/nil/nil/cmd/nild/nildconfig"
	"github.com/NilFoundation/nil/nil/common/check"
	"github.com/NilFoundation/nil/nil/common/concurrent"
	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/internal/cobrax"
	"github.com/NilFoundation/nil/nil/internal/cobrax/cmdflags"
	"github.com/NilFoundation/nil/nil/internal/db"
	"github.com/NilFoundation/nil/nil/internal/profiling"
	"github.com/NilFoundation/nil/nil/internal/readthroughdb"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/cometa"
	"github.com/NilFoundation/nil/nil/services/indexer"
	"github.com/NilFoundation/nil/nil/services/nilservice"
	"github.com/NilFoundation/nil/nil/services/rpc/transport"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	appTitle = "=;Nil"
)

var logFilter string

func main() {
	logger := logging.NewLogger("nild")

	cfg := parseArgs(logger)

	logging.ApplyComponentsFilter(logFilter)

	profiling.Start(cfg.PprofPort)

	database, err := openDb(cfg.DB.Path, cfg.AllowDbDrop, logger)
	check.PanicIfErr(err)

	if len(cfg.ReadThrough.SourceAddr) != 0 {
		database, err = readthroughdb.NewReadThroughWithEndpoint(
			context.Background(),
			cfg.ReadThrough.SourceAddr,
			database,
			cfg.ReadThrough.ForkMainAtBlock)
		check.PanicIfErr(err)
	}

	exitCode := nilservice.Run(
		context.Background(),
		cfg.Config,
		database,
		nil,
		concurrent.MakeTask(
			"badger GC",
			func(ctx context.Context) error {
				return database.LogGC(ctx, cfg.DB.DiscardRatio, cfg.DB.GcFrequency)
			}))

	database.Close()
	os.Exit(exitCode)
}

func loadConfig() (*nildconfig.Config, error) {
	cfg := &nildconfig.Config{
		Config: nilservice.NewDefaultConfig(),
		DB:     db.NewDefaultBadgerDBOptions(),
		ReadThrough: &nildconfig.ReadThroughOptions{
			ForkMainAtBlock: transport.LatestBlockNumber,
		},
	}

	if err := cobrax.LoadConfigFromFile(cobrax.GetConfigNameFromArgs(), cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func addAllowDbClearFlag(fset *pflag.FlagSet, cfg *nildconfig.Config) {
	fset.BoolVar(
		&cfg.AllowDbDrop,
		"allow-db-clear",
		cfg.AllowDbDrop,
		"allow to clear database in case of outdated version")
}

func addRpcNodeFlags(fset *pflag.FlagSet, cfg *nildconfig.Config) {
	fset.Var(&cfg.RpcNode.ArchiveNodeList, "archive-nodes", "list of archive nodes")
}

func addBasicFlags(fset *pflag.FlagSet, cfg *nildconfig.Config) {
	fset.UintSliceVar(&cfg.MyShards, "my-shards", cfg.MyShards, "run only specified shard(s)")
	addAllowDbClearFlag(fset, cfg)
	fset.Uint32Var(
		&cfg.CollatorTickPeriodMs, "collator-tick-ms", cfg.CollatorTickPeriodMs, "collator tick period in milliseconds")
}

func doBootstrapRequestAndPatchConfig(
	ctx context.Context,
	bootstrapUrl string,
	cfg *nildconfig.Config,
	logger logging.Logger,
) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	bootstrapConfig, err := rpc_client.NewClient(bootstrapUrl, logger).GetBootstrapConfig(ctx)
	if err != nil {
		return err
	}

	// Patch config
	cfg.NShards = bootstrapConfig.NShards

	if cfg.ZeroState != nil {
		logger.Warn().Msg("overriding zero state config")
	}
	cfg.ZeroState = bootstrapConfig.ZeroStateConfig

	if cfg.Network.TcpPort == 0 {
		tcpPort := 3000 // Some port to work through libp2p
		logger.Info().Msgf("setting TCP port to %d", tcpPort)
		cfg.Network.TcpPort = tcpPort
	}

	cfg.BootstrapPeers = append(cfg.BootstrapPeers, bootstrapConfig.BootstrapPeers...)

	cfg.Network.DHTBootstrapPeers = append(cfg.Network.DHTBootstrapPeers, bootstrapConfig.DhtBootstrapPeers...)
	cfg.Network.DHTEnabled = true

	return nil
}

func parseArgs(logger logging.Logger) *nildconfig.Config {
	cfg, err := loadConfig()
	check.PanicIfErr(err)

	rootCmd := &cobra.Command{
		Use:           "nild [global flags] [command]",
		Short:         "nild cluster app",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cobrax.AddConfigFlag(rootCmd.PersistentFlags())

	var logLevel, libp2pLogLevel string
	cobrax.AddLogLevelFlag(rootCmd.PersistentFlags(), &logLevel)
	cobrax.AddCustomLogLevelFlag(rootCmd.PersistentFlags(), "libp2p-log-level", "", &libp2pLogLevel)

	rootCmd.PersistentFlags().StringVar(&cfg.DB.Path, "db-path", cfg.DB.Path, "path to database")
	rootCmd.PersistentFlags().Float64Var(
		&cfg.DB.DiscardRatio, "db-discard-ratio", cfg.DB.DiscardRatio, "discard ratio for badger GC")
	rootCmd.PersistentFlags().DurationVar(
		&cfg.DB.GcFrequency, "db-gc-interval", cfg.DB.GcFrequency, "frequency for badger GC")
	rootCmd.PersistentFlags().IntVar(&cfg.RPCPort, "http-port", cfg.RPCPort, "http port for rpc server")
	rootCmd.PersistentFlags().Var(
		&cfg.BootstrapPeers,
		"bootstrap-peers",
		"peers for snapshot fetching or transaction sending, must go in the order of shards")
	rootCmd.PersistentFlags().StringVar(
		&cfg.AdminSocketPath,
		"admin-socket-path",
		cfg.AdminSocketPath,
		"unix socket path to start admin server on (disabled if empty)")
	rootCmd.PersistentFlags().StringVar(
		&cfg.ReadThrough.SourceAddr,
		"read-through-db-addr",
		cfg.ReadThrough.SourceAddr,
		"address of the read-through database server. If provided, the local node will be run in read-through mode.")
	rootCmd.PersistentFlags().Var(
		&cfg.ReadThrough.ForkMainAtBlock,
		"read-through-fork-main-at-block",
		"all blocks generated later than this MainChain block won't be fetched; latest block by default")
	rootCmd.PersistentFlags().StringVar(
		&logFilter,
		"log-filter",
		"",
		"filter logs by component, e.g. 'all:-sync:-rpc' - enable all logs, but disable sync and rpc logs")

	cobrax.AddPprofPortFlag(rootCmd.PersistentFlags(), &cfg.PprofPort)

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run nil application server",
		Run: func(cmd *cobra.Command, args []string) {
			cfg.RunMode = nilservice.NormalRunMode
		},
	}
	runCmd.Flags().Uint32Var(&cfg.NShards, "nshards", cfg.NShards, "number of shardchains")
	runCmd.Flags().BoolVar(&cfg.SplitShards, "split-shards", cfg.SplitShards, "run each shard in separate process")
	runCmd.Flags().StringVar(&cfg.CometaConfig, "cometa-config", "", "path to Cometa config")
	runCmd.Flags().StringVar(
		&cfg.ValidatorKeysPath, "validator-keys-path", cfg.ValidatorKeysPath, "path to write validator keys")
	runCmd.Flags().BoolVar(&cfg.EnableDevApi, "dev-api", cfg.EnableDevApi, "enable development API")
	runCmd.Flags().StringVar(&cfg.IndexerConfig, "indexer-config", "", "path to Indexer config")

	addBasicFlags(runCmd.Flags(), cfg)
	cmdflags.AddNetwork(runCmd.Flags(), cfg.Network)
	cmdflags.AddTelemetry(runCmd.Flags(), cfg.Telemetry)

	replayCmd := &cobra.Command{
		Use:   "replay-block",
		Short: "Start server in single-shard mode to replay particular block",
		Run: func(cmd *cobra.Command, args []string) {
			cfg.RunMode = nilservice.BlockReplayRunMode
		},
	}
	replayCmd.Flags().Var(&cfg.Replay.BlockIdFirst, "first-block", "first block id to replay")
	replayCmd.Flags().Var(&cfg.Replay.BlockIdLast, "last-block", "last block id to replay")
	replayCmd.Flags().Var(&cfg.Replay.ShardId, "shard-id", "shard id to replay block from")

	var bootstrapUrl string
	archiveCmd := &cobra.Command{
		Use:   "archive",
		Short: "Run nil archive node",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.RunMode = nilservice.ArchiveRunMode

			if bootstrapUrl != "" {
				return doBootstrapRequestAndPatchConfig(cmd.Context(), bootstrapUrl, cfg, logger)
			}

			return nil
		},
	}

	addBasicFlags(archiveCmd.Flags(), cfg)
	cmdflags.AddNetwork(archiveCmd.Flags(), cfg.Network)
	cmdflags.AddTelemetry(archiveCmd.Flags(), cfg.Telemetry)
	// N.B. Despite the fact that we override the config with the loaded configuration,
	// we still handle this flag in the usual way rather than resorting to a hack like GetConfigNameFromArgs.
	// Network-supplied values should win by default, but because we merge only a small, well-defined
	// subset of settings, any rare conflicts can be resolved explicitly and safely.
	archiveCmd.Flags().StringVar(
		&bootstrapUrl, "bootstrap-url", "", "url to fetch initial configuration from (genesis config etc.")

	rpcCmd := &cobra.Command{
		Use:   "rpc",
		Short: "Run nil rpc server",
		Run: func(cmd *cobra.Command, args []string) {
			cfg.RunMode = nilservice.RpcRunMode
		},
	}
	rpcCmd.Flags().BoolVar(&cfg.EnableDevApi, "dev-api", cfg.EnableDevApi, "enable development API")

	addRpcNodeFlags(rpcCmd.Flags(), cfg)
	addAllowDbClearFlag(rpcCmd.Flags(), cfg)
	cmdflags.AddNetwork(rpcCmd.Flags(), cfg.Network)
	cmdflags.AddTelemetry(rpcCmd.Flags(), cfg.Telemetry)

	versionCmd := cobrax.VersionCmd(appTitle)
	devnetCmd := DevnetCommand()

	rootCmd.AddCommand(runCmd, replayCmd, archiveCmd, rpcCmd, devnetCmd, versionCmd)
	cobrax.ExitOnHelp(rootCmd)

	check.PanicIfErr(rootCmd.Execute())

	logging.SetupGlobalLogger(logLevel)
	check.PanicIfErr(logging.SetLibp2pLogLevel(libp2pLogLevel))

	if cfg.Replay.BlockIdLast == 0 {
		cfg.Replay.BlockIdLast = cfg.Replay.BlockIdFirst
	}

	if cfg.CometaConfig != "" {
		cfg.Cometa = &cometa.Config{}
		cfg.Cometa.ResetToDefault()
		ok := cfg.Cometa.InitFromFile(cfg.CometaConfig)
		check.PanicIfNotf(ok, "failed to load cometa config from %s", cfg.CometaConfig)
	} else if cfg.ShouldStartCometa() {
		cfg.Cometa = &cometa.Config{}
		cfg.Cometa.ResetToDefault()
		cfg.Cometa.UseBadger = true
	}

	if cfg.IndexerConfig != "" {
		cfg.Indexer = &indexer.Config{}
		cfg.Indexer.ResetToDefault()
		ok := cfg.Indexer.InitFromFile(cfg.IndexerConfig)
		check.PanicIfNotf(ok, "failed to load indexer config from %s", cfg.IndexerConfig)
	}

	return cfg
}

func openDb(dbPath string, allowDrop bool, logger logging.Logger) (db.DB, error) {
	dbExists := true
	if _, err := os.Open(dbPath); err != nil {
		if !os.IsNotExist(err) {
			logger.Error().Err(err).Msg("Error opening db path")
			return nil, err
		}
		dbExists = false
	}

	// each shard will interact with DB via this client
	badger, err := db.NewBadgerDb(dbPath)
	if err != nil {
		return nil, err
	}

	tx, err := badger.CreateRwTx(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	logger.Info().Msg("Checking scheme format...")
	isVersionOutdated, err := db.IsVersionOutdated(tx)
	if err != nil {
		return nil, err
	}

	if isVersionOutdated {
		if !allowDrop {
			return nil, errors.New("database schema is outdated; remove database or use --allow-db-clear")
		}

		logger.Info().Msg("Clearing database from old data...")
		if err := badger.DropAll(); err != nil {
			return nil, err
		}
	}

	if !dbExists || isVersionOutdated {
		if err := db.WriteVersionInfo(tx, types.NewVersionInfo()); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	}

	return badger, nil
}
