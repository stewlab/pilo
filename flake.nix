{
  description = "Pilo application Nix flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable"; # Use unstable for newer packages
  };

  outputs = { self, nixpkgs, ... }@inputs:
    let
      system = "x86_64-linux"; # Assuming x86_64-linux, adjust if needed
      pkgs = import nixpkgs {
        inherit system;
        config.allowUnfree = true;
      };
    in
    {
      packages.${system}.default = import ./nix/pilo.nix { inherit pkgs; };

      devShells.${system}.default = pkgs.mkShell {
        buildInputs = [ pkgs.go ];
      };
    };
}