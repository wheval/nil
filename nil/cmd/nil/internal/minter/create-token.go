package minter

import (
	"fmt"

	"github.com/NilFoundation/nil/nil/cmd/nil/common"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/cliservice"
	"github.com/spf13/cobra"
)

func CreateTokenCommand(cfg *common.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-token [address] [amount] [name]",
		Short: "Create a custom token",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreateToken(cmd, args, cfg)
		},
		SilenceUsage: true,
	}

	return cmd
}

func runCreateToken(cmd *cobra.Command, args []string, cfg *common.Config) error {
	service := cliservice.NewService(cmd.Context(), common.GetRpcClient(), cfg.PrivateKey, nil)

	var address types.Address
	if err := address.Set(args[0]); err != nil {
		return err
	}

	var amount types.Value
	if err := amount.Set(args[1]); err != nil {
		return err
	}

	name := args[2]

	tokenId, err := service.TokenCreate(address, amount, name)
	if err != nil {
		return err
	}
	if !common.Quiet {
		fmt.Print("Created Token ID: ")
	}
	fmt.Println(tokenId)
	return nil
}
