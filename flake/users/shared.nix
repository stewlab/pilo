# ./users/shared.nix
# This file contains shared settings for all users.
{ pkgs, ... }:

{
  # Shared settings can go here. For example, you might want to
  # enforce a common set of tools or shell settings.
  home.stateVersion = "25.05";

  programs.zsh = {
    enable = true;
    autosuggestion.enable = true;
    enableCompletion = true;
    syntaxHighlighting.enable = true;
    oh-my-zsh = {
      enable = true;
      theme = "bureau";
      plugins = [ "colored-man-pages" "common-aliases" "git" "fzf" ];
    };
    initContent = "";
  };

  programs.git = {
    enable = true;
  };
}