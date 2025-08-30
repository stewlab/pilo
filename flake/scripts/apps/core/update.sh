#!/usr/bin/env bash
set -e

FLAKE_DIR="@FLAKE_DIR@"

echo "Updating flake inputs..."
nix flake update "$FLAKE_DIR" "$@"