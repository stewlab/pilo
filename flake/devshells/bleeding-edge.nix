{ unstablePkgs, ... }:

unstablePkgs.mkShell {
  packages = with unstablePkgs; [
    rustc
    cargo
    go
    gopls
    gcc
    python3
    uv
    fzf
    nerd-fonts.jetbrains-mono
    cc65
    neovim
    lunarvim
    godot
  ];
  shellHook = ''
    # --- Set Zsh Theme ---
    # export ZSH_THEME="imajes"
    # --- Welcome Message ---
    echo "⚡ Welcome to a bleeding-edge development environment (Unstable Nixpkgs)!"
  '';
}