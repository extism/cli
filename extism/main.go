package main

import (
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
		Version: "0.2.0",
		Long:    banner,
		Short:   "A CLI for Extism, https://extism.org",
	}
	cmd.AddCommand(cli.CallCmd())
	cmd.AddCommand(cli.LibCmd())
	return cmd
}

func main() {
	rootCmd().Execute()
}
