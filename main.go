package main

import (
	"github.com/spf13/cobra"
)

func main() {

	rootCmd := cobra.Command{
		Use:   "extism",
		Short: "A CLI for Extism plugins",
	}

	rootCmd.AddCommand(callCmd())
	rootCmd.AddCommand(libCmd())
	rootCmd.Execute()
}
