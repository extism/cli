package main

import (
	"io/ioutil"
	"os"
	"os/exec"
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
	build    bool
}

type wasmSource struct {
	path string
	mode os.FileMode
	data []byte
}

func (w *wasmSource) Get() ([]byte, error) {
	if len(w.data) > 0 {
		return w.data, nil
	}
	s, err := os.Stat(w.path)
	if err != nil {
		return []byte{}, err
	}
	w.mode = s.Mode()

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
	kernel := wasmSource{path: kernelPath}

	if (args.all || args.kernel) && args.build {
		cmd := exec.Command("bash", "build.sh")
		cmd.Dir = args.Path("extism", "extism", "kernel")
		cli.Print("Building extism-runtime.wasm")
		output, err := cmd.CombinedOutput()
		if err != nil {
			cli.Log("Error building kernel:", err)
			cli.Print(string(output))
			return err
		}
	}

	if (args.all || args.wasm) && args.build {
		cmd := exec.Command("make")
		cmd.Dir = args.Path("extism", "plugins")
		cli.Print("Building plugins in extism/plugins")
		output, err := cmd.CombinedOutput()
		if err != nil {
			cli.Log("Error building plugins:", err)
			cli.Print(string(output))
			return err
		}

		cmd = exec.Command("make")
		cmd.Dir = args.Path("extism", "c-pdk")
		cli.Print("Building plugins in extism/c-pdk")
		output, err = cmd.CombinedOutput()
		if err != nil {
			cli.Log("Error building c-pdk plugins:", err)
			cli.Print(string(output))
			return err
		}
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

	// Get configured extra wasm files
	dataFile, err := args.loadDataFile()
	if err != nil {
		cli.Log("Unable to load data file", err)
		return err
	}
	sources := map[string]wasmSource{}

	for k, v := range dataFile.TestWasm {
		sources[k] = wasmSource{path: args.Path(v)}
	}

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

	search := NewSearch(nil, "", repos...)
	search.Iter(func(name string) error {
		fname := filepath.Base(name)

		// Update kernel
		if args.all || args.kernel {
			if fname == "extism-runtime.wasm" && name != kernelPath {
				cli.Print("Updating", name)
				if !args.dryRun {
					k, err := kernel.Get()
					if err != nil {
						cli.Log("Unable to load kernel", err)
						return err
					}
					err = ioutil.WriteFile(name, k, 0o655)
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
					err = ioutil.WriteFile(name, data, p.mode)
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
