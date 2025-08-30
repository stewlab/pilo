# nix-fhs.nix

{ pkgs, ... }:

pkgs.buildFHSEnv {
  name = "nix-fhs";
  targetPkgs = pkgs: with pkgs; [ curl bash coreutils gnugrep ];
  runScript = "bash";
}