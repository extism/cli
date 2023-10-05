package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/extism/cli"
	"github.com/spf13/cobra"
)

type extismData struct {
	Repos []repo `json:"repos"`
}

type devInitArgs struct {
	devArgs
	parallel int
	noLink   bool
}

func runDevInit(cmd *cobra.Command, args *devInitArgs) error {
	args.root = args.args[0]
	data, err := args.loadDataFile()
	if err != nil {
		data = &extismData{
			Repos: defaultRepos,
		}
	} else {
		args.mergeRepos(data)
	}

	cli.Print("Initializing Extism dev repos in", args.root)
	err = os.MkdirAll(args.root, 0o755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	pool := NewPool(args.parallel)
	for _, r := range data.Repos {
		cli.Log("Repos", data.Repos)
		RunTask(pool, func(repo repo) {
			repo.clone(args.root)
		}, r)
	}
	pool.Wait()

	if err := args.link(); err != nil {
		return err
	}

	if err := args.saveDataFile(data); err != nil {
		return err
	}

	return nil
}

type devExecArgs struct {
	devArgs
	category string
	repo     string
	shell    string
}

func runDevExec(cmd *cobra.Command, args *devExecArgs) error {
	data, err := args.loadDataFile()
	if err != nil {
		return err
	}

	rx := &regexp.Regexp{}
	if args.repo != "" {
		rx = regexp.MustCompile(args.repo)
	}

	for _, repo := range data.Repos {
		if args.category != "" && repo.Category.String() != args.category {
			continue
		}

		if args.repo != "" {
			if !rx.MatchString(repo.Url) {
				continue
			}
		}
		userName, repoName := repo.split()
		p := filepath.Join(args.root, userName, repoName)
		cli.Log("Executing", args.args[0], "in", p, "using", args.shell)
		cli.Print()
		cli.Print(p)
		cmd := exec.Command(args.shell, "-c", args.args[0])
		cmd.Dir = p
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "EXTISM_DEV_ROOT="+args.root)
		cmd.Env = append(cmd.Env, "EXTISM_DEV_RUNTIME="+filepath.Join(args.root, "extism", "extism"))
		cmd.Env = append(cmd.Env, "EXTISM_DEV_REPO="+repo.Url)
		cmd.Env = append(cmd.Env, "EXTISM_DEV_CATEGORY"+repo.Category.String())
		cmd.Env = append(cmd.Env, "PATH="+os.Getenv("PATH")+":"+filepath.Join(args.root, ".bin"))
		if err := cmd.Run(); err != nil {
			cli.Print("Error: command failed in", p)
		}
	}
	return nil
}

type devFindArgs struct {
	devArgs
	category     string
	onlyFilename bool
	filename     string
	edit         bool
	editor       string
	repo         string
}

func runDevFind(cmd *cobra.Command, args *devFindArgs) error {
	rg, err := exec.LookPath("rg")
	if err != nil {
		return errors.New("unable to find `rg` executable, install using `cargo install ripgrep`")
	}

	if args.edit {
		_, err := exec.LookPath(args.editor)
		if err != nil {
			return errors.New("editor not found: " + args.editor)
		}
	}

	cmdArgs := []string{"--color", "never", "--files-with-matches"}

	if args.filename != "" {
		cmdArgs = append(cmdArgs, "-g", args.filename)

		if args.onlyFilename {
			cmdArgs = append(cmdArgs, "--files")
		}
	}

	cmdArgs = append(cmdArgs, args.args[0])

	data, err := args.loadDataFile()
	if err != nil {
		return err
	}

	rx := &regexp.Regexp{}
	if args.repo != "" {
		rx = regexp.MustCompile(args.repo)
	}

	for _, repo := range data.Repos {
		if args.category != "" && repo.Category.String() != args.category {
			continue
		}

		if args.repo != "" {
			if !rx.MatchString(repo.Url) {
				continue
			}
		}
		userName, repoName := repo.split()
		p := filepath.Join(args.root, userName, repoName)
		a := cmdArgs[:]
		a = append(a, p)
		cmd := exec.Command(rg, a...)
		if !args.edit {
			cmd.Stdout = os.Stdout
		}
		// cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "EXTISM_DEV_ROOT="+args.root)
		cmd.Env = append(cmd.Env, "EXTISM_DEV_RUNTIME="+filepath.Join(args.root, "extism", "extism"))
		cmd.Env = append(cmd.Env, "EXTISM_DEV_REPO="+repo.Url)
		cmd.Env = append(cmd.Env, "EXTISM_DEV_CATEGORY"+repo.Category.String())
		cmd.Env = append(cmd.Env, "PATH="+os.Getenv("PATH")+":"+filepath.Join(args.root, ".bin"))
		if !args.edit {
			if err := cmd.Run(); err != nil {
				cli.Log("rg returned non-zero exit code in", p+":", err)
			}
		} else {
			output, err := cmd.Output()
			if err != nil {
				cli.Log("rg returned non-zero exit code in", p+":", err)
				continue
			}

			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if len(line) == 0 {
					continue
				}
				cli.Print(line)
				cmd := exec.Command(args.editor, line)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Stdin = os.Stdin
				cmd.Run()
			}
		}
	}
	return nil
}

type devAddArgs struct {
	devArgs
	url      string
	category string
}

func runDevAdd(cmd *cobra.Command, args *devAddArgs) error {
	data, err := args.loadDataFile()
	if err != nil {
		return err
	}

	r := repo{
		Url: args.url,
	}
	r.Category.Parse(args.category)
	for _, s := range data.Repos {
		if s.Url == r.Url {
			cli.Print("Repo already exists, not adding")
			return nil
		}
	}
	data.Repos = append(data.Repos, r)
	args.saveDataFile(data)
	return nil
}

type devRemoveArgs struct {
	devArgs
	url string
}

func runDevRemove(cmd *cobra.Command, args *devRemoveArgs) error {
	data, err := args.loadDataFile()
	if err != nil {
		return err
	}

	out := []repo{}
	for _, s := range data.Repos {
		if strings.HasSuffix(s.Url, args.url) {
			out = append(out, s)
		}
	}
	data.Repos = out
	args.saveDataFile(data)
	return nil
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

func SetupDevCmd(dev *cobra.Command) error {
	defaultRoot, err := getDefaultRoot()
	if err != nil {
		return err
	}

	// Init
	initArgs := &devInitArgs{}
	devInit := &cobra.Command{
		Use:          "init [flags] dev_root",
		Short:        "Initialize dev repos",
		SilenceUsage: true,
		RunE:         cli.RunArgs(runDevInit, initArgs),
		Args:         cobra.ExactArgs(1),
	}
	devInit.Flags().IntVarP(&initArgs.parallel, "parallel", "p", 4, "Number of repos to download in parallel")
	devInit.Flags().BoolVar(&initArgs.noLink, "local", false, "Do not set as global extism-dev path")
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
	devExec.Flags().StringVar(&execArgs.root, "root", defaultRoot, "Root of extism development repos, all packages will be cloned into directories matching their github URLs inside this directory")
	devExec.Flags().StringVarP(&execArgs.category, "category", "c", "", "Category: sdk, pdk, plugin, runtime or other")
	devExec.Flags().StringVarP(&execArgs.repo, "repo", "r", "", "Regex filter used on the repo name")
	devExec.Flags().StringVarP(&execArgs.shell, "shell", "s", defaultShell, "Shell to use when executing commands")
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
	devFind.Flags().StringVar(&findArgs.root, "root", defaultRoot, "Root of extism development repos, all packages will be cloned into directories matching their github URLs inside this directory")
	devFind.Flags().StringVarP(&findArgs.category, "category", "c", "", "Category: sdk, pdk, plugin, runtime or other")
	devFind.Flags().StringVarP(&findArgs.repo, "repo", "r", "", "Regex filter used on the repo name")
	devFind.Flags().BoolVar(&findArgs.onlyFilename, "only-filename", false, "Only search filenames")
	devFind.Flags().StringVar(&findArgs.filename, "filename", "", "Filter for filenames")
	devFind.Flags().StringVar(&findArgs.editor, "editor", defaultEditor, "Editor command")
	devFind.Flags().BoolVar(&findArgs.edit, "edit", false, "Edit matching files")
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
	devAdd.Flags().StringVar(&addArgs.root, "root", defaultRoot, "Root of extism development repos, all packages will be cloned into directories matching their github URLs inside this directory")
	devAdd.Flags().StringVarP(&addArgs.url, "url", "u", "", "Repository URL, for example git@github.com:extism/extism")
	devAdd.MarkFlagRequired("url")
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
	devRemove.Flags().StringVar(&removeArgs.root, "root", defaultRoot, "Root of extism development repos, all packages will be cloned into directories matching their github URLs inside this directory")
	devRemove.Flags().StringVarP(&removeArgs.url, "url", "u", "", "Repository URL, for example git@github.com:extism/extism")
	devRemove.MarkFlagRequired("url")
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

	return nil
}
