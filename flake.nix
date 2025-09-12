{
  description = "Pilo application Nix flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable"; # Use unstable for newer packages
  };

  outputs = { self, nixpkgs, ... }@inputs:
    let
      # List of supported systems
      supportedSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];

      # Helper function to generate outputs for each system
      forAllSystems = f: nixpkgs.lib.genAttrs supportedSystems (system: f {
        pkgs = import nixpkgs {
          inherit system;
          config.allowUnfree = true;
        };
      });
    in
    {
      # Generate packages for each supported system
      packages = forAllSystems ({ pkgs }: {
        default = import ./nix/pilo.nix { inherit pkgs; };
      });

      # Generate devShells for each supported system
      devShells = forAllSystems ({ pkgs }: {
        default = pkgs.mkShell {
          buildInputs = [ pkgs.go pkgs.nix ];
        };
      });
    };
}