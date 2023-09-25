# extism CLI

The [Extism](https://github.com/extism/extism) CLI can be used to execute Extism plugins and manage libextism installations.

## Installation

```shell
$ go install github.com/extism/cli/extism@latest
```

### Call a plugin

The following will call the `count_vowels` function in the `count_vowels.wasm` module with the input "qwertyuiop":

```shell
$ PLUGIN_URL="https://github.com/extism/plugins/releases/latest/download/count_vowels.wasm"
$ extism call $PLUGIN_URL count_vowels --input qwertyuiop
```

> **Note**: The first parameter to `call` can also be a path to a Wasm file on disk.

See `extism call --help` for a list of all the flags

### Listing libextism versions

To list the available libextism versions:

```shell
$ extism lib versions
```

### Install libextism

To install the latest version of `libextism` to `/usr/local`, this will overwrite any existing installation at the same path:

```shell
$ sudo extism lib install
```

To build the latest build from github:

```shell
$ extism lib install --version git
```

### Uninstall libextism

To uninstall the shared object and header installed in `/usr/local`:

```shell
$ sudo extism lib uninstall
```

Or from another prefix:

```shell
$ extism lib uninstall --prefix ~/.local
```

### Check a libextism installation

The `lib check` command will print the version of the installed `libextism` library:

```shell
$ extism lib check
```

