#!/usr/bin/env bash
#
# usage: `./scripts/release.sh`

set -eu
version=$(echo -n "$1" | sed 's/v//g')
TAG="v$version"

git tag "$TAG"
git push origin "$TAG"
gh release create "$TAG" --generate-notes
