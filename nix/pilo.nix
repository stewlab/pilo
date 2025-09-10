# ./nix/pilo.nix
{ pkgs, ... }:

pkgs.buildGoModule {
  pname = "pilo";
  version = "0.1.0";
  src = ../.;
  vendorSha256 = pkgs.lib.fakeSha256;
  nativeBuildInputs = with pkgs; [
    pkg-config
  ];
  buildInputs = with pkgs; if stdenv.isLinux then [
    # GUI & Audio Libs for Linux
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
  ] else if stdenv.isDarwin then [
    # GUI & Audio Libs for macOS
    frameworks.Cocoa
    frameworks.Security
    frameworks.SystemConfiguration
    frameworks.CoreAudio
    frameworks.AudioToolbox
    frameworks.CoreMIDI
    portaudio
  ] else [];
}