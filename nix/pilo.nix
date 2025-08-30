# ./nix/pilo.nix
{ pkgs, ... }:

pkgs.buildGoModule {
  pname = "pilo";
  version = "0.1.0";
  src = ../.;
  vendorHash = "sha256-Oy0hlui2o79NU1FVTkiRSUv417CGJ9FVpJxydDSQYkE=";
  nativeBuildInputs = with pkgs; [
    pkg-config
  ];
  buildInputs = with pkgs; [
    # GUI & Audio Libs from devshell
    mesa
    libglvnd
    vulkan-loader
    libxkbcommon
    wayland
    xorg.libX11
    xorg.libXcursor
    xorg.libXrandr
    xorg.libXinerama
    xorg.libXi
    xorg.libXxf86vm
    xorg.libXext
    xorg.libXfixes
    xorg.libXdamage
    xorg.libXcomposite
    at-spi2-core
    xorg.libxcb
    portaudio
    alsa-lib
  ];
}