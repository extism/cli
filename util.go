package main

import (
	"runtime"

	"github.com/spf13/cobra"
)

type Args interface {
	SetArgs(args []string)
	GetArgs() []string
}

func runArgs[T Args](f func(cmd *cobra.Command, args T) error, call T) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error { call.SetArgs(args); return f(cmd, call) }
}

func getSharedObjectExt() string {
	if runtime.GOOS == "macos" {
		return "dylib"
	} else if runtime.GOOS == "windows" {
		return "dll"
	} else {
		return "so"
	}
}

func getSharedObjectFileName() string {
	if runtime.GOOS == "windows" {
		return "extism." + getSharedObjectExt()
	} else {
		return "libextism." + getSharedObjectExt()
	}
}
