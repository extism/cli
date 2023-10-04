package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/extism/go-sdk"
	"github.com/spf13/cobra"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/sys"
)

type callArgs struct {
	args           []string
	input          string
	loop           int
	wasi           bool
	logLevel       string
	allowedPaths   []string
	allowedHosts   []string
	timeout        uint64
	memoryMaxPages int
	config         []string
	setConfig      string
	manifest       bool
	stdin          bool
}

func readStdin() []byte {
	var buf []byte = make([]byte, 4096)
	var dest = []byte{}

	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			break
		}

		dest = append(dest, buf[0:n]...)
	}

	return dest
}

func (a *callArgs) SetArgs(args []string) {
	a.args = args
}

func (a *callArgs) getAllowedPaths() map[string]string {
	allowedPaths := map[string]string{}
	for _, path := range a.allowedPaths {
		split := strings.Split(path, ":")
		switch len(split) {
		case 1:
			allowedPaths[path] = path
		case 2:
			allowedPaths[split[0]] = split[1]
		default:
			allowedPaths[split[0]] = strings.Join(split[1:], ":")
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
		split := strings.SplitN(cfg, "=", 2)
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
	wasm := call.args[0]
	funcName := call.args[1]

	// Manifest
	var manifest extism.Manifest
	if call.manifest {
		Log("Reading from manifest:", wasm)
		f, err := os.Open(wasm)
		if err != nil {
			return err
		}
		err = json.NewDecoder(f).Decode(&manifest)
		if err != nil {
			return err
		}
		Log("Read manifest:", manifest)
		defer f.Close()
	} else {
		if strings.HasPrefix(wasm, "http://") || strings.HasPrefix(wasm, "https://") {
			Log("Loading wasm file as url:", wasm)
			manifest.Wasm = append(manifest.Wasm, extism.WasmUrl{Url: wasm})
		} else {
			Log("Adding wasm file to manifest:", wasm)
			manifest.Wasm = append(manifest.Wasm, extism.WasmFile{Path: wasm})
		}
	}

	// Allowed hosts
	Log("Adding allowed hosts:", call.allowedHosts)
	manifest.AllowedHosts = append(manifest.AllowedHosts, call.allowedHosts...)

	// Allowed paths
	if manifest.AllowedPaths == nil {
		manifest.AllowedPaths = map[string]string{}
	}

	for k, v := range call.getAllowedPaths() {
		Log("Adding path mapping:", k+":"+v)
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
		Log("Adding config key", k+"="+v)
		manifest.Config[k] = v
	}

	// Memory
	if call.memoryMaxPages != 0 {
		Log("Max pages", call.memoryMaxPages)
		manifest.Memory.MaxPages = uint32(call.memoryMaxPages)
	}

	var logLevel extism.LogLevel = extism.Off
	switch call.logLevel {
	case "info":
		logLevel = extism.Info
	case "debug":
		logLevel = extism.Debug
	case "warn":
		logLevel = extism.Warn
	case "error":
		logLevel = extism.Error
	case "trace":
		logLevel = extism.Trace
	}

	pluginConfig := extism.PluginConfig{
		ModuleConfig:  wazero.NewModuleConfig().WithSysWalltime(),
		RuntimeConfig: wazero.NewRuntimeConfig().WithCloseOnContextDone(call.timeout > 0),
		LogLevel:      &logLevel,
		EnableWasi:    call.wasi,
	}

	if call.timeout > 0 {
		Log("Setting timeout", call.timeout)
		manifest.Timeout = call.timeout
	}

	Log("Creating plugin")
	plugin, err := extism.NewPlugin(ctx, manifest, pluginConfig, []extism.HostFunction{})
	if err != nil {
		return err
	}
	defer plugin.Close()

	input := []byte(call.input)
	if call.stdin {
		Log("Reading input from stdin")
		input = readStdin()
	}
	Log("Got", len(input), "bytes of input data")

	// Call the plugin in a loop
	for i := 0; i < call.loop; i++ {
		Log("Calling", funcName)
		exit, res, err := plugin.Call(funcName, input)
		if err != nil {
			if exit == sys.ExitCodeDeadlineExceeded {
				return errors.New("timeout")
			}
			return err
		}
		Log("Call returned", len(res), "bytes")
		fmt.Print(string(res))

		if call.loop > 1 {
			fmt.Println()
		}
	}

	return nil
}

func CallCmd() *cobra.Command {
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
	flags.BoolVar(&call.stdin, "stdin", false, "Read input from stdin")
	flags.IntVar(&call.loop, "loop", 1, "Number of times to call the function")
	flags.BoolVar(&call.wasi, "wasi", false, "Enable WASI")
	flags.StringArrayVar(&call.allowedPaths, "allow-path", []string{}, "Allow a path to be accessed from inside the Wasm sandbox, a path can be either a plain path or a map from HOST_PATH:GUEST_PATH")
	flags.StringArrayVar(&call.allowedHosts, "allow-host", []string{}, "Allow access to an HTTP host, if no hosts are listed then all requests will fail. Globs may be used for wildcards")
	flags.Uint64Var(&call.timeout, "timeout", 0, "Timeout in milliseconds")
	flags.IntVar(&call.memoryMaxPages, "memory-max", 0, "Maximum number of pages to allocate")
	flags.StringArrayVar(&call.config, "config", []string{}, "Set config values, should be in KEY=VALUE format")
	flags.StringVar(&call.setConfig, "set-config", "", "Create config object using JSON, this will be merged with any `config` arguments")
	flags.BoolVarP(&call.manifest, "manifest", "m", false, "When set the input file will be parsed as a JSON encoded Extism manifest instead of a WASM file")
	flags.StringVar(&call.logLevel, "log-level", "", "Set log level: trace, debug, warn, info, error")
	return cmd
}
