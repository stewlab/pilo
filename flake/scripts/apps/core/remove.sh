#!/usr/bin/env bash
set -e

if is_nixos; then
  echo "On NixOS, packages should be removed from configuration.nix and then run 'rebuild'."
  exit 1
else
  echo "Removing packages with nix profile..."
  nix profile remove "$@"
fi