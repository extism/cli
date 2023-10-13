package main

import (
	"errors"
	"os"

	"github.com/extism/cli"
	"github.com/spf13/cobra"
)

type devInitArgs struct {
	devArgs
	parallel int
	local    bool
	category string
}

func runDevInit(cmd *cobra.Command, args *devInitArgs) error {
	if Root == "" {
		return errors.New("No root set, use `--root` to configure a root path")
	}
	data, err := args.loadDataFile()
	if err != nil {
		data = &extismData{
			Repos: defaultRepos,
		}
	} else {
		data.mergeRepos()
	}

	if args.category != "" {
		data.filterRepos(args.category)
	}

	cli.Print("Initializing Extism dev repos in", Root)
	err = os.MkdirAll(Root, 0o755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	pool := NewPool(args.parallel)
	for _, r := range data.Repos {
		cli.Log("Repos", data.Repos)
		RunTask(pool, func(repo repo) {
			repo.clone()
		}, r)
	}
	pool.Wait()

	if !args.local {
		if err := args.link(); err != nil {
			return err
		}
	}

	if err := args.saveDataFile(data); err != nil {
		return err
	}

	return nil
}
