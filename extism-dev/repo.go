package main

import (
	_ "embed"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/extism/cli"
)

//go:embed repos.json
var repos []byte

type repo struct {
	Url      string `json:"url"`
	Category string `json:"category"`
}

func (r repo) path() string {
	a, b := r.split()
	return filepath.Join(Root, a, b)
}

func (r repo) split() (string, string) {
	split := strings.Split(r.Url, "/")
	userName := split[len(split)-2]
	repoName := split[len(split)-1]
	if strings.HasPrefix(r.Url, "git@") {
		x := strings.Split(userName, ":")
		userName = x[1]
	}
	return userName, repoName
}

func (repo repo) clone() bool {
	cli.Log("Initializing", repo.Url)
	userName, repoName := repo.split()
	os.MkdirAll(userName, 0o755)

	full := filepath.Join(Root, userName, repoName)
	cli.Print("Cloning", repo.Url, "to", full)
	_, err := os.Stat(full)
	if err == nil {
		cli.Print("Warning:", repo.Url, "already exists")
		return true
	}

	cli.Log("Running git clone", repo.Url, full)
	if err := exec.Command("git", "clone", repo.Url, full).Run(); err != nil {
		cli.Print("Warning: git clone", repo.Url, "failed:", err)
	}

	return false
}

var defaultRepos []repo

func init() {
	if err := json.Unmarshal(repos, &defaultRepos); err != nil {
		panic(err)
	}
}
