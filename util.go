package cli

import (
	"fmt"
	"io/ioutil"
	"runtime"

	"github.com/spf13/cobra"
)

type Args interface {
	SetArgs(args []string)
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

func copyFile(src string, dest string) error {
	fmt.Println("Copying", src, "to", dest)
	bytesRead, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dest, bytesRead, 0755)
	if err != nil {
		return err
	}

	return nil
}
