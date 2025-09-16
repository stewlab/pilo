{
  description = "A Rust FHS development environment for compatibility";

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
        devShells.default = pkgs.buildFHSEnv {
          name = "rustico-fhs";
          targetPkgs = pkgs: with pkgs; [
            git
            pkg-config
            cmake
            zsh
            rustc
            cargo
            rust-analyzer
            clippy
            rustfmt
            alsa-lib
            libvorbis
            libopus
            flac
            libjack2
            ffmpeg
            gst_all_1.gstreamer
            gst_all_1.gst-plugins-base
            gst_all_1.gst-plugins-good
            gst_all_1.gst-plugins-bad
            gst_all_1.gst-plugins-ugly
            gst_all_1.gst-libav
            gst_all_1.gst-vaapi
            glib
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
            fontconfig
            freetype
            fzf
            vscodium
            neovim
          ];

          profile = ''
            export RUST_SRC_PATH="${pkgs.rustPlatform.rustLibSrc}"
            
            echo "🦀 Entering 'rustico-fhs' FHS environment"
            echo ""
            rustc --version
            cargo --version
          '';

          runScript = "zsh";
        };
      });
}