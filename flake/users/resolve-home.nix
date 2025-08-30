# This function checks if a home.nix file exists for a given user.
# If it exists, it returns the path to the file.
# If it does not exist, it returns an empty list, which is a safe
# value to import in Nix.
path: user:
let
  homeFile = path + "/${user}/home.nix";
in
  if builtins.pathExists homeFile then
    homeFile
  else
    # Returning an empty list is a neat trick to avoid errors when the file doesn't exist.
    # Importing an empty list has no effect.
    []