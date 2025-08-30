# ./apps/default.nix
{ pkgs, unstablePkgs, lib, self, ... }:

let
  # --- App Imports ---

  # Helper function to import a Nix-based app from a file.
  # This takes a file path, imports it, and formats it as a Flake app.
  importNixApp = file:
    let
      # Derives the app name from the filename (e.g., "claude-code.nix" -> "claude-code").
      name = lib.removeSuffix ".nix" file;
      # The actual package is defined in the imported file.
      package = import ./${file} { inherit pkgs unstablePkgs; };
    in
    lib.nameValuePair name {
      type = "app";
      # The `program` attribute points to the executable inside the built package.
      # The specific binary name `claude-code` is assumed based on the package.
      program = "${package}/bin/${name}";
    };

  # --- App Discovery ---

  # Scan the current directory for all `.nix` files, excluding `default.nix`.
  # These files are assumed to be app definitions.
  nixAppFiles = lib.filter (file: file != "default.nix" && lib.hasSuffix ".nix" file) (builtins.attrNames (builtins.readDir ./.));

  # --- App Sets ---

  # Create an attribute set of Nix-based applications.
  # `listToAttrs` converts the list of apps into a set where keys are app names.
  nixApps = lib.listToAttrs (map importNixApp nixAppFiles);

in

# --- Final Apps Output ---
# The final output merges the Nix-based apps and script-based apps into a single set.
# This makes all discovered applications available in the flake's `apps` output.
nixApps // {
  default = {
    type = "app";
    program = "${self.packages.${pkgs.system}.pilo}/bin/pilo";
  };
}
