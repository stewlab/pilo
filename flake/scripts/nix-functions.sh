#!/usr/bin/env bash

# Function to check if running on NixOS
is_nixos() {
  [ -f /run/current-system/sw/bin/nixos-rebuild ]
}

# Function to check if sudo is needed for Nix commands
needs_sudo() {
  [ -e /nix/var/nix/daemon-socket/socket ]
}