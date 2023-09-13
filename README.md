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

### Install libextism

To install the latest version of `libextism` to `/usr/local`, this will update any existing installation:

```shell
$ sudo extism lib install --version latest
```

### Uninstall libextism

To uninstall the shared object and header installed in `/usr/local`:

```shell
$ sudo extism lib uninstall
```