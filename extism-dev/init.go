package main

import (
	"os"

	"github.com/extism/cli"
	"github.com/spf13/cobra"
)

type devInitArgs struct {
	devArgs
	parallel int
	local    bool
}

func runDevInit(cmd *cobra.Command, args *devInitArgs) error {
	data, err := args.loadDataFile()
	if err != nil {
		data = &extismData{
			Repos: defaultRepos,
		}
	} else {
		args.mergeRepos(data)
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
