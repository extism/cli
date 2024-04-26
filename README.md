# extism CLI

The [Extism](https://github.com/extism/extism) CLI can be used to generate or
execute Extism plugins and manage libextism installations.

## Installation

### Using curl/sh:

```shell
curl https://get.extism.org/cli | sh
```

See the help output for more options:

```shell
curl https://get.extism.org/cli | sh -s -- -h
```

### From source:

```shell
go install github.com/extism/cli/extism@latest
```

### Using Nix:

```shell
nix-shell -p extism-cli
```

### Manual

You can also download and extract the latest release from
[https://github.com/extism/cli/releases](https://github.com/extism/cli/releases)

## Generate a Plugin

To quickly start writing an Extism plugin in any of the supported PDK languages,
using `extism` CLI to generate minimal boilerplate may be helpful:

```sh
mkdir js-plugin && cd js-plugin
extism generate plugin 

  Select a PDK language to use for your plugin:  
                                                 
    1. Rust                                      
  > 2. JavaScript                                
    3. Go                                        
    4. Zig                                       
    5. C#                                        
    6. F#                                        
    7. C                                         
    8. Haskell                                   
    9. AssemblyScript                            
# or pass a path to the output via `-o`, see more options running `extism generate -h`
```

This will output a quickstart plugin project with the necessary configuration
and dependencies ready. If no output directory was specified, the current
directory will be used.

**NOTE:**: You may still need to install language tools such as compilers or
other system dependencies to compile the plugin to WebAssembly.

To further improve this, we will eventually include a `Dockerfile` in each
generated project, so a build environment can be easily created with all tools
necessary. If you're interested in contributing to this effort, please join us
on [Discord](https://extism.org), and check out the PDK template repository of
interest, listed in
[pdk-templates.json](https://github.com/extism/cli/blob/main/pdk-templates.json)

## Call a plugin

The following will call the `count_vowels` function in the `count_vowels.wasm`
module with the input "qwertyuiop":

```shell
PLUGIN_URL="https://github.com/extism/plugins/releases/latest/download/count_vowels.wasm"
extism call $PLUGIN_URL count_vowels --input qwertyuiop
```

> **Note**: The first parameter to `call` can also be a path to a Wasm file on
> disk.

See `extism call --help` for a list of all the flags

## Listing libextism versions

To list the available libextism versions:

```shell
extism lib versions
```

To list the available triples for a version:

```shell
extism lib versions v0.0.1-alpha
```

### Install libextism

To install the latest version of `libextism` to `/usr/local` on macOS and Linux
and `.` on Windows, this will overwrite any existing installation at the same
path:

```shell
sudo PATH=$PATH env extism lib install
```

To install to `$HOME/.local`:

```shell
extism lib install --prefix ~/.local
```

To install the latest build from github:

```shell
sudo PATH=$PATH env extism lib install --version git
```

### Uninstall libextism

To uninstall the shared object and header installed in `/usr/local`:

```shell
sudo PATH=$PATH env extism lib uninstall
```

Or from another prefix:

```shell
extism lib uninstall --prefix ~/.local
```

### Check a libextism installation

The `lib check` command will print the version of the installed `libextism`
library:

```shell
extism lib check
```
