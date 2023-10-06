package main

import (
	"github.com/extism/cli"
	"github.com/spf13/cobra"
)

type devListArgs struct {
	devArgs
	category string
}

func runDevList(cmd *cobra.Command, args *devListArgs) error {
	data, err := args.loadDataFile()
	if err != nil {
		return err
	}

	for _, repo := range data.Repos {
		if args.category != "" && repo.Category.String() != args.category {
			continue
		}
		cli.Print(repo.path())
	}
	return nil
}
