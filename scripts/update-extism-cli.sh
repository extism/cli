#!/usr/bin/env bash
#
# usage: `./scripts/update-extism-cli.sh`

GIT_HASH=$(git rev-parse HEAD)

# update extism
cd extism || exit 1
go get -u "github.com/extism/cli@$GIT_HASH"

# update extism-dev
cd ../extism-dev || exit 1
go get -u "github.com/extism/cli@$GIT_HASH"
go mod tidy

# Create commit and push
git commit -am "chore: update extism and extism-dev deps"
