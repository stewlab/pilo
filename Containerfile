# syntax=docker/dockerfile:1

# ---- Builder Stage ----
# This stage builds the Go application.
FROM quay.io/fedora/fedora:42 AS builder

# Install build dependencies
RUN dnf install -y \
    git \
    golang \
    delve \
    pkg-config \
    mesa-libGL-devel \
    xauth \
    vulkan-loader-devel \
    libglvnd-devel \
    libxkbcommon-devel \
    wayland-devel \
    wayland-protocols-devel \
    libX11-devel \
    libXcursor-devel \
    libXrandr-devel \
    libXinerama-devel \
    libXi-devel \
    libXxf86vm-devel \
    libXext-devel \
    libXfixes-devel \
    libXdamage-devel \
    libXcomposite-devel \
    xorg-x11-proto-devel \
    at-spi2-core-devel \
    libxcb-devel \
    portaudio-devel \
    alsa-lib-devel && \
    dnf clean all

# Set Go environment variables
ENV GOPROXY=direct
ENV GOSUMDB=off

# Set up the workdir
WORKDIR /app

# Copy go module files and download dependencies first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build arguments for customization
ARG APP_NAME=app
ARG BINARY_PATH=bin/${APP_NAME}
ARG MAIN_GO_PATH=./

ARG BUILD_VERSION=0.0.1
 
# Build the application
RUN go build -ldflags="-X pilo/internal/cli.Version=${BUILD_VERSION}" -o ${BINARY_PATH} ${MAIN_GO_PATH}

# ---- Runner Stage ----
# This stage creates the final, smaller image.
FROM quay.io/fedora/fedora:42

# Install only runtime dependencies
RUN dnf install -y \
    mesa-libGL \
    vulkan-loader \
    libglvnd \
    libxkbcommon \
    xauth \
    libX11 \
    libXcursor \
    libXrandr \
    libXinerama \
    libXi \
    libXxf86vm \
    libXext \
    libXfixes \
    libXdamage \
    libXcomposite \
    at-spi2-core \
    libxcb \
    portaudio \
    alsa-lib \
    fzf \
    neovim && \
    dnf clean all

# Set up the workdir and user
RUN adduser -u 1001 -d /app -s /bin/sh appuser && \
    chown -R appuser:appuser /app

# Create and permission the /nix directory for the single-user installation
RUN mkdir -m 0755 /nix && chown appuser /nix

# Switch to the appuser to install Nix in the user's profile
USER appuser
WORKDIR /app

# Install Nix as the appuser
RUN sh <(curl --proto '=https' --tlsv1.2 -L https://nixos.org/nix/install) --no-daemon

# Switch back to root to copy the binary and set permissions
USER root

# Build arguments for customization
ARG APP_NAME=app
ENV APP_NAME_ENV=${APP_NAME}

ARG BINARY_PATH=bin/${APP_NAME}
ENV BINARY_PATH_ENV=${BINARY_PATH}

# Copy the compiled binary from the builder stage
COPY --from=builder /app/${BINARY_PATH} /usr/local/bin/${APP_NAME_ENV}
RUN chmod +x /usr/local/bin/${APP_NAME_ENV}

# Switch back to the appuser for the final runtime environment
USER appuser
WORKDIR /app

# Set the entrypoint to run the application, ensuring the Nix profile is sourced
# This allows the app to start even if the Nix setup isn't fully complete.
CMD ["/bin/sh", "-c", "[ -f /app/.nix-profile/etc/profile.d/nix.sh ] && . /app/.nix-profile/etc/profile.d/nix.sh; exec /usr/local/bin/${APP_NAME_ENV}"]
