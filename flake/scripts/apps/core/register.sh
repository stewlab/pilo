#!/usr/bin/env bash
set -e

FLAKE_DIR="@FLAKE_DIR@"

# Use the first argument as the flake name, defaulting to 'config'
FLAKE_NAME=${1:-config}

echo "Adding flake to registry as '$FLAKE_NAME'..."
nix registry add "$FLAKE_NAME" "path:$FLAKE_DIR"

echo "Successfully registered '$FLAKE_NAME' to '$FLAKE_DIR'."
echo "You can now use commands like 'nix run $FLAKE_NAME#rebuild'."