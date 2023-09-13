package main

import (
	"errors"

	"github.com/spf13/cobra"
)

type libArgs struct {
	args   []string
	prefix string
}

type libInstallArgs struct {
	libArgs
	version string
}

type libUninstallArgs struct {
	libArgs
}

func (a *libInstallArgs) SetArgs(args []string) {
	a.args = args
}

func (a *libInstallArgs) GetArgs() []string {
	return a.args
}

func (a *libUninstallArgs) SetArgs(args []string) {
	a.args = args
}

func (a *libUninstallArgs) GetArgs() []string {
	return a.args
}

func runLibInstall(cmd *cobra.Command, installArgs *libInstallArgs) error {
	return errors.New("TODO")
}

func runLibUninstall(cmd *cobra.Command, installArgs *libUninstallArgs) error {
	return errors.New("TODO")
}

func libCmd() *cobra.Command {
	lib := &cobra.Command{
		Use:   "lib",
		Short: "Manage libextism",
	}

	// Install
	installArgs := &libInstallArgs{}
	libInstall := &cobra.Command{
		Use:   "install",
		Short: "Install libextism",
		RunE:  runArgs(runLibInstall, installArgs),
	}
	libInstall.Flags().StringVar(&installArgs.version, "version", "latest",
		"Install a specified Extism version, `latest` can be used to specify the latest release and `git` can be used to install from git")
	libInstall.Flags().StringVar(&installArgs.prefix, "prefix", "/usr/local",
		"Prefix to install libextism and extism.h into, the shared object will be copied to $PREFIX/lib and the header will be copied to $PREFIX/include")
	lib.AddCommand(libInstall)

	// Uninstall
	uninstallArgs := &libUninstallArgs{}
	libUninstall := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall libextism",
		RunE:  runArgs(runLibUninstall, uninstallArgs),
	}
	libUninstall.Flags().StringVar(&uninstallArgs.prefix, "prefix", "/usr/local", "Prefix previously to used to install libextism")
	lib.AddCommand(libUninstall)

	return lib
}
