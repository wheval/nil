package smartaccount

import (
	"fmt"

	"github.com/NilFoundation/nil/nil/cmd/nil/common"
	"github.com/NilFoundation/nil/nil/services/cliservice"
	"github.com/spf13/cobra"
)

func BalanceCommand(cfg *common.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance",
		Short: "Get the balance of the smart account whose address specified in config.address field",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBalance(cmd, args, cfg)
		},
		SilenceUsage: true,
	}

	return cmd
}

func runBalance(cmd *cobra.Command, _ []string, cfg *common.Config) error {
	service := cliservice.NewService(cmd.Context(), common.GetRpcClient(), cfg.PrivateKey, nil)
	balance, err := service.GetBalance(cfg.Address)
	if err != nil {
		return err
	}
	if !common.Quiet {
		fmt.Print("Smart account balance: ")
	}
	fmt.Println(balance)
	return nil
}
