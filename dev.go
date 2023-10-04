package cli

import (
	"encoding/json"
	"os/exec"
	"regexp"
	"strings"

	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

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
	case Other:
		return "other"
	}
	return "other"
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

var defaultRepos []repo = []repo{
	{
		Url:      "git@github.com:extism/extism",
		Category: Runtime,
	},
	{
		Url:      "git@github.com:extism/go-sdk",
		Category: Runtime,
	},
	{
		Url:      "git@github.com:extism/js-sdk",
		Category: Runtime,
	},
	{
		Url:      "git@github.com:extism/python-sdk",
		Category: SDK,
	},
	{
		Url:      "git@github.com:extism/ruby-sdk",
		Category: SDK,
	},
	{
		Url:      "git@github.com:extism/elixir-sdk",
		Category: SDK,
	},
	{
		Url:      "git@github.com:extism/php-sdk",
		Category: SDK,
	},
	{
		Url:      "git@github.com:extism/haskell-sdk",
		Category: SDK,
	},
	{
		Url:      "git@github.com:extism/ocaml-sdk",
		Category: SDK,
	},
	{
		Url:      "git@github.com:extism/cpp-sdk",
		Category: SDK,
	},
	{
		Url:      "git@github.com:extism/java-sdk",
		Category: SDK,
	},
	{
		Url:      "git@github.com:extism/dotnet-sdk",
		Category: SDK,
	},
	{
		Url:      "git@github.com:extism/zig-sdk",
		Category: SDK,
	},
	{
		Url:      "git@github.com:extism/zig-pdk",
		Category: PDK,
	},
	{
		Url:      "git@github.com:extism/go-pdk",
		Category: PDK,
	},
	{
		Url:      "git@github.com:extism/rust-pdk",
		Category: PDK,
	},
	{
		Url:      "git@github.com:extism/assemblyscript-pdk",
		Category: PDK,
	},
	{
		Url:      "git@github.com:extism/c-pdk",
		Category: PDK,
	},
	{
		Url:      "git@github.com:extism/haskell-pdk",
		Category: PDK,
	},
	{
		Url:      "git@github.com:extism/extism-dbg",
		Category: Other,
	},
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
	Log("Loading data file from", p)
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
	Log("Saving data file to", p)
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(*data)
}

func (a *devArgs) SetArgs(args []string) {
	a.args = args
}

type devInitArgs struct {
	devArgs
}

func runDevInit(cmd *cobra.Command, args *devInitArgs) error {
	data := &extismData{
		Repos: defaultRepos,
	}

	Log("Initializing Extism dev repos in", args.root)
	Log("Repos", data.Repos)
	os.MkdirAll(args.root, 0o755)

	for _, repo := range data.Repos {
		Log("Initializing", repo.Url)
		userName, repoName := repo.Split()
		os.MkdirAll(userName, 0o755)

		full := filepath.Join(args.root, userName, repoName)
		Print("Cloning", repo.Url, "to", full)
		_, err := os.Stat(full)
		if err == nil {
			Print("Warning:", repo.Url, "already exists")
			continue
		}

		Log("Running git clone", repo.Url, full)
		if err := exec.Command("git", "clone", repo.Url, full).Run(); err != nil {
			Print("Warning: git clone", repo.Url, "failed:", err)
		}
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
		Log("Executing", each.args, "in", p)
		cmd := exec.Command(each.args[0], each.args[1:]...)
		cmd.Dir = p
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "EXTISM_DEV_PATH="+each.root)
		cmd.Env = append(cmd.Env, "EXTISM_DEV_RUNTIME_PATH="+filepath.Join(each.root, "extism", "extism"))
		if err := cmd.Run(); err != nil {
			Print("Error in", p+":", err)
			Print()
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
			Print("Repo already exists, not adding")
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

func DevCmd() *cobra.Command {
	dev := &cobra.Command{
		Use:   "dev",
		Short: "Manage extism dev environment",
	}

	// Init
	initArgs := &devInitArgs{}
	devInit := &cobra.Command{
		Use:          "init",
		Short:        "Initialize dev repos",
		SilenceUsage: true,
		RunE:         runArgs(runDevInit, initArgs),
	}
	devInit.Flags().StringVar(&initArgs.root, "root", filepath.Join(os.Getenv("HOME")), "Root of extism development repos, all packages will be cloned into directories matching their github URLs inside this directory")
	dev.AddCommand(devInit)

	// Each
	eachArgs := &devEachArgs{}
	devEach := &cobra.Command{
		Use:          "each",
		Short:        "Run a command in each repo",
		SilenceUsage: true,
		RunE:         runArgs(runDevEach, eachArgs),
	}
	devEach.Flags().StringVar(&eachArgs.root, "root", filepath.Join(os.Getenv("HOME")), "Root of extism development repos, all packages will be cloned into directories matching their github URLs inside this directory")
	devEach.Flags().StringVarP(&eachArgs.category, "category", "c", "", "Category: sdk, pdk, runtime or other")
	devEach.Flags().StringVarP(&eachArgs.filter, "filter", "f", "", "Regex filter used on the repo name")
	dev.AddCommand(devEach)

	// Add
	addArgs := &devAddArgs{}
	devAdd := &cobra.Command{
		Use:          "add",
		Short:        "Add a repo",
		SilenceUsage: true,
		RunE:         runArgs(runDevAdd, addArgs),
	}
	devAdd.Flags().StringVar(&addArgs.root, "root", filepath.Join(os.Getenv("HOME")), "Root of extism development repos, all packages will be cloned into directories matching their github URLs inside this directory")
	devAdd.Flags().StringVarP(&addArgs.url, "url", "u", "", "Repository URL, for example git@github.com:extism/extism")
	devAdd.MarkFlagRequired("url")
	devAdd.Flags().StringVarP(&addArgs.category, "category", "c", "other", "Category: sdk, pdk, runtime or other")
	dev.AddCommand(devAdd)

	// Remove
	removeArgs := &devRemoveArgs{}
	devRemove := &cobra.Command{
		Use:          "remove",
		Aliases:      []string{"rm"},
		Short:        "Remove a repo",
		SilenceUsage: true,
		RunE:         runArgs(runDevRemove, removeArgs),
	}
	devRemove.Flags().StringVar(&removeArgs.root, "root", filepath.Join(os.Getenv("HOME")), "Root of extism development repos, all packages will be cloned into directories matching their github URLs inside this directory")
	devRemove.Flags().StringVarP(&removeArgs.url, "url", "u", "", "Repository URL, for example git@github.com:extism/extism")
	devRemove.MarkFlagRequired("url")
	dev.AddCommand(devRemove)

	return dev
}
