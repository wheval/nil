package minter

import (
	"github.com/NilFoundation/nil/nil/cmd/nil/internal/common"
	"github.com/spf13/cobra"
)

func GetCommand(cfg *common.Config) *cobra.Command {
	serverCmd := &cobra.Command{
		Use:   "minter",
		Short: "Interact with the minter on the cluster",
	}

	serverCmd.AddCommand(CreateTokenCommand(cfg))
	serverCmd.AddCommand(ChangeTokenAmountCommand(cfg, true))
	serverCmd.AddCommand(ChangeTokenAmountCommand(cfg, false))

	return serverCmd
}
