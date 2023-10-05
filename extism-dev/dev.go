package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

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
	category    string
	filename    string
	edit        bool
	editor      string
	repo        string
	replace     string
	interactive bool
}

func (a *devFindArgs) prompt(msg ...any) bool {
	if !a.interactive {
		return true
	}

	fmt.Print(msg...)
	fmt.Print("? [y/n] ")

	c := 'n'
	_, err := fmt.Scanf("%c\n", &c)
	if err != nil {
		cli.Log("prompt failed:", err)
		return false
	}
	return c == 'y'
}

func runDevFind(cmd *cobra.Command, args *devFindArgs) error {
	data, err := args.loadDataFile()
	if err != nil {
		return err
	}
	query := ""
	if len(args.args) > 0 {
		query = args.args[0]
	}
	dirs := []string{}

	repoRegex := &regexp.Regexp{}
	if args.repo != "" {
		repoRegex = regexp.MustCompile(args.repo)
	}

	for _, repo := range data.Repos {
		if args.repo != "" {
			if !repoRegex.MatchString(repo.Url) {
				continue
			}
		}
		if args.category == "" || repo.Category.String() == args.category {
			userName, repoName := repo.split()
			p := filepath.Join(args.root, userName, repoName)
			dirs = append(dirs, p)
		}
	}

	search := NewSearch(args, query, dirs...)
	if args.filename != "" {
		search.FilterFilenames(args.filename)
	}

	if args.replace != "" {
		return search.Replace(args.replace)
	} else {
		if args.edit {
			lock := sync.Mutex{}
			return search.Iter(func(path string) error {
				lock.Lock()
				defer lock.Unlock()
				if !args.prompt("Edit ", path) {
					return nil
				}
				cli.Print("Editing", path)
				cmd := exec.Command(args.editor, path)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Stdin = os.Stdin
				return cmd.Run()
			})
		} else {
			return search.Iter(func(path string) error {
				cli.Print(path)
				return nil
			})
		}
	}

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
