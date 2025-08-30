#!/bin/sh

# This script installs the NixOS or Home Manager configuration
# based on the operating system, following platform-best-practices.

set -e

# --- Configuration ---
CONFIG_DIR=$(cd "$(dirname "$0")" && pwd)
BACKUP_DIR="$CONFIG_DIR/backups/config"
MODIFY_NIXOS_CONFIG=false

# --- Argument Parsing ---
while [ "$#" -gt 0 ]; do
    case "$1" in
        --modify-nixos-config)
            MODIFY_NIXOS_CONFIG=true
            shift 1
            ;;
        *)
            echo "Unknown argument: $1"
            exit 1
            ;;
    esac
done

# --- Helper Functions ---
ensure_flakes_for_nix_user() {
    echo "Ensuring Nix Flakes are enabled for the current user..."
    local user_nix_conf_dir="$HOME/.config/nix"
    local user_nix_conf_file="$user_nix_conf_dir/nix.conf"
    local flakes_line="experimental-features = nix-command flakes"

    if grep -q "experimental-features.*flakes" "$user_nix_conf_file" 2>/dev/null; then
        echo "Flakes are already enabled for this user."
    else
        echo "Enabling flakes in $user_nix_conf_file."
        mkdir -p "$user_nix_conf_dir"
        if [ -s "$user_nix_conf_file" ] && [ -n "$(tail -c 1 "$user_nix_conf_file")" ]; then
            echo "" >> "$user_nix_conf_file"
        fi
        echo "$flakes_line" >> "$user_nix_conf_file"
        echo "NOTE: You may need to restart your shell for changes to take effect."
    fi
}

backup_and_modify_nixos_config() {
    local config_file="/etc/nixos/configuration.nix"
    if [ ! -f "$config_file" ]; then
        echo "ERROR: $config_file not found."
        exit 1
    fi

    if grep -q 'nix.settings.experimental-features.*flakes' "$config_file"; then
        echo "Flakes already enabled in $config_file. No changes needed."
        return
    fi

    echo "Backing up $config_file..."
    mkdir -p "$BACKUP_DIR"
    local timestamp=$(date +%Y%m%d-%H%M%S)
    local backup_file="$BACKUP_DIR/configuration.nix.backup-$timestamp"
    
    cp "$config_file" "$backup_file"
    gzip "$backup_file"
    
    echo "Backup created at $backup_file.gz"

    echo "Adding flakes setting to $config_file (requires sudo)..."
    sudo sed -i '/^{/a \  nix.settings.experimental-features = [ "nix-command" "flakes" ];' "$config_file"
    echo "Successfully modified $config_file."
}


# --- Main Script ---
echo "Using configuration from: $CONFIG_DIR"
cd "$CONFIG_DIR"

if [ -f /etc/NIXOS ]; then
    echo "NixOS detected."
    if [ "$MODIFY_NIXOS_CONFIG" = true ]; then
        backup_and_modify_nixos_config
    else
        echo "--------------------------------------------------------------------"
        echo "INFO: To enable flakes automatically, run this script with:"
        echo "  --modify-nixos-config"
        echo "Or, manually add the following to /etc/nixos/configuration.nix:"
        echo '  nix.settings.experimental-features = [ "nix-command" "flakes" ];'
        echo "--------------------------------------------------------------------"
    fi
    
    echo "Applying NixOS configuration..."
    sudo nixos-rebuild switch --flake .#nixos
else
    echo "Non-NixOS Linux detected."
    ensure_flakes_for_nix_user
    
    echo "Applying Home Manager configuration for user 'nixuser'..."
    nix run home-manager/release-25.05 -- switch --flake .#"nixuser"@x86_64-linux
fi

echo "Installation complete."