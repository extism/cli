#!/usr/bin/env bash
#
# usage: ./scripts/set-version 0.0.0

set -eu
version=$(echo -n $1 | sed 's/v//g')
echo -n $version > extism/VERSION
