#!/usr/bin/env bash
set -e

FLAKE_DIR="@FLAKE_DIR@"

# Default to the 'base' shell if no argument is provided.
SHELL_NAME=${1:-default}

echo "🚀 Entering development shell: ${SHELL_NAME}"

# Directly call 'nix develop' with the specified shell name.
# The flake path is implicitly the current directory, which is correct
# because the command-wrapper executes from the flake root.
# Any additional arguments are passed after the '--'.
# Only shift arguments if a shell name was explicitly provided.
if [ -n "$1" ]; then
  shift
fi
nix develop "${FLAKE_DIR}#${SHELL_NAME}" "$@"

echo "👋 Exited development shell: ${SHELL_NAME}"