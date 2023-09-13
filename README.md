# extism CLI

The [Extism](https://github.com/extism/extism) CLI can be used to execute Extism plugins and manage libextism installations.

## Installation

```shell
$ go install github.com/extism/cli
```

### Call a plugin

The following will call the `count_vowels` function in the `count-vowels.wasm` module with the input "qwertyuiop":

```shell
$ extism call count-vowels.wasm count_vowels --input qwertyuiop
```

See `extism call --help` for a list of all the flags

### Listing libextism versions

To list the available libextism versions:

```shell
$ extism lib versions
```

### Install libextism

To install the latest version of `libextism` to `/usr/local`, this will update any existing installation:

```shell
$ sudo extism lib install --version latest
```

To build the latest version from git:

```shell
$ extism lib install --version git --prefix ~/.local
```

This will clone the git repo into `~/.extism`, to configure a different directory:

```shell
$ extism lib install --git /path/to/repo --prefix ~/.local
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
