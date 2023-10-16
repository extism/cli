package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sync"

	"github.com/extism/cli"
	"github.com/spf13/cobra"
)

type devFindArgs struct {
	devArgs
	category    string
	filename    string
	edit        bool
	editor      string
	repo        string
	replace     string
	dryRun      bool
	interactive bool
}

var promptLock = sync.Mutex{}
var editLock = sync.Mutex{}

func (a *devFindArgs) prompt(msg ...any) bool {
	if !a.interactive {
		return true
	}

	promptLock.Lock()
	defer promptLock.Unlock()

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

	repoRegex := &regexp.Regexp{}
	if args.repo != "" {
		repoRegex = regexp.MustCompile(args.repo)
	}

	repos := []repo{}
	for _, repo := range data.Repos {
		if args.repo != "" {
			if !repoRegex.MatchString(repo.Url) {
				continue
			}
		}
		if args.category == "" || repo.Category == args.category {
			repos = append(repos, repo)
		}
	}

	search := NewSearch(args, query, repos...)
	if args.filename != "" {
		search.FilterFilenames(args.filename)
	}

	if args.replace != "" {
		return search.Replace(args.replace)
	} else {
		if args.edit {
			return search.Iter(func(path string) error {
				editLock.Lock()
				defer editLock.Unlock()
				if args.dryRun {
					cli.Print("Edit", path)
					return nil
				}
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
