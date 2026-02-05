{
  description = "A Python development environment using uv and poetry";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
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
            python312
            uv
            poetry
            fzf
            neovim
            lunarvim
          ];
          shellHook = ''
            echo "🐍 Entering 'pythonEnv' dev shell (Python 3.12, Stable Nixpkgs)"
            if [ ! -d ".venv" ]; then
              echo "Creating Python virtual environment..."
              uv venv
            fi
            source .venv/bin/activate
            if [ -f "requirements.txt" ]; then
              echo "Installing dependencies from requirements.txt..."
              uv pip install -r requirements.txt
            elif [ -f "pyproject.toml" ]; then
              echo "Installing dependencies with Poetry/uv..."
              uv install
            fi
            echo ""
          '';
        };
      });
}