#!/usr/bin/env bash
#
# usage: `./scripts/release.sh`

set -eu

VERSION=$(cat extism/VERSION)
TAG="v$VERSION"

git tag "$TAG"
git push origin "$TAG"
gh release create "$TAG" --generate-notes
