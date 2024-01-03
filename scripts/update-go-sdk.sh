#!/usr/bin/env bash
#
# usage: `./scripts/update-go-sdk.sh`

VERSION=${1-main}

# update extism
cd extism || exit 1
go get -u "github.com/extism/go-sdk@$VERSION"
go mod tidy

# update extism-dev
cd ../extism-dev || exit 1
go get -u "github.com/extism/go-sdk@$VERSION"
go mod tidy

# Create commit and push
git commit -am "chore: update go-sdk"
