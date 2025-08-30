{ pkgs, username, ... }:

{
  # Define a user account.
  users.users.${username} = {
    isNormalUser = true;
    description = "";
    extraGroups = [ "networkmanager" "wheel" ];
    shell = pkgs.zsh;
  };
}