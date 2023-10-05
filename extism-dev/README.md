# extism-dev

A command-line tool for managing Extism repos

## Installation

```shell
go install github.com/extism/cli/extism-dev
```

## Usage

### Init

The first step is to initialize your `extism-dev` root path:

```shell
extism-dev init ~/devel
```

This will download all the repos into `~/dev` using the github orginization as the namespace.
For example, `git@github.com:extism/extism` will be downloaded into `~/dev/extism/extism`

### 
