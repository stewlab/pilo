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

### Package Management

-   `pilo install-pkg [pkg...]`: Installs packages to your user profile (non-NixOS) or provides a temporary shell (NixOS).
-   `pilo remove [pkg...]`: Removes packages from your user profile (non-NixOS).
-   `pilo upgrade`: Upgrades all packages in your user profile or NixOS system.
-   `pilo search [query]`: Searches for packages in `nixpkgs`.
-   `pilo add-app [pname] [version] [url] [sha256]`: Adds a new Flake App to your flake/packages directory.

### Development Shells

Pilo offers two types of development shells: temporary, on-the-fly shells and persistent, project-specific shells defined in your flake.

-   **`pilo shell [pkg...]`**: This command creates a temporary, ephemeral shell with the specified packages. It is useful for quick tasks or trying out new tools without modifying your system configuration. The shell and its packages are gone once you exit.

-   **`pilo develop [shell]`**: This is the primary command for entering a persistent development shell defined in your flake. It serves as a convenient, high-level wrapper around `nix develop`.

#### Managing Persistent Devshells

The `pilo devshell` command provides a suite of tools for managing your development shells. While `pilo develop` is for *using* shells, `pilo devshell` is for *managing* them.

-   **`pilo devshell add [name]`**: Adds a new, empty development shell definition to your `flake/devshells/` directory.
-   **`pilo devshell remove [name]`**: Removes an existing development shell definition.
-   **`pilo devshell enter [name]`**: Enters the specified development shell. While its outcome is similar to `pilo develop`, it exists within the `devshell` subcommand for a consistent management workflow (e.g., add a shell, then immediately enter it).
-   **`pilo devshell run [name] [command]`**: Executes a single command within the context of the specified development shell without entering it interactively.

**Important Note on Configuration Changes**:

Any changes made to your `base-config.json` or your flake files (including adding or removing development shells) will **not** take effect until you run `pilo rebuild`. This command applies your changes to the system's Nix configuration, making them available for use.

### GUI

-   `pilo gui`: Launches the Fyne GUI for Pilo.

### Other Commands

-   `pilo completion [bash|zsh|fish|powershell]`: Generate completion script for your shell.

### Configuration (`base-config.json`)

Pilo is configured through the `base-config.json` file, located in `~/.config/pilo/flake/`. This file allows you to customize your Nix environment. Below is a detailed breakdown of the available options.

-   **`commit_triggers`** (array of strings): A list of keywords that, when present in a commit message, will trigger a `pilo rebuild`. This is useful for automating system updates when you commit changes to your configuration.

-   **`packages`** (array of objects): A list of packages to be installed. Each object has two keys:
    -   `name` (string): The name of the package from `nixpkgs`.
    -   `installed` (boolean): Whether the package should be installed.

-   **`aliases`** (object): A map of custom shell aliases. The key is the alias name and the value is the command it should execute.

-   **`push_on_commit`** (boolean): If `true`, `pilo` will automatically push your configuration to the remote Git repository after each commit.

-   **`remote_url`** (string): The URL of the remote Git repository where your Pilo configuration is stored.

-   **`remote_branch`** (string): The default branch to use for the remote repository.

-   **`system`** (object): Contains system-specific settings:
    -   `username` (string): The primary username for the system.
    -   `desktop` (string): The desktop environment to use (e.g., `"gnome"`, `"plasma"`).
    -   `type` (string): The type of Nix installation (`"nixos"`, `"home-manager"`).
    -   `ollama` (object): Configuration for Ollama models.
        -   `models` (string): A comma-separated list of Ollama models to install.

-   **`users`** (array of objects): A list of users to be configured by Home Manager. Each object contains:
    -   `username` (string): The username of the user.
    -   `email` (string): The user's email address for Git configuration.
    -   `name` (string): The user's full name for Git configuration.

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

-   `./container_build.sh start-dev`: Starts a persistent development container in the background. Your project directory is mounted into the container, so changes are reflected live.
-   `./container_build.sh shell-dev`: Opens an interactive shell inside the running development container.
-   `./container_build.sh run-dev`: Compiles and runs your application inside the development container. This is ideal for quick testing without rebuilding the image.
-   `./container_build.sh stop-dev`: Stops and removes the development container.

#### Reusing the Containerfile for Other Go Applications

The `Containerfile` in this project is designed to be generic and can be used to build other Go applications with minimal changes. The `builder` stage is parameterized using build arguments (`ARG`).

To build a different Go application, you can pass the following arguments to the `podman build` or `docker build` command:

-   `APP_NAME`: The desired name for the final binary.
-   `MAIN_GO_PATH`: The path to the package containing the `main.go` file (e.g., `./cmd/my-other-app`).
-   `LDFLAGS_STRING`: A string of linker flags, typically used for injecting version information (e.g., `-X main.Version=1.0.0`).

Example command:

```bash
podman build \
  --build-arg APP_NAME="my-other-app" \
  --build-arg MAIN_GO_PATH="./cmd/my-other-app" \
  --build-arg LDFLAGS_STRING="-X main.Version=1.0.0" \
  -t my-other-app-image \
  -f Containerfile .
```