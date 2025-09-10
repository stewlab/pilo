{ lib, ... }:

let
  # Import the base configuration from the JSON file
  usersConfig = builtins.fromJSON (builtins.readFile ../users.json);

  # Function to generate a home-manager configuration for a given user
  mkUser = user: {
    home.stateVersion = "25.05";
    home.username = user.username;
    home.homeDirectory = "/home/${user.username}";

    # Import shared settings
    imports = [ ./shared.nix ];

    programs.git = {
      enable = true;
      userName = user.name;
      userEmail = user.email;
    };

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
  };

  # Generate the user configurations from the list in base-config.json
  users = lib.listToAttrs (map (user: {
    name = user.username;
    value = mkUser user;
  }) usersConfig.users);

in
{
  home-manager.users = users;
}
