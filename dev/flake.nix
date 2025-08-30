{
  description = "A reliable Go development shell for Fyne applications";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs {
        inherit system;
        config.allowUnfree = true;
      };
    in
    {
      devShells.${system}.default = pkgs.mkShell {
        buildInputs = with pkgs; [
          # --- Core Go Tools ---
          go_1_23
          gopls
          delve
          pkg-config

          # --- System & GUI Libs (from your working Rust shell) ---
          dbus
          glib
          # Audio
          alsa-lib
          # GUI
          wayland
          xorg.libX11
          xorg.libXcursor
          xorg.libXi
          xorg.libXrandr
          libxkbcommon
          mesa
          libglvnd
          at-spi2-core
          xorg.libxcb
          xorg.libXinerama
          xorg.libXxf86vm
          xorg.libXext
          xorg.libXfixes
          xorg.libXdamage
          xorg.libXcomposite
        ];

        nativeBuildInputs = with pkgs; [
          fzf
          nerd-fonts.jetbrains-mono
        ];

        shellHook = ''
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
            xorg.libXxf86vm
            xorg.libXext
            xorg.libXfixes
            xorg.libXdamage
            xorg.libXcomposite
          ])}:$LD_LIBRARY_PATH"

          # --- Welcome Message ---
          echo ""
          echo "✅ Entering Go development environment..."
          echo "----------------------------------------"
          go version
          echo "----------------------------------------"
          echo "✅ Environment ready! You can now use 'go run .'"
          echo ""
        '';
      };
    };
}