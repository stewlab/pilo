# Pilo

A command-line tool for managing your Nix environment.

## Build

First, clone the repository:

```bash
git clone https://github.com/stewlab/pilo.git
cd pilo
```

Then, to build the `pilo` application, run the following command:

```bash
nix develop ./flake#go --command go build -o bin/pilo .
```

This command enters a Nix development shell that has Go installed, and then it builds the Pilo binary, placing it in the `bin` directory.

## Installation

To install Pilo, you can use `nix profile`:

```bash
nix profile install github:stewlab/pilo
```

This will install the `pilo` binary to your user profile, making it available in your shell.

## Testing

You can test Pilo without installing it by running it in a temporary shell:

```bash
nix shell github:stewlab/pilo
```

This will make the `pilo` command available in your current shell session.

## Usage

Pilo provides several commands to manage your Nix environment.

### System & Configuration Management

-   `pilo install`: Installs and configures the Pilo flake on your system.
-   `pilo rebuild`: Rebuilds your NixOS or Home Manager configuration.
-   `pilo update [input]`: Updates flake inputs. Optionally updates a single [input].
-   `pilo rollback`: Rolls back to the previous generation.
-   `pilo gc`: Runs the garbage collector to free up disk space.
-   `pilo list [packages|generations]`: Lists installed packages or system generations.
-   `pilo backup`: Creates a backup of the current Pilo configuration.
-   `pilo restore`: Restores the Pilo configuration from a remote Git repository.

### Package Management

-   `pilo install-pkg [pkg...]`: Installs packages to your user profile (non-NixOS) or provides a temporary shell (NixOS).
-   `pilo remove [pkg...]`: Removes packages from your user profile (non-NixOS).
-   `pilo upgrade`: Upgrades all packages in your user profile or NixOS system.
-   `pilo search [query]`: Searches for packages in `nixpkgs`.
-   `pilo add-app [pname] [version] [url] [sha256]`: Adds a new Flake App to your flake/packages directory.

### Development Shells

-   `pilo shell [pkg...]`: Enters a temporary shell with the specified packages.
-   `pilo develop [shell]`: Enters a persistent development shell from the flake.
-   `pilo devshell add [name]`: Adds a new development shell.
-   `pilo devshell remove [name]`: Removes a development shell.
-   `pilo devshell enter [name]`: Enters a development shell.
-   `pilo devshell run [name] [command]`: Runs a command in a development shell.

### GUI

-   `pilo gui`: Launches the Fyne GUI for Pilo.

### Other Commands

-   `pilo completion [bash|zsh|fish|powershell]`: Generate completion script for your shell.