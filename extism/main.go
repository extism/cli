package main

import (
	_ "embed"
	"os"
	"strings"

	shell "github.com/brianstrauch/cobra-shell"
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

//go:embed VERSION
var version string

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "extism",
		Version: strings.TrimSpace(version),
		Long:    banner,
		Short:   "A CLI for Extism, https://extism.org",
	}
	cmd.PersistentFlags().BoolVarP(&cli.LoggingEnabled, "verbose", "v", false, "Enable additional logging")
	cmd.PersistentFlags().BoolVarP(&cli.PrintingDisabled, "quiet", "q", false, "Suppress output")
	cmd.PersistentFlags().StringVar(&cli.GithubToken, "github-token", os.Getenv("GITHUB_TOKEN"),
		"Github access token, can also be set using the $GITHUB_TOKEN env variable")
	cmd.AddCommand(cli.CallCmd())
	cmd.AddCommand(cli.LibCmd())
	cmd.AddCommand(cli.GenerateCmd())
	cmd.AddCommand(shell.New(cmd, nil))
	return cmd
}

func main() {
	err := rootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
