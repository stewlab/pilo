{ pkgs, ... }:

pkgs.mkShell {
  packages = with pkgs; [
    python312
    uv
    poetry
    fzf
    nerd-fonts.jetbrains-mono
    neovim
    lunarvim
  ];
  shellHook = ''
    # --- Set Zsh theme ---
    # export ZSH_THEME="jtriley"
    # --- Welcome Message ---
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
}