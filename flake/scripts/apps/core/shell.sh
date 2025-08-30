#!/usr/bin/env bash
set -e

if [ -z "$1" ]; then
  echo "Usage: shell <pkgname> [pkgname...]"
  echo "Enters a temporary shell with the specified packages."
  exit 1
fi

echo "Entering a temporary shell with the requested packages..."
nix-shell -p "$@"