package transaction

import (
	"fmt"

	"github.com/NilFoundation/nil/nil/cmd/nil/internal/common"
	"github.com/NilFoundation/nil/nil/cmd/nil/internal/config"
	libcommon "github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/services/cliservice"
	"github.com/spf13/cobra"
)

var logger = logging.NewLogger("transactionCommand")

func GetCommand(cfgPath *string) *cobra.Command {
	serverCmd := &cobra.Command{
		Use:   "transaction [hash]",
		Short: "Retrieve a transaction from the cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommand(cmd, cfgPath, args)
		},
		SilenceUsage: true,
	}

	serverCmd.AddCommand(GetInternalTransactionCommand())

	return serverCmd
}

func runCommand(cmd *cobra.Command, cfgPath *string, args []string) error {
	cfg, err := config.LoadConfig(*cfgPath, logger)
	if err != nil {
		return err
	}
	common.InitRpcClient(cfg, logger)

	service := cliservice.NewService(cmd.Context(), common.GetRpcClient(), nil, nil)

	var hash libcommon.Hash
	if err := hash.Set(args[0]); err != nil {
		return err
	}

	if hash != libcommon.EmptyHash {
		txnDataJson, err := service.FetchTransactionByHashJson(hash)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to fetch the transaction")
			return err
		}
		if !common.Quiet {
			fmt.Print("Transaction data: ")
		}
		fmt.Println(string(txnDataJson))
	}
	return nil
}
