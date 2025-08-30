#!/usr/bin/env bash
set -e


FLAKE_DIR="@FLAKE_DIR@"

if is_nixos; then
  echo "Rebuilding NixOS system..."
  sudo nixos-rebuild switch --flake "$FLAKE_DIR" "$@"
else
  echo "Applying Home Manager configuration..."
  home-manager switch --flake "$FLAKE_DIR" "$@"
fi