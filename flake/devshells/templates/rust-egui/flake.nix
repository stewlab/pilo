{
  description = "A Rust development environment for egui GUI applications";

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
          buildInputs = with pkgs; [
            git
            pkg-config
            cmake
            zsh

            rustc
            cargo
            rust-analyzer
            clippy
            rustfmt

            ffmpeg
            glib
            dbus

          ] ++ pkgs.lib.optionals pkgs.stdenv.isLinux [
            pkgs.alsa-lib
            pkgs.libjack2
            pkgs.wayland
            pkgs.xorg.libX11
            pkgs.xorg.libXcursor
            pkgs.xorg.libXi
            pkgs.xorg.libXrandr
            pkgs.libxkbcommon
            pkgs.mesa
            pkgs.libglvnd
            pkgs.at-spi2-core
            pkgs.xorg.libxcb
            pkgs.xorg.libXinerama
            pkgs.xorg.libXext
            pkgs.xorg.libXfixes
            pkgs.xorg.libXdamage
            pkgs.xorg.libXcomposite
            pkgs.gst_all_1.gstreamer
            pkgs.gst_all_1.gst-plugins-base
            pkgs.gst_all_1.gst-plugins-good
            pkgs.gst_all_1.gst-plugins-bad
            pkgs.gst_all_1.gst-plugins-ugly
            pkgs.gst_all_1.gst-libav
            pkgs.gst_all_1.gst-vaapi

          ] ++ pkgs.lib.optionals pkgs.stdenv.isDarwin [
            pkgs.libiconv
            pkgs.darwin.apple_sdk.frameworks.AppKit
            pkgs.darwin.apple_sdk.frameworks.CoreAudio
            pkgs.darwin.apple_sdk.frameworks.CoreGraphics
            pkgs.darwin.apple_sdk.frameworks.OpenGL

          ] ++ (with pkgs; [
            libvorbis
            libopus
            flac
          ]);

          nativeBuildInputs = with pkgs; [
            fzf
            vscodium
            neovim
          ];

          shellHook = ''
            export FLAC_LIB_DIR="${pkgs.flac.out}/lib"
            export FLAC_INCLUDE_DIR="${pkgs.flac.dev}/include"

            export LD_LIBRARY_PATH="${pkgs.lib.makeLibraryPath (with pkgs; [
              wayland
              libxkbcommon
              mesa
              dbus
              at-spi2-core
              libglvnd
              xorg.libxcb
              xorg.libX11
              xorg.libXcursor
              xorg.libXi
              xorg.libXrandr
              xorg.libXinerama
              xorg.libXext
              xorg.libXfixes
              xorg.libXdamage
              xorg.libXcomposite
            ])}:$LD_LIBRARY_PATH"

            echo ""
            echo "🦀 Entering Rust development environment..."
            echo "----------------------------------------"
            rustc --version
            cargo --version
            echo "----------------------------------------"
            echo "✅ Environment ready!"
            echo ""
          '';
        };
      });
}