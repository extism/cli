package main

import (
	_ "embed"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/extism/cli"
	"github.com/spf13/cobra"
)

//go:embed repos.json
var repos []byte

type repoCategory int

const (
	Runtime = iota
	SDK
	PDK
	Other
)

func (s repoCategory) String() string {
	switch s {
	case Runtime:
		return "runtime"
	case SDK:
		return "sdk"
	case PDK:
		return "pdk"
	default:
		return "other"
	}
}

func (s *repoCategory) Parse(cat string) {
	if cat == "runtime" {
		*s = Runtime
	} else if cat == "sdk" {
		*s = SDK
	} else if cat == "pdk" {
		*s = PDK
	} else {
		*s = Other
	}
}
func (s repoCategory) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *repoCategory) UnmarshalJSON(data []byte) (err error) {
	var cat string
	if err := json.Unmarshal(data, &cat); err != nil {
		return err
	}
	s.Parse(cat)
	return nil
}

type repo struct {
	Url      string       `json:"url"`
	Category repoCategory `json:"category"`
}

func (r repo) Split() (string, string) {
	split := strings.Split(r.Url, "/")
	userName := split[len(split)-2]
	repoName := split[len(split)-1]
	if strings.HasPrefix(r.Url, "git@") {
		x := strings.Split(userName, ":")
		userName = x[1]
	}
	return userName, repoName
}

var defaultRepos []repo

func init() {
	if err := json.Unmarshal(repos, &defaultRepos); err != nil {
		panic(err)
	}
}

type devArgs struct {
	args []string
	root string
}

type extismData struct {
	Repos []repo `json:"repos"`
}

func (a devArgs) loadDataFile() (*extismData, error) {
	p := filepath.Join(a.root, ".extism.dev.json")
	cli.Log("Loading data file from", p)
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out extismData
	if err := json.NewDecoder(f).Decode(&out); err != nil {
		return nil, err
	}

	return &out, nil
}

func (a devArgs) saveDataFile(data *extismData) error {
	p := filepath.Join(a.root, ".extism.dev.json")
	cli.Log("Saving data file to", p)
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(*data)
}

func (a devArgs) link() error {
	dest := filepath.Join(homeDir(), ".extism.dev")
	abs, err := filepath.Abs(a.root)
	if err != nil {
		return err
	}

	os.Remove(dest)
	cli.Print("Creating symlink from", abs, "to", dest)
	return os.Symlink(abs, dest)
}

func (args *devArgs) mergeRepos(data *extismData) {
	m := map[string]repo{}
	for _, r := range defaultRepos {
		m[r.Url] = r
	}

	for _, r := range data.Repos {
		m[r.Url] = r
	}

	data.Repos = []repo{}
	for _, v := range m {
		data.Repos = append(data.Repos, v)
	}
	sort.Slice(data.Repos, func(a, b int) bool {
		if data.Repos[a].Category != data.Repos[b].Category {
			return data.Repos[a].Category < data.Repos[b].Category
		}

		return data.Repos[a].Url < data.Repos[b].Url
	})
}

func (a *devArgs) SetArgs(args []string) {
	a.args = args
}

type devInitArgs struct {
	devArgs
	parallel int
	noLink   bool
}

func (repo repo) clone(root string) {
	cli.Log("Initializing", repo.Url)
	userName, repoName := repo.Split()
	os.MkdirAll(userName, 0o755)

	full := filepath.Join(root, userName, repoName)
	cli.Print("Cloning", repo.Url, "to", full)
	_, err := os.Stat(full)
	if err == nil {
		cli.Print("Warning:", repo.Url, "already exists")
		return
	}

	cli.Log("Running git clone", repo.Url, full)
	if err := exec.Command("git", "clone", repo.Url, full).Run(); err != nil {
		cli.Print("Warning: git clone", repo.Url, "failed:", err)
	}
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

type devEachArgs struct {
	devArgs
	category string
	filter   string
	shell    string
}

func runDevEach(cmd *cobra.Command, each *devEachArgs) error {
	data, err := each.loadDataFile()
	if err != nil {
		return err
	}

	for _, repo := range data.Repos {
		if each.category != "" && repo.Category.String() != each.category {
			continue
		}

		if each.filter != "" {
			if !regexp.MustCompile(each.filter).MatchString(repo.Url) {
				continue
			}
		}
		userName, repoName := repo.Split()
		p := filepath.Join(each.root, userName, repoName)
		cli.Log("Executing", each.args[0], "in", p, "using", each.shell)
		cmd := exec.Command(each.shell, "-c", each.args[0])
		cmd.Dir = p
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "EXTISM_DEV_ROOT="+each.root)
		cmd.Env = append(cmd.Env, "EXTISM_DEV_RUNTIME="+filepath.Join(each.root, "extism", "extism"))
		cmd.Env = append(cmd.Env, "EXTISM_DEV_REPO="+repo.Url)
		cmd.Env = append(cmd.Env, "EXTISM_DEV_CATEGORY"+repo.Category.String())
		if err := cmd.Run(); err != nil {
			cli.Print("ERROR: command failed in", p)
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
		if s.Url != args.url {
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

	// Each
	eachArgs := &devEachArgs{}
	devEach := &cobra.Command{
		Use:          "each",
		Short:        "Run a command in each repo",
		SilenceUsage: true,
		RunE:         cli.RunArgs(runDevEach, eachArgs),
		Args:         cobra.ExactArgs(1),
	}
	devEach.Flags().StringVar(&eachArgs.root, "root", defaultRoot, "Root of extism development repos, all packages will be cloned into directories matching their github URLs inside this directory")
	devEach.Flags().StringVarP(&eachArgs.category, "category", "c", "", "Category: sdk, pdk, runtime or other")
	devEach.Flags().StringVarP(&eachArgs.filter, "filter", "f", "", "Regex filter used on the repo name")
	devEach.Flags().StringVarP(&eachArgs.shell, "shell", "s", "bash", "Shell to use when executing commands")
	dev.AddCommand(devEach)

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
	devAdd.Flags().StringVarP(&addArgs.category, "category", "c", "other", "Category: sdk, pdk, runtime or other")
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
