#!/usr/bin/env bash
#
# usage: `./scripts/cleanup-update-branches.sh`

set -e

git branch -l "update-deps-*" --format "%(refname:short)"| while read -r branch ; do 
  echo "Removing local branch: $branch"
  git branch -D $branch
done

git branch -r --format "%(refname:short)" | grep -o 'update-deps-.*' | while read -r branch ; do
  echo "Removing remote branch: $branch"
  git push origin --delete $branch
done



