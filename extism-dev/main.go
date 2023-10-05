package main

import (
	"os"

	"github.com/extism/cli"
	"github.com/spf13/cobra"
)

var banner string = `
____    
|            
|___ \ / |_ o  ___             |  __
|     x  |  | |___  |\/| __  __| |__| \  /
|___ / \ |_ |  ___| |  |    |__| |___  \/
`

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "extism-dev",
		Version: "0.0.0",
		Long:    banner,
		Short:   "The Extism repo manager, https://extism.org",
	}
	cmd.PersistentFlags().BoolVarP(&cli.LoggingEnabled, "verbose", "v", false, "Enable additional logging")
	cmd.PersistentFlags().BoolVarP(&cli.PrintingDisabled, "quiet", "q", false, "Enable additional logging")
	if err := SetupDevCmd(cmd); err != nil {
		panic(err)
	}
	return cmd
}

func main() {
	err := rootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
