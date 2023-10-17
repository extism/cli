package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"github.com/extism/cli"
)

type devArgs struct {
	args []string
}

func (a devArgs) Path(p ...string) string {
	return filepath.Join(Root, filepath.Join(p...))
}

func (a devArgs) loadDataFile() (*extismData, error) {
	p := filepath.Join(Root, ".extism.dev.json")
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
	p := filepath.Join(Root, ".extism.dev.json")
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
	abs, err := filepath.Abs(Root)
	if err != nil {
		return err
	}

	os.Remove(dest)
	cli.Print("Creating symlink from", abs, "to", dest)
	return os.Symlink(abs, dest)
}

func (data *extismData) filterRepos(category string) {
	repos := []repo{}

	for _, r := range data.Repos {
		if r.Category == category {
			repos = append(repos, r)
		}
	}

	data.Repos = repos
}

func (data *extismData) mergeRepos() {
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

func (data *extismData) mergeExtraWasm() {
	m := map[string]string{}
	for k, v := range defaultExtraWasm {
		m[k] = v
	}

	for k, v := range data.ExtraWasm {
		m[k] = v
	}

	data.ExtraWasm = m
}

func (a *devArgs) SetArgs(args []string) {
	a.args = args
}
