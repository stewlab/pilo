{
  description = "A Go development environment";

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
          packages = with pkgs; [
            go_1_23
            gopls
            delve
            pkg-config

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
            portaudio
            alsa-lib
          ];

          nativeBuildInputs = with pkgs; [
             fzf
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

            echo "🚀 Entering 'goEnv' dev shell (Go 1.23, Stable Nixpkgs)"
            echo ""
          '';
        };
      });
}