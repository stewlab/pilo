#!/usr/bin/env bash
set -e

# Source the nix-functions.sh script to get the is_nixos function
source "$(dirname "$0")/../../nix-functions.sh"

if is_nixos; then
  echo "Rolling back NixOS system..."
  sudo nixos-rebuild switch --rollback
else
  echo "Rolling back user profile..."
  nix profile rollback
fi