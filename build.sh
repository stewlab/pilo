####################
# Local Build

# check flake
# nix flake check ./flake

# install
# sudo nixos-rebuild switch --flake ./flake#nixos

# build using internal system flake
# nix develop ./flake#go --command go build -o bin/pilo .
# nix-develop go --command go build -o bin/pilo .

# build with self-contained flake
# nix develop ./dev --command go build -o bin/pilo .

set -euo pipefail

usage() {
  cat <<'EOF'
Usage: ./build.sh [--flake] [--nixos]

  --flake   Build using the repo root flake (nix build .#default) and copy result to ./bin/pilo
  --nixos   Build NixOS configuration from ./flake (if flake/hosts/nixos exists)

Default: build with go inside the Go devshell template (fast local build).
EOF
}

BUILD_FLAKE=0
BUILD_NIXOS=0

for arg in "$@"; do
  case "$arg" in
    --flake) BUILD_FLAKE=1 ;;
    --nixos) BUILD_NIXOS=1 ;;
    -h|--help) usage; exit 0 ;;
    *)
      echo "Unknown argument: $arg" >&2
      usage
      exit 2
      ;;
  esac
done

# group dev
# Build NixOS configuration if --nixos flag is passed
if [ "$BUILD_NIXOS" = "1" ]; then
  if [ -d "flake/hosts/nixos" ]; then
    echo "Building NixOS configuration..."
    (cd flake && nixos-rebuild build --flake .#nixos)
  else
    echo "NixOS host directory not found, skipping NixOS build."
  fi
fi

echo "Building Pilo binary..."
VERSION=$(git describe --tags --always --dirty)
LDFLAGS="-X pilo/internal/cli.Version=${VERSION}"

if [ "$BUILD_FLAKE" = "1" ]; then
  # Reproducible build via Nix flake
  rm -f ./result
  nix build .#default

  if [ ! -x ./result/bin/pilo ]; then
    echo "Expected ./result/bin/pilo after nix build, but it wasn't found." >&2
    echo "Build output directory contains:" >&2
    ls -la ./result >&2 || true
    exit 1
  fi

  mkdir -p ./bin
  cp -f ./result/bin/pilo ./bin/pilo
  chmod +x ./bin/pilo
  echo "Built via flake -> ./bin/pilo"
else
  # Fast local build inside dev shell (uses local sources; not fully reproducible)
  nix develop ./flake/devshells/templates/go -c go build -ldflags="${LDFLAGS}" -o ./bin/pilo .
fi

# group prod
# nix-develop go --command go build -o bin/pilo .
# ./bin/pilo setup
# ./bin/pilo rebuild # (sudo nixos-rebuild switch --flake ~/.config/pilo/flake#nixos)


#####################
# Container Builds
# sh container_build.sh build

# run in sandboxed container
# sh container_build.sh run gui

# Start the container (run once per dev session)
# sh container_build.sh start-dev

# Compile and run your code
# sh container_build.sh run-dev

# Access the container shell
# sh container_build.sh shell-dev

# Stop the container
# sh container_build.sh stop-dev


####################
# Nix Packaging

# go mod vendor

# Test Build
# nix build .#default

# Test Install
# nix shell .#default
# or from git (public)
# nix shell github:stewlab/pilo
# or from git (private)
# nix shell git+https://github.com/stewlab/pilo.git


####################
# Debugging
# CGO_CFLAGS="-O -g" go build -gcflags="all=-N -l" -o ./bin/pilo-debug .
# CGO_CFLAGS="-O -g" go run -gcflags="all=-N -l" . gui & echo $!

# dlv exec ./bin/pilo-debug -- gui
# dlv attach 56499