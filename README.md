# Pilo

A command-line tool for managing your Nix environment.

## Installation

There are two primary ways to use Pilo: through a persistent installation or in a temporary shell.

### Persistent Installation (Recommended for Regular Use)

For long-term use, you can install Pilo directly into your user profile. This makes the `pilo` command permanently available in all your shell sessions, just like any other application installed on your system.

```bash
nix profile install github:stewlab/pilo
```

This command modifies your user's Nix profile (`~/.nix-profile`) and the installation is tracked by Nix's generation management, allowing you to roll back if needed.

### Temporary Shell (for Testing or Occasional Use)

If you want to try out Pilo without permanently installing it, or if you only need it for a single task, you can use `nix shell`. This command creates an ephemeral environment where the `pilo` command is available only for the current shell session.

```bash
nix shell github:stewlab/pilo
```

This approach does not modify your system's permanent configuration, making it a safe and non-intrusive way to use the tool. Once you exit the shell, Pilo will no longer be in your `PATH`.

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

Then, to build the `pilo` application, run the following command:

```bash
nix develop ./flake#go --command go build -o bin/pilo .
```

This command enters a Nix development shell that has Go installed, and then it builds the Pilo binary, placing it in the `bin` directory.

### Containerized Development

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

### Continuous Integration with GitHub Actions

This project includes a GitHub Actions workflow to automate the building of the application binary. The workflow is defined in `.github/workflows/build.yml` and performs the following steps:

1.  **Triggers**: The workflow is triggered automatically on every push to the `main` branch.
2.  **Build Environment**: It sets up a clean Ubuntu environment with the correct Go version.
3.  **Build**: It checks out the code, downloads dependencies, and compiles the `pilo` binary.
4.  **Artifacts**: The compiled binary is uploaded as a build artifact named `pilo-binary`.

You can find the artifacts on the "Actions" tab of your GitHub repository, under the specific workflow run. This ensures that you always have a fresh build of your application available after merging changes into your main branch.