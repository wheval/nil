package contract

import (
	"fmt"

	"github.com/NilFoundation/nil/nil/cmd/nil/common"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/cliservice"
	"github.com/spf13/cobra"
)

func GetTokensCommand(cfg *common.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tokens [address]",
		Short: "Get the tokens held by a smart contract as a map tokenId -> balance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTokens(cmd, args, cfg)
		},
	}

	return cmd
}

func runTokens(cmd *cobra.Command, args []string, cfg *common.Config) error {
	var address types.Address
	if err := address.Set(args[0]); err != nil {
		return err
	}

	service := cliservice.NewService(cmd.Context(), common.GetRpcClient(), cfg.PrivateKey, nil)
	tokens, err := service.GetTokens(address)
	if err != nil {
		return err
	}
	if !common.Quiet {
		fmt.Println("Contract tokens:")
	}
	for k, v := range tokens {
		fmt.Printf("%s\t%s", k, v)
		if name := types.GetTokenName(k); len(name) > 0 && !common.Quiet {
			fmt.Printf("\t[%s]", name)
		}
		fmt.Println()
	}
	return nil
}
