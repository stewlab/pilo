#!/usr/bin/env bash
set -e


if is_nixos; then
  echo "Running garbage collection (system-wide)..."
  if needs_sudo; then
    sudo nix-collect-garbage -d
  else
    nix-collect-garbage -d
  fi
else
  echo "Running garbage collection (user)..."
  nix-collect-garbage -d
fi