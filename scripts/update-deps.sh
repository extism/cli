#!/usr/bin/env bash

TIMESTAMP=$(date +%s)
BRANCH="update-deps-$TIMESTAMP"
git checkout -b $BRANCH
go get -u
go mod tidy
git commit -am "chore: update cli deps"
GIT_HASH=$(git rev-parse HEAD)
git push origin $BRANCH

cd extism
go get -u
go get -u github.com/extism/cli@$GIT_HASH

cd ../extism-dev
go get -u
go get -u github.com/extism/cli@$GIT_HASH
git commit -am "chore: update extism and extism-dev deps"
push push origin $BRANCH

echo "Update complete"
echo "If gh is installed, a pull-request can be opened with the following command"
echo "gh pr create --title \"chore: update dependencies\" --fill"
