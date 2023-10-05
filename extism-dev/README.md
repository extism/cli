# extism-dev

A command-line tool for managing Extism repos

## Dependencies

- go
- git

## Installation

```shell
go install github.com/extism/cli/extism-dev
```

## Usage

### Init

The first step is to initialize your `extism-dev` root path:

```shell
extism-dev init --root ~/devel
```

This will download all the repos into `~/dev` using the github orginization as the namespace.
For example, `git@github.com:extism/extism` will be downloaded into `~/dev/extism/extism`

It will also create a `.extism.dev.json` file in the root directory that is used to configure which repos to include. This file can be
updated using `extism-dev add` and `extism-dev remove`.

### Exec

Once the environment is setup, you can use `extism-dev exec` to run commands in every repo. For example, 
to list every open PR using 'gh':

```shell
extism-dev exec 'gh pr list'
```

### Find

`extism-dev find` can be used to search all the repos at once and do regex-based text substitution.

To replace the version of the "base64" crate in every `Cargo.toml` file:

```shell
extism-dev find --filename 'Cargo.toml' 'base64 = ".*"' --replace 'base64 = "1.0.0"'
```
