package main

import (
	"context"
	"fmt"
	"os"

	"github.com/NilFoundation/nil/nil/services/stresser"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "stresser --config <config-file>",
		Short: "Run stresser",
	}

	var configFile string
	rootCmd.Flags().StringVar(&configFile, "config", "", "config file")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	st, err := stresser.NewStresserFromFile(configFile)
	if err != nil {
		fmt.Println("Failed to create stresser:", err)
		os.Exit(1)
	}
	if err = st.Run(context.Background()); err != nil {
		fmt.Println("Failed to run stresser:", err)
		os.Exit(1)
	}
}
