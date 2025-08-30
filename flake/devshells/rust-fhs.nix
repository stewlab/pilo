{ pkgs, ... }:

pkgs.buildFHSEnv {
  name = "rustico-fhs";
  # targetPkgs = ps: with ps; [
  targetPkgs = pkgs: with pkgs; [
    # Core tools
    git
    pkg-config
    cmake
    zsh
    # Rust toolchain
    rustc
    cargo
    rust-analyzer
    clippy
    rustfmt
    # Audio libraries
    alsa-lib
    libvorbis
    libopus
    flac
    libjack2
    # FFmpeg
    ffmpeg
    # GStreamer
    gst_all_1.gstreamer
    gst_all_1.gst-plugins-base
    gst_all_1.gst-plugins-good
    gst_all_1.gst-plugins-bad
    gst_all_1.gst-plugins-ugly
    gst_all_1.gst-libav
    gst_all_1.gst-vaapi
    glib
    # GUI dependencies
    wayland
    libxkbcommon
    mesa
    dbus
    at-spi2-core
    libglvnd
    xorg.libxcb # X C Binding library
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
    # Misc
    fzf
    nerd-fonts.jetbrains-mono
    vscodium
    neovim
  ];

  profile = ''
    # --- Set Zsh theme ---
    # export ZSH_THEME="jtriley"

    # --- Configure for rust-analyzer ---
    # This provides the path to the standard library source code.
    export RUST_SRC_PATH="${pkgs.rustPlatform.rustLibSrc}"
    
    # --- Rust toolchain is managed by Nix ---

    # --- Welcome Message ---
    echo "🦀 Entering 'rustico-fhs' FHS environment"
    echo ""
    rustc --version
    cargo --version
  '';

  runScript = "zsh";
}