package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/extism/go-sdk"
	"github.com/spf13/cobra"
	"github.com/tetratelabs/wazero"
)

type callArgs struct {
	args           []string
	input          string
	loop           int
	wasi           bool
	logLevel       *extism.LogLevel
	allowedPaths   []string
	allowedHosts   []string
	timeout        int
	memoryMaxPages int
	config         []string
	setConfig      string
	manifest       bool
}

func (a *callArgs) SetArgs(args []string) {
	a.args = args
}

func (a *callArgs) getAllowedPaths() map[string]string {
	allowedPaths := map[string]string{}
	for _, path := range a.allowedPaths {
		split := strings.SplitN(path, ":", 1)
		switch len(split) {
		case 1:
			allowedPaths[path] = path
		case 2:
			allowedPaths[split[0]] = split[1]
		default:
			continue
		}
	}

	return allowedPaths
}

func (a *callArgs) getConfig() (map[string]string, error) {
	config := map[string]string{}
	if a.setConfig != "" {
		err := json.Unmarshal([]byte(a.setConfig), &config)
		if err != nil {
			return config,
				errors.Join(errors.New("Invalid value for --set-config flag"), err)
		}
	}
	for _, cfg := range a.config {
		split := strings.SplitN(cfg, "=", 1)
		switch len(split) {
		case 1:
			config[cfg] = ""
		case 2:
			config[split[0]] = split[1]
		default:
			continue
		}
	}
	return config, nil
}

func runCall(cmd *cobra.Command, call *callArgs) error {
	if len(call.args) < 1 {
		return errors.New("An input file is required")
	} else if len(call.args) < 2 {
		return errors.New("A function name is required")
	}

	ctx := context.Background()
	cancel := func() {}
	wasm := call.args[0]
	funcName := call.args[1]

	// Manifest
	var manifest extism.Manifest
	if call.manifest {
		f, err := os.Open(wasm)
		if err != nil {
			return err
		}
		err = json.NewDecoder(f).Decode(&manifest)
		if err != nil {
			return err
		}
		defer f.Close()
	} else {
		manifest.Wasm = append(manifest.Wasm, extism.WasmFile{Path: wasm})
	}

	// Allowed hosts
	manifest.AllowedHosts = append(manifest.AllowedHosts, call.allowedHosts...)

	// Allowed paths
	if manifest.AllowedPaths == nil {
		manifest.AllowedPaths = map[string]string{}
	}

	for k, v := range call.getAllowedPaths() {
		manifest.AllowedPaths[k] = v
	}

	// Config
	if manifest.Config == nil {
		manifest.Config = map[string]string{}
	}
	config, err := call.getConfig()
	if err != nil {
		return err
	}
	for k, v := range config {
		manifest.Config[k] = v
	}

	// Memory
	if call.memoryMaxPages != 0 {
		manifest.Memory.MaxPages = uint32(call.memoryMaxPages)
	}

	pluginConfig := extism.PluginConfig{
		ModuleConfig:  wazero.NewModuleConfig().WithSysWalltime(),
		RuntimeConfig: wazero.NewRuntimeConfig().WithCloseOnContextDone(call.timeout > 0),
		LogLevel:      call.logLevel,
		EnableWasi:    call.wasi,
	}

	if call.timeout > 0 {
		manifest.Timeout = time.Millisecond * time.Duration(call.timeout)

		// TODO: why do I have to set this myself?
		ctx, cancel = context.WithTimeout(ctx, manifest.Timeout)
	}
	defer cancel()

	plugin, err := extism.NewPlugin(ctx, manifest, pluginConfig, []extism.HostFunction{})
	if err != nil {
		return err
	}
	defer plugin.Close()

	// Call the plugin in a loop
	for i := 0; i < call.loop; i++ {
		_, res, err := plugin.Call(funcName, []byte(call.input))
		if err != nil {
			return err
		}
		fmt.Print(string(res))

		if call.loop > 1 {
			fmt.Println()
		}
	}

	return nil
}

func callCmd() *cobra.Command {
	call := &callArgs{}
	cmd :=
		&cobra.Command{
			Use:          "call",
			Short:        "Call a plugin function",
			SilenceUsage: true,
			RunE:         runArgs(runCall, call),
		}
	flags := cmd.Flags()
	flags.StringVarP(&call.input, "input", "i", "", "Input data")
	flags.IntVar(&call.loop, "loop", 1, "Number of times to call the function")
	flags.BoolVar(&call.wasi, "wasi", false, "Enable WASI")
	flags.StringArrayVar(&call.allowedPaths, "allow-path", []string{}, "Allow a path to be accessed from inside the Wasm sandbox, a path can be either a plain path or a map from HOST_PATH:GUEST_PATH")
	flags.StringArrayVar(&call.allowedHosts, "allow-host", []string{}, "Allow access to an HTTP host, if no hosts are listed then all requests will fail. Globs may be used for wildcards")
	flags.IntVar(&call.timeout, "timeout", 0, "Timeout in milliseconds")
	flags.IntVar(&call.memoryMaxPages, "memory-max", 0, "Maximum number of pages to allocate")
	flags.StringArrayVar(&call.config, "config", []string{}, "Set config values, should be in KEY=VALUE format")
	flags.StringVar(&call.setConfig, "set-config", "", "Create config object using JSON, this will be merged with any `config` arguments")
	flags.BoolVarP(&call.manifest, "manifest", "m", false, "When set the input file will be parsed as a JSON encoded Extism manifest instead of a WASM file")
	return cmd
}