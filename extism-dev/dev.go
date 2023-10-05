package main

import (
	"os"
	"path/filepath"

	"github.com/extism/cli"
	"github.com/spf13/cobra"
)

type extismData struct {
	Repos []repo `json:"repos"`
}

func homeDir() string {
	d, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return d
}

func getDefaultRoot() (string, error) {
	defaultRoot := os.Getenv("EXTISM_DEV_ROOT")
	if defaultRoot == "" {
		link := filepath.Join(homeDir(), ".extism.dev")
		cli.Log("Checking", link)
		path, err := os.Readlink(link)
		if err == nil {
			defaultRoot = path
		}
		defaultRoot = path
	}
	return defaultRoot, nil
}

var Root string = ""

func SetupDevCmd(dev *cobra.Command) error {
	defaultRoot, err := getDefaultRoot()
	if err != nil {
		return err
	}
	dev.PersistentFlags().StringVar(&Root, "root", defaultRoot, "Root of extism development repos, all packages will be cloned into directories matching their github URLs inside this directory")

	// Init
	initArgs := &devInitArgs{}
	devInit := &cobra.Command{
		Use:          "init [flags] dev_root",
		Short:        "Initialize dev repos",
		SilenceUsage: true,
		RunE:         cli.RunArgs(runDevInit, initArgs),
		Args:         cobra.NoArgs,
	}
	devInit.Flags().IntVarP(&initArgs.parallel, "parallel", "p", 4, "Number of repos to download in parallel")
	devInit.Flags().BoolVar(&initArgs.local, "local", false, "Do not set as global extism-dev path")
	devInit.MarkPersistentFlagRequired("root")
	dev.AddCommand(devInit)

	// Exec
	execArgs := &devExecArgs{}
	devExec := &cobra.Command{
		Use:          "exec",
		Short:        "Run a command in each repo",
		SilenceUsage: true,
		RunE:         cli.RunArgs(runDevExec, execArgs),
		Args:         cobra.ExactArgs(1),
	}

	defaultShell := os.Getenv("SHELL")
	if defaultShell == "" {
		defaultShell = "sh"
	}
	devExec.Flags().StringVarP(&execArgs.category, "category", "c", "", "Category: sdk, pdk, plugin, runtime or other")
	devExec.Flags().StringVarP(&execArgs.repo, "repo", "r", "", "Regex filter used on the repo name")
	devExec.Flags().StringVarP(&execArgs.shell, "shell", "s", defaultShell, "Shell to use when executing commands")
	devExec.Flags().IntVarP(&execArgs.parallel, "parallel", "p", 1, "Number of commands to execute in parallel")
	dev.AddCommand(devExec)

	// Find
	findArgs := &devFindArgs{}
	devFind := &cobra.Command{
		Use:          "find [flags] pattern",
		Short:        "Search for files across all repositories",
		SilenceUsage: true,
		RunE:         cli.RunArgs(runDevFind, findArgs),
	}
	defaultEditor := os.Getenv("EDITOR")
	if defaultEditor == "" {
		defaultEditor = "/usr/bin/editor"
	}
	devFind.Flags().StringVarP(&findArgs.category, "category", "c", "", "Category: sdk, pdk, plugin, runtime or other")
	devFind.Flags().StringVarP(&findArgs.repo, "repo", "r", "", "Regex filter used on the repo name")
	devFind.Flags().StringVar(&findArgs.filename, "filename", "", "Filter for filenames")
	devFind.Flags().StringVar(&findArgs.replace, "replace", "", "Replacement string")
	devFind.Flags().StringVar(&findArgs.editor, "editor", defaultEditor, "Editor command")
	devFind.Flags().BoolVar(&findArgs.edit, "edit", false, "Edit matching files")
	devFind.Flags().BoolVarP(&findArgs.interactive, "interactive", "i", false, "Prompt before editing or replacing")
	dev.AddCommand(devFind)

	// Add
	addArgs := &devAddArgs{}
	devAdd := &cobra.Command{
		Use:          "add [flags] url",
		Short:        "Add a repo",
		SilenceUsage: true,
		RunE:         cli.RunArgs(runDevAdd, addArgs),
		Args:         cobra.ExactArgs(1),
	}
	devAdd.Flags().StringVarP(&addArgs.category, "category", "c", "other", "Category: sdk, pdk, plugin, runtime or other")
	dev.AddCommand(devAdd)

	// Remove
	removeArgs := &devRemoveArgs{}
	devRemove := &cobra.Command{
		Use:          "remove [flags] url",
		Aliases:      []string{"rm"},
		Short:        "Remove a repo",
		SilenceUsage: true,
		RunE:         cli.RunArgs(runDevRemove, removeArgs),
		Args:         cobra.ExactArgs(1),
	}
	devRemove.Flags().BoolVar(&removeArgs.keep, "keep", false, "Don't remove directory after removing repo")
	dev.AddCommand(devRemove)

	// Path
	devPath := &cobra.Command{
		Use:          "path",
		Short:        "Print the initialized global extism-dev path",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli.Print(defaultRoot)
			return nil
		},
	}
	dev.AddCommand(devPath)

	// List
	listArgs := &devListArgs{}
	devList := &cobra.Command{
		Use:          "list",
		Short:        "List repos and their paths",
		SilenceUsage: true,
		RunE:         cli.RunArgs(runDevList, listArgs),
	}
	devList.Flags().StringVarP(&listArgs.category, "category", "c", "", "Category: sdk, pdk, plugin, runtime or other")
	dev.AddCommand(devList)

	// Update
	updateArgs := &devUpdateArgs{}
	devUpdate := &cobra.Command{
		Use:          "update",
		Short:        "Common batch updates",
		SilenceUsage: true,
		RunE:         cli.RunArgs(runDevUpdate, updateArgs),
	}
	devUpdate.Flags().BoolVar(&updateArgs.kernel, "kernel", false, "Update kernel files across repos")
	dev.AddCommand(devUpdate)

	return nil
}
