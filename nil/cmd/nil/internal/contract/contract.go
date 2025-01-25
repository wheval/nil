package contract

import (
	"github.com/NilFoundation/nil/nil/cmd/nil/internal/common"
	"github.com/spf13/cobra"
)

func GetCommand(cfg *common.Config) *cobra.Command {
	serverCmd := &cobra.Command{
		Use:   "contract",
		Short: "Interact with a contract on the cluster",
	}

	serverCmd.AddCommand(
		GetAddressCommand(cfg),
		GetBalanceCommand(cfg),
		GetTokensCommand(cfg),
		GetCodeCommand(cfg),
		GetCallReadonlyCommand(cfg),
		GetDeployCommand(cfg),
		GetSendExternalTransactionCommand(cfg),
		GetEstimateFeeCommand(cfg),
		GetTopUpCommand(cfg),
		GetSeqnoCommand(),
	)

	return serverCmd
}
