{ pkgs, lib, ... }:

pkgs.mkShell {
  # buildInputs is the more idiomatic name for packages that provide libraries/headers
  buildInputs = with pkgs; [
    # --- Core Tools ---
    git
    pkg-config
    cmake
    zsh

    # --- Rust Toolchain ---
    # The Rust toolchain is managed by Nix.
    rustc
    cargo
    rust-analyzer
    clippy
    rustfmt

    # --- Cross-Platform GUI & System Libs ---
    # These are generally safe on both Linux and macOS
    ffmpeg
    glib
    dbus

    # --- Linux-only GUI & Audio Libs ---
  ] ++ pkgs.lib.optionals pkgs.stdenv.isLinux [
    # Audio
    pkgs.alsa-lib
    pkgs.libjack2
    # GUI
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
    # GStreamer (Linux is more common for this)
    pkgs.gst_all_1.gstreamer
    pkgs.gst_all_1.gst-plugins-base
    pkgs.gst_all_1.gst-plugins-good
    pkgs.gst_all_1.gst-plugins-bad
    pkgs.gst_all_1.gst-plugins-ugly
    pkgs.gst_all_1.gst-libav
    pkgs.gst_all_1.gst-vaapi

  # --- macOS-only Frameworks ---
  ] ++ pkgs.lib.optionals pkgs.stdenv.isDarwin [
    pkgs.libiconv # often needed on macOS
    # Add required macOS frameworks here for GUI/audio
    pkgs.darwin.apple_sdk.frameworks.AppKit
    pkgs.darwin.apple_sdk.frameworks.CoreAudio
    pkgs.darwin.apple_sdk.frameworks.CoreGraphics
    pkgs.darwin.apple_sdk.frameworks.OpenGL

  # --- Common Audio libs (available on both platforms) ---
  ] ++ (with pkgs; [
    libvorbis
    libopus
    flac
  ]);

  # Native build inputs are for tools used during the build, like IDEs/editors
  nativeBuildInputs = with pkgs; [
    fzf
    nerd-fonts.jetbrains-mono
    vscodium
    neovim
    # lunarvim # Note: LunarVim often requires its own installation script
  ];

  shellHook = ''
    # --- Set library paths for build scripts ---
    # Note: Often not needed, as pkg-config is usually sufficient.
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

    # --- Welcome Message ---
    echo ""
    echo "🦀 Entering Rust development environment..."
    echo "----------------------------------------"
    rustc --version
    cargo --version
    echo "----------------------------------------"
    echo "✅ Environment ready!"
    echo ""
  '';
}