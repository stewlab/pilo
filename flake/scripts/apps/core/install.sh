#!/usr/bin/env bash
set -e

if is_nixos; then
  echo "On NixOS, permanent packages should be added to configuration.nix."
  echo "Providing a temporary shell with the requested packages..."
  nix-shell -p "$@"
else
  echo "Installing packages with nix profile..."
  nix profile install "$@"
fi