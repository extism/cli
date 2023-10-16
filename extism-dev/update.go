package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/extism/cli"
	"github.com/spf13/cobra"
)

type devUpdateArgs struct {
	devArgs
	kernel   bool
	wasm     bool
	all      bool
	dryRun   bool
	repo     string
	category string
}

type wasmSource struct {
	path string
	data []byte
}

func (w *wasmSource) Get() ([]byte, error) {
	if len(w.data) > 0 {
		return w.data, nil
	}

	d, err := ioutil.ReadFile(w.path)
	if err != nil {
		return []byte{}, err
	}

	w.data = d
	return w.data, nil
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
		if args.category != "" && repo.Category != args.category {
			continue
		}

		if args.repo != "" {
			rx := regexp.MustCompile(args.repo)
			if !rx.Match([]byte(repo.path())) {
				continue
			}
		}
		repos = append(repos, repo)
	}

	sources := map[string]wasmSource{}

	pluginsDir := args.Path("extism", "plugins", "target", "wasm32-unknown-unknown", "release")

	// Get plugins from `plugins` directory
	files, err := os.ReadDir(pluginsDir)
	if err != nil {
		files = []os.DirEntry{}
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		sources[name] = wasmSource{path: filepath.Join(pluginsDir, name)}
		if name == "count_vowels.wasm" {
			sources["code.wasm"] = sources[name]
		} else if name == "loop_forever.wasm" {
			sources["loop.wasm"] = sources[name]
		}
	}

	// A few from the C-PDK still
	sources["code-functions.wasm"] = wasmSource{path: args.Path("extism", "c-pdk", "examples", "host-functions", "host-functions.wasm")}
	sources["globals.wasm"] = wasmSource{path: args.Path("extism", "c-pdk", "examples", "globals", "globals.wasm")}

	search := NewSearch(nil, "", repos...)
	search.Iter(func(name string) error {
		fname := filepath.Base(name)

		// Update kernel
		if args.all || args.kernel {
			if fname == "extism-runtime.wasm" && name != kernelPath {
				cli.Print("Updating", name)
				if !args.dryRun {
					err := ioutil.WriteFile(name, kernelData, 0o655)
					if err != nil {
						cli.Print("Error copying extism-kernel file to", name+":", err)
					}
				}
			}
		}

		// Update test wasm modules
		if args.all || args.wasm {
			if p, ok := sources[fname]; ok && name != p.path {
				cli.Print("Updating", name)
				data, err := p.Get()
				if err != nil {
					cli.Print("Error reading file", err)
					return err
				}
				if !args.dryRun {
					err = ioutil.WriteFile(name, data, 0o655)
					if err != nil {
						cli.Print("Error copying", p, "to", name+":", err)
					}
				}
			}
		}

		return nil
	})

	return nil
}
