{
  description = "Pilo application Nix flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable"; # Use unstable for newer packages
  };

  outputs = { self, nixpkgs, ... }@inputs:
    let
      # List of supported systems
      supportedSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];

      # Helper function to generate outputs for each system
      forAllSystems = f: nixpkgs.lib.genAttrs supportedSystems (system: f {
        pkgs = import nixpkgs {
          inherit system;
          config.allowUnfree = true;
        };
      });
    in
    {
      # Generate packages for each supported system
      packages = forAllSystems ({ pkgs }: {
        default = import ./nix/pilo.nix { inherit pkgs; };
      });

      # Generate devShells for each supported system
      devShells = forAllSystems ({ pkgs }: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            go_1_23
            gopls
            delve
            pkg-config
            nix
          ]
          ++ (pkgs.lib.optionals pkgs.stdenv.isLinux (with pkgs; [
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
          ]));

          nativeBuildInputs = with pkgs; [
            fzf
            neovim
          ];

          shellHook = pkgs.lib.optionalString pkgs.stdenv.isLinux ''
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
          '';
        };
      });
    };
}