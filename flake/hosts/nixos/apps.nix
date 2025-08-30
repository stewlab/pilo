{ systemPackages, ... }:

{
  # Core system-wide packages.
  environment.systemPackages = systemPackages;

  # Enable ZSH
  programs.zsh.enable = true;

  # Enable Steam
  programs.steam.enable = true;
}