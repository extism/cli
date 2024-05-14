#!/usr/bin/env bash
#
# usage: `./scripts/update-go-sdk.sh`

VERSION=${1-main}

# update cli
go get -u "github.com/extism/go-sdk@$VERSION"
go mod tidy

# Create commit and push
git commit -am "chore: update go-sdk"
