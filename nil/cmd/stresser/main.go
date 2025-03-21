package main

import (
	"context"
	"fmt"
	"os"

	"github.com/NilFoundation/nil/nil/services/stresser"
	"github.com/spf13/cobra"
)

func main() {
	var configFile string
	rootCmd := &cobra.Command{
		Use:   "stresser --config <config-file>",
		Short: "Run stresser",
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := stresser.NewStresserFromFile(configFile)
			if err != nil {
				return fmt.Errorf("Failed to create stresser: %w", err)
			}
			if err = st.Run(context.Background()); err != nil {
				return fmt.Errorf("Failed to run stresser: %w", err)
			}
			return nil
		},
	}

	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
