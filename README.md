# extism-cli

The `extism` CLI is used to manage [Extism](https://github.com/extism/extism) installations

## Install

Using curl:

```sh
$ curl https://raw.githubusercontent.com/extism/cli/main/install.sh | sh
```

Using pip:

```sh
$ pip3 install git+https://github.com/extism/cli
```

## Usage

The simplest use-case is to download an build the source code then install the library and header file into 
the installation prefix of your choice.

```sh
$ extism install # install to ~/.local/lib and ~/.local/include
```

By default the prefix is `~/.local`, but it can easily be configured:

```sh
$ extism --sudo --prefix /usr/local install
```