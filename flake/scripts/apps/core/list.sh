#!/usr/bin/env bash
set -e

if is_nixos; then
  echo "--- System Packages (from configuration.nix) ---"
  nix-instantiate --eval -E 'with import <nixpkgs/nixos> { configuration = /etc/nixos/configuration.nix; }; builtins.map (p: p.name) config.environment.systemPackages' | xargs -n 1
  echo ""
  echo "--- User Packages ---"
  nix profile list
else
  echo "--- User Packages ---"
  nix profile list
fi