#!/usr/bin/env sh

set -e

dest=${1-$HOME/.local/bin}
curl --fail https://raw.githubusercontent.com/extism/cli/main/extism_cli/__init__.py > /tmp/extism-cli
mv /tmp/extism-cli "$dest"/extism
chmod +x "$dest"/extism
