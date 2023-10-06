package main

import (
	"context"
	"regexp"
	"strings"

	"github.com/extism/cli"
	"github.com/extism/go-sdk"
	"github.com/spf13/cobra"
)

type devCallArgs struct {
	devArgs
	category string
	repo     string
	parallel int
	timeout  int
}

func runDevCall(cmd *cobra.Command, args *devCallArgs) error {
	plugin := args.args[0]
	function := args.args[1]

	data, err := args.loadDataFile()
	if err != nil {
		return err
	}

	rx := &regexp.Regexp{}
	if args.repo != "" {
		rx = regexp.MustCompile(args.repo)
	}

	wasm := []extism.Wasm{}
	if strings.HasPrefix(plugin, "http://") || strings.HasPrefix(plugin, "https://") {
		wasm = append(wasm, extism.WasmUrl{Url: plugin})
	} else {
		wasm = append(wasm, extism.WasmFile{Path: plugin})
	}

	pool := NewPool(args.parallel)
	for i, r := range data.Repos {
		RunTask(pool, func(repo repo) {
			if args.category != "" && repo.Category != args.category {
				return
			}

			if args.repo != "" {
				if !rx.MatchString(repo.Url) {
					return
				}
			}
			p := repo.path()
			cli.Log("Running plugin", plugin, "in", p)
			if args.parallel <= 1 {
				if i > 0 {
					cli.Print()
				}
				cli.Print(p)
			}
			ctx := context.Background()
			config := extism.PluginConfig{}
			manifest := extism.Manifest{
				Wasm:         wasm,
				AllowedPaths: map[string]string{p: "/"},
			}
			if args.timeout != 0 {
				manifest.Timeout = uint64(args.timeout)
			}
			plug, err := extism.NewPlugin(ctx, manifest, config, []extism.HostFunction{})
			if err != nil {
				cli.Print("Unable to create plugin:", err)
				return
			}
			_, output, err := plug.Call(function, []byte(repo.Url))
			if err != nil {
				cli.Print("Plugin call failed in", p+":", err)
			}
			cli.Print(string(output))
		}, r)
	}
	pool.Wait()
	return nil
}
