package main

import (
	"errors"
	"fmt"

	"github.com/extism/cli"
	"github.com/spf13/cobra"
)

type devAddArgs struct {
	devArgs
	url      string
	category string
}

func runDevAdd(cmd *cobra.Command, args *devAddArgs) error {
	args.url = args.args[0]
	data, err := args.loadDataFile()
	if err != nil {
		return err
	}

	r := repo{
		Url: args.url,
	}
	r.Category.Parse(args.category)
	if !r.clone() {
		return errors.New(fmt.Sprint("Unable to clone repo:", r.Url))
	}
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
