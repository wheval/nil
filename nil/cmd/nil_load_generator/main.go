package main

import (
	"context"
	"time"

	"github.com/NilFoundation/nil/nil/common/check"
	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/internal/telemetry"
	"github.com/NilFoundation/nil/nil/services/nil_load_generator"
	"github.com/spf13/cobra"
)

func main() {
	cfg := &nil_load_generator.Config{}
	componentName := "nil_load_generator"
	logger := logging.NewLogger(componentName)
	rootCmd := &cobra.Command{
		Use:   componentName,
		Short: "Run nil load generator",
	}

	rootCmd.Flags().StringVar(&cfg.Endpoint, "endpoint", "http://127.0.0.1:8529/", "rpc endpoint")
	rootCmd.Flags().StringVar(&cfg.OwnEndpoint, "own-endpoint", "tcp://127.0.0.1:8525", "own rpc endpoint")
	rootCmd.Flags().StringVar(&cfg.FaucetEndpoint, "faucet-endpoint", "tcp://127.0.0.1:8527", "faucet rpc endpoint")
	rootCmd.Flags().Uint32Var(&cfg.CheckBalance, "check-balance", 10, "frequency of balance check in iterations")
	rootCmd.Flags().Uint32Var(&cfg.SwapPerIteration, "swap-per-iteration", 1000, "amount of swaps per iteration")
	rootCmd.Flags().BoolVar(&cfg.Metrics, "metrics", false, "export metrics via grpc")
	rootCmd.Flags().StringVar(&cfg.MintTokenAmount0, "mint-token-amount0", "3000000000000000", "mint amount for token0")
	rootCmd.Flags().StringVar(&cfg.MintTokenAmount1, "mint-token-amount1", "10000000000000", "mint amount for token1")
	rootCmd.Flags().StringVar(&cfg.ThresholdAmount, "threshold-amount", "3000000000000000000", "threshold amount")
	rootCmd.Flags().StringVar(&cfg.SwapAmount, "swap-amount", "1000", "swap amount")
	rootCmd.Flags().StringVar(&cfg.RpcSwapLimit, "rpc-swap-limit", "1000000", "rpc swap limit")
	rootCmd.Flags().Uint32Var(&cfg.UniswapAccounts, "rpc-uniswap-accounts", 5, "number of uniswap accounts")
	rootCmd.Flags().StringVar(&cfg.LogLevel, "log-level", "info", "log level: trace|debug|info|warn|error|fatal|panic")
	rootCmd.Flags().DurationVar(&cfg.WaitClusterStartup, "wait-cluster-startup", 5*time.Minute, "time to wait for cluster startup")

	check.PanicIfErr(rootCmd.Execute())

	if err := telemetry.Init(context.Background(), &telemetry.Config{ServiceName: componentName, ExportMetrics: cfg.Metrics}); err != nil {
		logger.Err(err).Send()
		panic("Can't init telemetry")
	}
	defer telemetry.Shutdown(context.Background())

	if err := nil_load_generator.Run(context.Background(), *cfg, logger); err != nil {
		logger.Error().Err(err).Msg("Error during nil load generator run")
		panic(err)
	}
}
