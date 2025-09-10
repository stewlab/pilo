# ./packages/default.nix
{ pkgs, unstablePkgs, self, inputs, ... }:
let
  # Create a wrapper script for a given app name
  mkWrapper = appName: pkgs.writeShellScriptBin "nix-${appName}" ''
    #!${pkgs.runtimeShell}
    set -e
    nix run path:${self}#${appName} -- "$@"
  '';

  # Get all the app names from the flake's apps output
  appNames = builtins.attrNames self.apps.${pkgs.system};

  # Read config from JSON. This makes packages manageable via the pilo API.
  config = builtins.fromJSON (builtins.readFile ../packages.json);

  # Get package names from the JSON, filtering for installed packages.
  packageNames = map (pkg: pkg.name) (builtins.filter (pkg: pkg.installed) config.packages);

  # Function to resolve a package path from multiple sources (nixpkgs, inputs)
  resolvePackage = pathStr:
    let
      path = pkgs.lib.splitString "." pathStr;
      # Check if the first element of the path is a known input
      isInput = pkgs.lib.hasAttr (builtins.head path) inputs;
      # Define the base object to search in
      base = if isInput then inputs else pkgs;
      # Define the search path
      searchPath = if isInput then path else path;
    in
    pkgs.lib.getAttrFromPath searchPath base;

  # Convert package names to derivations, allowing for nested attrpaths like "kdePackages.kcalc" or "my-flake.packages.my-package".
  defaultPackages = map resolvePackage packageNames;
in
(
let
  # Function to import a package from a file
  importPackage = file: {
    name = builtins.replaceStrings [ ".nix" ] [ "" ] file;
    value = import ./${file} { inherit pkgs; pnpm = unstablePkgs.pnpm_9; };
  };

  # Read all files in the current directory, filter for .nix files (excluding default.nix)
  packageFiles = builtins.filter (file: file != "default.nix" && pkgs.lib.strings.hasSuffix ".nix" file) (builtins.attrNames (builtins.readDir ./.));

  # Create an attribute set of packages
  packages = builtins.listToAttrs (map importPackage packageFiles);
in
packages // {
  # Dynamically create wrappers for all apps
  command-wrappers = pkgs.buildEnv {
    name = "nix-command-wrappers";
    paths = map mkWrapper appNames;
  };

  # Default system packages as a derivation
  default = pkgs.buildEnv {
    name = "system-packages";
    paths = defaultPackages;
  };

  # Expose the raw list of packages
  default-list = defaultPackages;
}
)