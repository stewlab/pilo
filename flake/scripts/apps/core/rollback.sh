#!/usr/bin/env bash
set -e

if is_nixos; then
  echo "Rolling back NixOS system..."
  sudo nixos-rebuild switch --rollback
else
  echo "Rolling back user profile..."
  nix profile rollback
fi