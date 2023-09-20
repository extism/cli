package cli

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/spf13/cobra"
)

type Args interface {
	SetArgs(args []string)
}

func runArgs[T Args](f func(cmd *cobra.Command, args T) error, call T) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error { call.SetArgs(args); return f(cmd, call) }
}

func getSharedObjectExt(os string) string {
	if os == "darwin" || os == "macos" {
		return "dylib"
	} else if os == "windows" || os == "windows-gnu" {
		return "dll"
	} else {
		return "so"
	}
}

func getSharedObjectFileName(os string) string {
	if os == "windows" || os == "windows-gnu" {
		return "extism." + getSharedObjectExt(os)
	} else {
		return "libextism." + getSharedObjectExt(os)
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

var LoggingEnabled = false
var PrintingDisabled = false

func Log(s ...any) {
	if LoggingEnabled {
		log.Println(s...)
	}
}

func Print(s ...any) {
	if LoggingEnabled {
		Log(s...)
	} else if !PrintingDisabled {
		fmt.Println(s...)
	}
}
