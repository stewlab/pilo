{ pkgs, ... }:

{
  fonts.packages = with pkgs; [
    nerd-fonts._3270
    nerd-fonts._0xproto
    nerd-fonts.adwaita-mono
    nerd-fonts.agave
    nerd-fonts.go-mono
    nerd-fonts.jetbrains-mono
    nerd-fonts.proggy-clean-tt
    nerd-fonts.recursive-mono
    nerd-fonts.roboto-mono
    nerd-fonts.sauce-code-pro
    nerd-fonts.symbols-only
    nerd-fonts.ubuntu
    nerd-fonts.ubuntu-mono
    nerd-fonts.ubuntu-sans
  ];
}