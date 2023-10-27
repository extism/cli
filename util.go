package cli

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/spf13/cobra"
)

var LoggingEnabled = false
var PrintingDisabled = false
var GithubToken = ""

type Args interface {
	SetArgs(args []string)
}

func RunArgs[T Args](f func(cmd *cobra.Command, args T) error, call T) func(cmd *cobra.Command, args []string) error {
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

func getStaticLibFileName(os string) string {
	if os == "windows" || os == "windows-gnu" {
		return "extism.lib"
	} else {
		return "libextism.a"
	}
}

func copyFile(src string, dest string) error {
	Print("Copying", src, "to", dest)
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
