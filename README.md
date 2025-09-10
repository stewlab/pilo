# Pilo

A command-line tool for managing your Nix environment.

## Installation

Pilo offers two ways to manage your Nix environment. For a detailed explanation of the flake and its structure, see the [`flake/README.md`](flake/README.md).

### Full Installation (Pilo Binary + Flake)

This method installs the `pilo` binary, which includes an embedded `flake/` directory with version control. The Pilo application manages its own Nix configuration.

```bash
nix profile install github:stewlab/pilo
```

When you run `pilo` for the first time, it will create the necessary configuration files in `~/.config/pilo`, providing a managed Nix environment.

### Standalone Flake (Nix Configuration Only)

If you prefer to manage your Nix environment manually, you can use the `flake/` directory as a standalone Nix configuration. This approach is for users who are comfortable working directly with Nix commands.

To use the flake, build and apply it to your system using standard Nix commands:

-   **For NixOS**: `sudo nixos-rebuild switch --flake .#your-host`
    -   **Note**: You will need to replace `flake/hosts/nixos/hardware-configuration.nix` with the one from your system, typically located at `/etc/nixos/hardware-configuration.nix`.
-   **For Home Manager**: `home-manager switch --flake .#your-username`

This method gives you full control over the Nix configuration, allowing you to integrate it into your existing setup.

## Usage

Pilo provides several commands to manage your Nix environment.

### First-Time Setup

When you run `pilo` for the first time, it will automatically create the necessary configuration files in `~/.config/pilo`. You do not need to perform any manual setup.

### System & Configuration Management

-   `pilo install`: Installs and configures the Pilo flake on your system.
-   `pilo rebuild`: Rebuilds your NixOS or Home Manager configuration.
-   `pilo update [input]`: Updates flake inputs. Optionally updates a single [input].
-   `pilo rollback`: Rolls back to the previous generation.
-   `pilo gc`: Runs the garbage collector to free up disk space.
-   `pilo list [packages|generations]`: Lists installed packages or system generations.
-   `pilo backup`: Creates a backup of the current Pilo configuration.
-   `pilo restore`: Restores the Pilo configuration from a remote Git repository.
-   `pilo config set-nix-path [path]`: Sets the path to the Nix binary.

### Package Management

-   `pilo install-pkg [pkg...]`: Installs packages to your user profile (non-NixOS) or provides a temporary shell (NixOS).
-   `pilo remove [pkg...]`: Removes packages from your user profile (non-NixOS).
-   `pilo upgrade`: Upgrades all packages in your user profile or NixOS system.
-   `pilo search [query]`: Searches for packages in `nixpkgs`.
-   `pilo add-app [pname] [version] [url] [sha256]`: Adds a new Flake App to your flake/packages directory.

### Development Shells

-   `pilo shell [pkg...]`: Creates a temporary, ephemeral shell with the specified packages.
-   `pilo develop [shell]`: Enters a persistent development shell defined in your flake.
-   `pilo devshell add [name]`: Adds a new development shell definition.
-   `pilo devshell remove [name]`: Removes an existing development shell definition.
-   `pilo devshell enter [name]`: Enters the specified development shell.
-   `pilo devshell run [name] [command]`: Executes a command within the specified development shell.

### User Management

-   `pilo users list`: Lists all configured users.
-   `pilo users add [username] [name] [email]`: Adds a new user.
-   `pilo users remove [username]`: Removes a user.
-   `pilo users update [old_username] [new_username] [name] [email]`: Updates a user's information.

### Alias Management

-   `pilo aliases add [name] [command]`: Adds a new alias.
-   `pilo aliases remove [name]`: Removes an alias.
-   `pilo aliases update [old_name] [new_name] [command]`: Updates an alias.
-   `pilo aliases duplicate [name] [command]`: Duplicates an alias.

### GUI

-   `pilo gui`: Launches the Fyne GUI for Pilo.

### Other Commands

-   `pilo completion [bash|zsh|fish|powershell]`: Generate completion script for your shell.

### Configuration (`base-config.json`)

Pilo is configured through the `base-config.json` file, located in `~/.config/pilo/flake/`. This file allows you to customize your Nix environment. Below is a detailed breakdown of the available options.

-   **`commit_triggers`** (array of strings): A list of keywords that, when present in a commit message, will trigger a `pilo rebuild`.
-   **`packages`** (array of objects): A list of packages to be installed.
-   **`aliases`** (object): A map of custom shell aliases.
-   **`push_on_commit`** (boolean): If `true`, `pilo` will automatically push your configuration to the remote Git repository after each commit.
-   **`remote_url`** (string): The URL of the remote Git repository where your Pilo configuration is stored.
-   **`remote_branch`** (string): The default branch to use for the remote repository.
-   **`system`** (object): Contains system-specific settings:
    -   `username` (string): The primary username for the system.
    -   `desktop` (string): The desktop environment to use (e.g., `"gnome"`, `"plasma"`).
    -   `type` (string): The type of Nix installation (`"nixos"`, `"home-manager"`).
    -   `ollama` (object): Configuration for Ollama models.
-   **`users`** (array of objects): A list of users to be configured by Home Manager.

## File Structure

-   `flake/`: Contains the Nix flake and its related configurations. See the [flake/README.md](flake/README.md) for more details.
-   `cmd/`: Contains the main application code.
-   `internal/`: Contains internal packages and libraries used by the application.
-   `main.go`: The entry point for the application.
-   `go.mod` and `go.sum`: Go module files for managing dependencies.
-   `Containerfile`: A file for building the application container.
-   `container_build.sh`: A script for building and managing the application container.

## Development

First, clone the repository:

```bash
git clone https://github.com/stewlab/pilo.git
cd pilo
```

### Via Nix devShell

To build the `pilo` application using the included [Go](https://go.dev/) devShell, run the following command:

```bash
nix develop ./flake#go --command go build -o bin/pilo .
```

This command enters a Nix development shell that has Go installed, and then it builds the Pilo binary, placing it in the `bin` directory.

### Via Container

This is managed via the `container_build.sh` script, which automates building and running the application in a container.

> **Note:** Before running the script for the first time, you may need to make it executable:
> ```bash
> chmod +x container_build.sh
> ```

#### Building the Application Image

To build the final application image, run the following command from the project root:

```bash
./container_build.sh build
```

This command builds the image and tags it as `pilo-app`.

#### Running the Application

Once the image is built, you can run the application with:

```bash
./container_build.sh run pilo
```

Any arguments passed after `run` will be forwarded to the Pilo application inside the container. For example, to see the help message:

```bash
./container_build.sh run pilo --help
```

#### Development Workflow

The script also provides commands to streamline development:

-   `./container_build.sh start-dev`: Starts a persistent development container in the background.
-   `./container_build.sh shell-dev`: Opens an interactive shell inside the running development container.
-   `./container_build.sh run-dev`: Compiles and runs your application inside the development container.
-   `./container_build.sh stop-dev`: Stops and removes the development container.

#### Reusing the Containerfile for Other Go Applications

The `Containerfile` in this project is designed to be generic and can be used to build other Go applications with minimal changes.

To build a different Go application, you can pass the following arguments to the `podman build` or `docker build` command:

-   `APP_NAME`: The desired name for the final binary.
-   `MAIN_GO_PATH`: The path to the package containing the `main.go` file.
-   `LDFLAGS_STRING`: A string of linker flags, typically used for injecting version information.

Example command:

```bash
podman build \
  --build-arg APP_NAME="my-other-app" \
  --build-arg MAIN_GO_PATH="./cmd/my-other-app" \
  --build-arg LDFLAGS_STRING="-X main.Version=1.0.0" \
  -t my-other-app-image \
  -f Containerfile .