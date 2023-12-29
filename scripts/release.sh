#!/usr/bin/env bash
#
# usage: `./scripts/release.sh`

set -e

gh release create v$(cat extism/VERSION) --generate-notes
