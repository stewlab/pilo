#!/usr/bin/env bash
set -e

echo "Searching packages in nixpkgs..."
nix search nixpkgs "$@"