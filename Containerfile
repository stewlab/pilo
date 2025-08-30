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
    alsa-lib-devel

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

# Build the application
RUN go build -o ${BINARY_PATH} ${MAIN_GO_PATH}

# ---- Runner Stage ----
# This stage creates the final, smaller image.
FROM quay.io/fedora/fedora:42

# Install only runtime dependencies
RUN dnf install -y \
    mesa-libGL \
    vulkan-loader \
    libglvnd \
    libxkbcommon \
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

# Set up the workdir
WORKDIR /app

# Build arguments for customization
ARG APP_NAME=app
ENV APP_NAME_ENV=${APP_NAME}

ARG BINARY_PATH=bin/${APP_NAME}
ENV BINARY_PATH_ENV=${BINARY_PATH}

# Copy the compiled binary from the builder stage
COPY --from=builder /app/${BINARY_PATH} /usr/local/bin/${APP_NAME_ENV}

RUN chmod +x /usr/local/bin/${APP_NAME_ENV}

# Set the entrypoint to run the application
CMD ["/bin/sh", "-c", "exec /usr/local/bin/${APP_NAME_ENV}"]
