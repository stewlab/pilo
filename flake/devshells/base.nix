# ./devshells/base.nix
{ pkgs, ... }:

# This file defines the 'base' development shell.
# It is intentionally kept simple to ensure it always works as a fallback.
pkgs.mkShell {
  # name = "base-nix-shell";
  buildInputs = with pkgs; [
    git
  ];
  shellHook = ''
    echo "👋 Welcome to the base development shell."
  '';
}