package main

import (
	"github.com/spf13/cobra"
)

type Args interface {
	SetArgs(args []string)
	GetArgs() []string
}

func runArgs[T Args](f func(cmd *cobra.Command, args T) error, call T) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error { call.SetArgs(args); return f(cmd, call) }
}
