{ inputs, username, systemPackages, lib, ... }:

{
  imports = [
    ./hardware-configuration.nix
    ./system.nix
    ./user.nix
    ./services.nix
    ./apps.nix
    ./fonts.nix
    # The desktop environment preset is imported from the top-level flake.nix
  ];

  options.pilo.ollama.modelsPath = lib.mkOption {
    type = lib.types.str;
    description = "The path to Ollama models.";
  };

  config = {
    # This value determines the NixOS release from which the default
    # settings for stateful data, like file locations and database versions
    # on your system were taken. It‘s perfectly fine and recommended to leave
    # this value at the release version of the first install of this system.
    # Before changing this value read the documentation for this option
    # (e.g. man configuration.nix or on https://nixos.org/nixos/options.html).
    system.stateVersion = "25.05"; # Did you read the comment?
  };
}