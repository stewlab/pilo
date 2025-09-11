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

# group dev
# Build NixOS configuration if --nixos flag is passed
if [ "$1" = "--nixos" ]; then
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
nix develop ./flake#go -c go build -ldflags="${LDFLAGS}" -tags osusergo,netgo -extldflags "-static" -o ./bin/pilo .

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