{
  description = "A basic Rust development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            rustc
            cargo
            rustup
            pkg-config
            fzf
            neovim
            lunarvim
          ];
          shellHook = ''
            echo "🦀 Entering 'rustStableEnv' (Rust stable)"
            echo ""
            rustc --version
          '';
        };
      });
}