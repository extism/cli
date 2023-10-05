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

func (repo repo) clone(root string) {
	cli.Log("Initializing", repo.Url)
	userName, repoName := repo.split()
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

var defaultRepos []repo

func init() {
	if err := json.Unmarshal(repos, &defaultRepos); err != nil {
		panic(err)
	}
}
