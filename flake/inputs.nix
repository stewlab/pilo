# ./inputs.nix
{
  nixpkgs = {
    url = "github:NixOS/nixpkgs/nixos-25.05";
  };
  nixpkgs-unstable = {
    url = "github:NixOS/nixpkgs/nixos-unstable";
  };
  home-manager = {
    url = "github:nix-community/home-manager/release-25.05";
    inputs.nixpkgs.follows = "nixpkgs";
  };
}