package main

import (
	"os"

	"github.com/extism/cli"
	"github.com/spf13/cobra"
)

var banner string = `
███████╗██╗░░██╗████████╗██╗░██████╗███╗░░░███╗
██╔════╝╚██╗██╔╝╚══██╔══╝██║██╔════╝████╗░████║
█████╗░░░╚███╔╝░░░░██║░░░██║╚█████╗░██╔████╔██║
██╔══╝░░░██╔██╗░░░░██║░░░██║░╚═══██╗██║╚██╔╝██║
███████╗██╔╝╚██╗░░░██║░░░██║██████╔╝██║░╚═╝░██║
╚══════╝╚═╝░░╚═╝░░░╚═╝░░░╚═╝╚═════╝░╚═╝░░░░░╚═╝
`

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "extism",
		Version: "0.3.4",
		Long:    banner,
		Short:   "A CLI for Extism, https://extism.org",
	}
	cmd.PersistentFlags().BoolVarP(&cli.LoggingEnabled, "verbose", "v", false, "Enable additional logging")
	cmd.PersistentFlags().BoolVarP(&cli.PrintingDisabled, "quiet", "q", false, "Enable additional logging")
	cmd.PersistentFlags().StringVar(&cli.GithubToken, "github-token", os.Getenv("GITHUB_TOKEN"),
		"Github access token, can also be set using the $GITHUB_TOKEN env variable")
	cmd.AddCommand(cli.CallCmd())
	cmd.AddCommand(cli.LibCmd())
	return cmd
}

func main() {
	err := rootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
