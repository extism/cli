#!/usr/bin/env bash
#
# usage: `./scripts/update-deps.sh`

set -eu

BRANCH="update-deps-$(date +%s)"
git checkout main
git checkout -b "$BRANCH"
go get -u
go mod tidy

# push to new branch, get HEAD hash
git commit -am "chore: update cli deps"
git push origin "$BRANCH"

echo "Update complete"
echo "If gh is installed, a pull-request can be opened with the following command"
echo "gh pr create --title \"chore: update dependencies\" --fill"
