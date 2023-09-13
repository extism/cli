package main

import (
	"github.com/extism/cli"
	"github.com/spf13/cobra"
)

func main() {

	rootCmd := cobra.Command{
		Use:   "extism",
		Short: "A CLI for Extism plugins",
	}

	rootCmd.AddCommand(cli.CallCmd())
	rootCmd.AddCommand(cli.LibCmd())
	rootCmd.Execute()
}
