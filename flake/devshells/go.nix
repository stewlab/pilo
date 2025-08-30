{ pkgs, ... }:

pkgs.mkShell {
  packages = with pkgs; [
    # --- Core Tools ---
    go_1_23
    gopls
    delve
    pkg-config

    # --- Linux-only GUI & Audio Libs ---
    # For go-gl/glfw
    mesa
    libglvnd
    libglvnd.dev
    vulkan-loader
    libxkbcommon
    wayland
    wayland-protocols
    xorg.libX11
    xorg.libX11.dev
    xorg.libXcursor
    xorg.libXrandr
    xorg.libXinerama
    xorg.libXi
    xorg.libXxf86vm
    xorg.libXext
    xorg.libXfixes
    xorg.libXdamage
    xorg.libXcomposite
    xorg.xorgproto
    at-spi2-core
    xorg.libxcb
    # audio tracker (tico)
    portaudio
    alsa-lib
  ];

  nativeBuildInputs = with pkgs; [
     fzf
     nerd-fonts.jetbrains-mono
     nerd-fonts.go-mono
     neovim
     lunarvim
   ];

   shellHook = ''
    export LD_LIBRARY_PATH="${pkgs.lib.makeLibraryPath (with pkgs; [
      wayland
      libxkbcommon
      mesa
      libglvnd
      at-spi2-core
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

    # --- Set Zsh theme ---
    # export ZSH_THEME="jtriley"
    # --- Welcome Message ---
    echo "🚀 Entering 'goEnv' dev shell (Go 1.23, Stable Nixpkgs)"
    echo ""
  '';
}