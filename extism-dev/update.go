package main

import (
	"io/ioutil"
	"path/filepath"

	"github.com/extism/cli"
	"github.com/spf13/cobra"
)

type devUpdateArgs struct {
	devArgs
	kernel bool
}

func runDevUpdate(cmd *cobra.Command, args *devUpdateArgs) error {
	data, err := args.loadDataFile()
	if err != nil {
		return err
	}

	kernelPath := args.Path("extism", "extism", "runtime", "src", "extism-runtime.wasm")
	kernelData, err := ioutil.ReadFile(kernelPath)
	if err != nil {
		return err
	}

	repos := []repo{}
	for _, repo := range data.Repos {
		repos = append(repos, repo)
	}
	search := NewSearch(nil, "", repos...)
	search.Iter(func(name string) error {
		fname := filepath.Base(name)

		if args.kernel {
			if fname == "extism-runtime.wasm" && name != kernelPath {
				cli.Print("Updating", name)
				err := ioutil.WriteFile(name, kernelData, 0o655)
				if err != nil {
					cli.Print("Error copying extism-kernel file", err)
				}
			}
		}
		return nil
	})

	return nil
}
