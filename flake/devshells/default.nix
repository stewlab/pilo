# ./devshells/default.nix
{ pkgs, unstablePkgs, lib, ... }:
let
  # Path to the current directory
  devshellsPath = ./.;

  # Read all .nix files in the directory, excluding this file
  devshellFiles = lib.filter (path: path != "default.nix") (builtins.attrNames (builtins.readDir devshellsPath));

  # Safely import each dev shell file, skipping any that fail to evaluate.
  importDevShell = file:
    let
      name = lib.removeSuffix ".nix" file;
      path = "${devshellsPath}/${file}";
      # Try to evaluate the devshell import.
      result = builtins.tryEval (import path { inherit pkgs unstablePkgs lib; });
    in
    # If the import fails, warn the user and return null.
    if !result.success then
      lib.warn "Failed to import devshell '${name}' from ${path}. Skipping." null
    # If it succeeds, create the attribute for the shell.
    else
      lib.nameValuePair name result.value;

  # Create an attribute set of all dev shells that were successfully imported.
  importedDevshells =
    let
      # Map over the files and try to import each one.
      maybeDevshells = map importDevShell devshellFiles;
      # Filter out the ones that failed (are null).
      validDevshells = lib.filter (s: s != null) maybeDevshells;
    in
    lib.listToAttrs validDevshells;

  # Define the default shell here to ensure it always exists and is pure.
  defaultShell = pkgs.mkShell {
    name = "default-nix-shell";
    buildInputs = with pkgs; [
      git
    ];
    shellHook = ''
      echo "👋 Welcome to the default development shell."
      echo "You can enter other shells by name, e.g., 'nix-develop rustico'"
    '';
  };

in
importedDevshells // {
  default = defaultShell;
}