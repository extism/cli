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

func main() {
	rootCmd := cobra.Command{
		Use:     "extism",
		Version: "0.2.0",
		Long:    banner,
		Short:   "A CLI for Extism, https://extism.org",
	}

	rootCmd.AddCommand(cli.CallCmd())
	rootCmd.AddCommand(cli.LibCmd())
	rootCmd.Execute()
}