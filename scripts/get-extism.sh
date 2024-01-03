#!/bin/sh

set -u

latest_url=https://api.github.com/repos/extism/cli/releases/latest
download_url=https://github.com/extism/cli/releases/download
out_prefix=/usr/local/bin
ask=y
quiet=n
version=""

# Determine default arch/os
os=$(uname -s | awk '{print tolower($0)}')
arch=$(uname -m)

usage() {
  echo "get-extism.sh - A script to help fetch the Extism CLI"
  echo ""
  echo "Flags"
  echo "-----"
  echo "  -h: print this usage message"
  echo "  -a: set the target machine architecture"
  echo "  -s: set the target operating system"
  echo "  -o: installation prefix (default: /usr/local/bin)"
  echo "  -y: don't ask before executing commands"
  echo "  -q: don't print updated to stdout"
}

latest_tag() {
  curl -s $latest_url | grep tag_name | awk '{ print $2 }' | sed 's/[",]//g'
}

untar() {
  print "Extracting release"
  _sudo=sudo
  case $out_prefix in
  $HOME/*)
    _sudo=""
    ;;
  esac
  $_sudo sh -c "tar -xzO extism > $out_prefix/extism"
  $_sudo chmod +x $out_prefix/extism
}

print() {
  if [ "$quiet" = "n" ]; then
    echo $@
  fi
}

err() {
  echo $@ >&2
  exit 1
}

while getopts "hyqa:s:o:v:" arg "$@"; do
    case "$arg" in
        h)
            usage
            exit 0
            ;;
        y)
            ask=n
            ;;
        q)
            quiet=y
            ;;
        a)
            arch=$OPTARG
            ;;
        s)
            os=$OPTARG
            ;;
        o)
            out_prefix=$OPTARG
            ;;
        v)
            version=$OPTARG
            ;;
        *)
            ;;
        esac
done

# Fix arch names
case "$arch" in
x86_64)
  arch=amd64 ;;
aarch64)
  arch=arm64 ;;
*)
  ;;
esac

# Get latest version if none was specified
if test -z $version
then
  version=`latest_tag`
fi

if [ "$ask" = "y" ] && [ ! -t 0 ]; then
    if [ ! -t 1 ]; then
        err "Unable to run interactively. Run with -y to accept defaults, -h for additional options"
    fi
fi


if [ "$ask" = "y" ]; then
  echo "Confirm installation:"
  echo "  Version: $version"
  echo "  OS: $os"
  echo "  Arch: $arch"
  echo "  Destination: $out_prefix/extism"
  echo -n "Proceed? [y, n]: "
  read reply < /dev/tty
else
  reply=y
fi
if [ "$reply" = "y" ]; then
  curl -L -s "$download_url/$version/extism-$version-$os-$arch.tar.gz" | untar
  print "extism executable installed to $out_prefix/extism"
else
  err "Exiting"
fi

