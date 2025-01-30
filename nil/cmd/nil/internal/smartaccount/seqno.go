package smartaccount

import (
	"fmt"

	"github.com/NilFoundation/nil/nil/cmd/nil/common"
	"github.com/NilFoundation/nil/nil/services/cliservice"
	"github.com/spf13/cobra"
)

func SeqnoCommand(cfg *common.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seqno",
		Short: "Get the seqno of the smart account whose address specified in config.address field",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSeqno(cmd, args, cfg)
		},
		SilenceUsage: true,
	}

	return cmd
}

func runSeqno(cmd *cobra.Command, _ []string, cfg *common.Config) error {
	service := cliservice.NewService(cmd.Context(), common.GetRpcClient(), cfg.PrivateKey, nil)
	seqno, err := service.GetSeqno(cfg.Address)
	if err != nil {
		return err
	}
	if !common.Quiet {
		fmt.Print("Smart account seqno: ")
	}
	fmt.Println(seqno)
	return nil
}
