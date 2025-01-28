package contract

import (
	"fmt"

	"github.com/NilFoundation/nil/nil/cmd/nil/common"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/cliservice"
	"github.com/spf13/cobra"
)

func GetEstimateFeeCommand(cfg *common.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "estimate-fee [address] [calldata or method] [args...]",
		Short: "Get the recommended fee credit for a transaction",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEstimateFee(cmd, args, cfg)
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringVar(&params.AbiPath, abiFlag, "", "The path to the ABI file")
	cmd.Flags().Var(&params.value, valueFlag, "The value for transfer")
	cmd.Flags().BoolVar(&params.internal, internalFlag, false, "Set the \"internal\" flag")
	cmd.Flags().BoolVar(&params.deploy, deployFlag, false, "Set the \"deploy\" flag")

	return cmd
}

func runEstimateFee(cmd *cobra.Command, args []string, cfg *common.Config) error {
	service := cliservice.NewService(cmd.Context(), common.GetRpcClient(), cfg.PrivateKey, nil)

	var address types.Address
	if err := address.Set(args[0]); err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}

	abi, err := common.ReadAbiFromFile(params.AbiPath)
	if err != nil {
		return err
	}

	calldata, err := common.PrepareArgs(abi, args[1], args[2:])
	if err != nil {
		return err
	}

	var txnFlags types.TransactionFlags
	if params.internal {
		txnFlags.SetBit(types.TransactionFlagInternal)
	}
	if params.deploy {
		txnFlags.SetBit(types.TransactionFlagDeploy)
	}

	res, err := service.EstimateFee(address, calldata, txnFlags, params.value)
	if err != nil {
		return err
	}

	if !common.Quiet {
		fmt.Print("Estimated fee: ")
	}
	fmt.Println(res)

	return nil
}
