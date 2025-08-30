{ pkgs, ... }:

pkgs.mkShell {
  packages = with pkgs; [
    rustc
    cargo
    rustup
    pkg-config
    fzf
    nerd-fonts.jetbrains-mono
    neovim
    lunarvim
  ];
  shellHook = ''
    # --- Set Zsh theme ---
    # export ZSH_THEME="jtriley"
    # --- Welcome Message ---
    echo "🦀 Entering 'rustStableEnv' (Rust stable)"
    echo ""
    rustc --version
  '';
}