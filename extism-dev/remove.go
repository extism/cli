package main

import (
	"os"
	"strings"

	"github.com/extism/cli"
	"github.com/spf13/cobra"
)

type devRemoveArgs struct {
	devArgs
	url  string
	keep bool
}

func runDevRemove(cmd *cobra.Command, args *devRemoveArgs) error {
	args.url = args.args[0]
	data, err := args.loadDataFile()
	if err != nil {
		return err
	}

	out := []repo{}
	for _, s := range data.Repos {
		if !strings.HasSuffix(s.Url, args.url) {
			out = append(out, s)
		} else {
			p := s.path()
			err = os.RemoveAll(p)
			if err != nil {
				cli.Print("Error: unable to remove", p)
				out = append(out, s)
			}
		}
	}
	data.Repos = out
	args.saveDataFile(data)
	return nil
}
