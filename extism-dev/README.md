# extism-dev

A command-line tool for managing Extism repos

## Dependencies

- go
- git

## Installation

```shell
go install github.com/extism/cli/extism-dev@latest
```

## Usage

### Init

The first step is to initialize your `extism-dev` root path:

```shell
extism-dev init --root ~/devel
```

Once a directory has been initialized you can run the same command without any `--root` argument to re-initialize the existing environment:

```shell
extism-dev init
```

### Clean

To remove the files created by `extism-dev`:

```shell
extism-dev clean
```

This will download all the repos into `~/dev` using the github orginization as the namespace.
For example, `git@github.com:extism/extism` will be downloaded into `~/dev/extism/extism`

It will also create a `.extism.dev.json` file in the root directory that is used to configure which repos to include. This file can be
updated using `extism-dev add` and `extism-dev remove`.

### Exec

Once the environment is setup, you can use `extism-dev exec` to run commands in every repo.

For example, to list every open PR using 'gh':

```shell
extism-dev exec -- gh pr list
```
The `--repo` flag can be used to select a specific repo, or set of repos:

```shell
extism-dev exec --repo 'go-sdk|js-sdk' -- gh pr list
```

The following environment variables are available when using `exec`:
- `EXTISM_DEV_ROOT` - the root path of the extism-dev environment
- `EXTISM_DEV_RUNTIME` - the path of the `extism/extism` project
- `EXTISM_DEV_REPO_URL` - the url of the target repo
- `EXTISM_DEV_REPO_CATEGORY` - the category of the target repo
- `$EXTISM_DEV_ROOT/.bin` is also added to the `$PATH` while executing commands

### Find

`extism-dev find` can be used to search all the repos at once and do regex-based text substitution.

Search for files that contain `base64`:

```shell
extism-dev find base64
```

Search for files that contain `base64` in `extism/extism`:

```shell
extism-dev find --repo 'extism/extism' base64
```

To replace the version of the "base64" crate in every `Cargo.toml` file:

```shell
extism-dev find --filename 'Cargo.toml' 'base64 = ".*"' --replace 'base64 = "1.0.0"'
```

