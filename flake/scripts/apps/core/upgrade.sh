#!/usr/bin/env bash
set -e

FLAKE_DIR="@FLAKE_DIR@"

if is_nixos; then
  echo "Upgrading NixOS system..."
  sudo nixos-rebuild switch --upgrade --flake "$FLAKE_DIR" "$@"
else
  echo "Upgrading user profile..."
  nix profile upgrade "$@"
fi