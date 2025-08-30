# Pilo's Nix Flake
{
  description = "A dynamic and modular NixOS configuration managed by Pilo";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
    nixpkgs-unstable.url = "github:NixOS/nixpkgs/nixos-unstable";
    home-manager = {
      url = "github:nix-community/home-manager/release-25.05";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, home-manager, nixpkgs-unstable, ... } @ inputs:
    let
      # Read dynamic inputs from the JSON file
      config = builtins.fromJSON (builtins.readFile ./base-config.json);
      system = lib.attrByPath [ "system" "type" ] "x86_64-linux" config;
      username = let
        raw = lib.attrByPath [ "system" "username" ] "" config;
      in if raw == "" then (builtins.head config.users).username else raw;

      desktop = lib.attrByPath [ "system" "desktop" ] null config;
      gitUrls = map (pkg: pkg.name) (builtins.filter (pkg: pkg.installed && pkgs.lib.strings.hasPrefix "github:" pkg.name) config.packages);
      dynamicInputs = builtins.listToAttrs (map (url:
        let
          # Extract repo name from URL to use as the attribute name
          repoName = builtins.elemAt (builtins.split "/" (builtins.replaceStrings [".git"] [""] url)) ((builtins.length (builtins.split "/" url)) - 1);
        in
        { name = repoName; value = { url = url; flake = false; }; }
      ) gitUrls);

      # Combine static and dynamic inputs
      allInputs = inputs // dynamicInputs;

      pkgs = import nixpkgs {
        inherit system;
        config.allowUnfree = true;
        overlays = [
          (final: prev: {
            electron = unstablePkgs.electron;
          })
        ];
      };

      unstablePkgs = import nixpkgs-unstable {
        inherit system;
        config.allowUnfree = true;
      };

      lib = nixpkgs.lib;
      # userConfigurations = import ./users { inherit lib; };
      userConfigurations = import ./users { inherit lib; config = { inherit username; }; };


      specialArgs = {
        inputs = allInputs;
        inherit unstablePkgs self;
      };

      packagesSet = import ./packages { inherit pkgs unstablePkgs self; inputs = allInputs; };

    in
    {
      homeConfigurations = lib.mapAttrs' (
        username: userConfig: lib.nameValuePair "${username}@${system}" (
          home-manager.lib.homeManagerConfiguration {
            inherit pkgs;
            extraSpecialArgs = specialArgs // { inherit username; };
            modules = [
              userConfig
              { home.username = username; home.homeDirectory = "/home/${username}"; }
            ];
          }
        )
      ) userConfigurations.users;

      apps.${system} = import ./apps { inherit pkgs unstablePkgs lib self; };

      packages.${system} = (builtins.removeAttrs packagesSet [ "default" "default-list" ]) // { pilo = packagesSet.pilo; };

      devShells.${system} = import ./devshells {
        inherit pkgs unstablePkgs lib;
      };

      nixosConfigurations = {
        nixos = lib.nixosSystem {
          inherit system;
          specialArgs = { inherit username; systemPackages = packagesSet.default-list; piloConfig = config; } // specialArgs;
          modules = [
            ./hosts/nixos/default.nix
            home-manager.nixosModules.home-manager
            {
              pilo.ollama.modelsPath = lib.attrByPath [ "system" "ollama" "models" ] "" config;
              home-manager.useGlobalPkgs = true;
              home-manager.useUserPackages = true;
              home-manager.extraSpecialArgs = { inherit pkgs unstablePkgs self; };
              home-manager.users.${username} = userConfigurations.home-manager.users.${username};
            }
          ] ++ (lib.optionals (desktop != null && desktop != "") [ ./desktops/${desktop}.nix ]);
        };
      };
    };
}
