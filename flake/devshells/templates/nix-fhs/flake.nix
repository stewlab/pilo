{
  description = "A basic FHS environment for Nix";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.buildFHSEnv {
          name = "nix-fhs";
          targetPkgs = pkgs: with pkgs; [ curl bash coreutils gnugrep ];
          runScript = "bash";
        };
      });
}