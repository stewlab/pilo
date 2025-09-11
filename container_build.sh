#!/usr/bin/env bash

set -euo pipefail

# --- Configuration ---
APP_IMAGE_NAME="pilo-app"
DEV_IMAGE_NAME="go-dev-env"
DEV_CONTAINER_NAME="pilo-dev-container"
APP_NAME="pilo"
MAIN_GO_PATH="./"
# The output path inside the builder stage is now /app/build/
BUILDER_BINARY_PATH="/app/build/pilo"
FINAL_BINARY_NAME="pilo"

# --- Script Logic ---

# --- Auto-detection of Container Engine ---
if command -v podman &> /dev/null; then
  CONTAINER_CMD="podman"
elif command -v docker &> /dev/null; then
  CONTAINER_CMD="docker"
else
  echo "Error: Neither Podman nor Docker found. Please install one of them." >&2
  exit 1
fi
echo "Using container engine: $CONTAINER_CMD"

# --- Functions ---

usage() {
  echo "Usage: $0 [start-dev|stop-dev|shell-dev|run-dev|build|run|rebuild-dev|rebuild-app|build-artifact] [app_args...]"
  echo "Development Commands:"
  echo "  start-dev    : Start the persistent development container."
  echo "  stop-dev     : Stop and remove the persistent development container."
  echo "  shell-dev    : Start an interactive shell in the running development container."
  echo "  run-dev      : Compile and run the application inside the dev container (starts it if not running)."
  echo "  rebuild-dev  : Force rebuild of the development image."
  echo ""
  echo "Application Commands:"
  echo "  build        : Build the final application image."
  echo "  run          : Run the final application image."
  echo "  rebuild-app  : Force rebuild of the application image."
  echo "  build-artifact: Build the application binary and place it in the artifacts directory."
  exit 1
}

image_exists() {
  $CONTAINER_CMD image inspect "$1" &> /dev/null
}

container_is_running() {
  $CONTAINER_CMD ps -q -f name="^${DEV_CONTAINER_NAME}$" | grep -q .
}

build_dev_image() {
  echo "Building development image: ${DEV_IMAGE_NAME}"
  $CONTAINER_CMD build --target builder -t "${DEV_IMAGE_NAME}" -f Containerfile .
}

build_app_image() {
  local extra_args="$1"
  local build_version="${2:-0.0.1}" # Default version if not provided
  local image_tag="${3:-${APP_IMAGE_NAME}}" # Default to APP_IMAGE_NAME if not provided
  local ldflags="-X pilo/internal/cli.Version=${build_version}"
  echo "Building final application image: ${image_tag} with version ${build_version}"
  $CONTAINER_CMD build \
    ${extra_args} \
    --build-arg APP_NAME="${FINAL_BINARY_NAME}" \
    --build-arg MAIN_GO_PATH="${MAIN_GO_PATH}" \
    --build-arg LDFLAGS_STRING="${ldflags}" \
    -t "${image_tag}" \
    -f Containerfile .
}

# Sets up the necessary options for running a GUI application inside the container.
# This function handles Wayland and X11 display servers.
setup_gui_options() {
    local -n opts_ref=$1 # Nameref to the array that will store the options
    
    # Pass DISPLAY and WAYLAND_DISPLAY from host to container
    opts_ref+=(--env DISPLAY="${DISPLAY}")
    opts_ref+=(--env WAYLAND_DISPLAY="${WAYLAND_DISPLAY}")

    # Handle X11 authorization
    if [[ -n "${DISPLAY:-}" && -n "${XAUTHORITY:-}" ]]; then
        local XAUTH_FILE=$(mktemp)
        xauth nlist "${DISPLAY}" | sed -e 's/^..../ffff/' | xauth -f "${XAUTH_FILE}" nmerge -
        chmod 644 "${XAUTH_FILE}" # Ensure the file is world-readable
        opts_ref+=(--env XAUTHORITY="/tmp/.docker.xauthority")
        opts_ref+=(-v "${XAUTH_FILE}:/tmp/.docker.xauthority:z")
        # Add a cleanup trap to remove the temporary xauthority file
        trap "rm -f ${XAUTH_FILE}" EXIT
    fi

    # Mount the X11 socket for XWayland/X11 applications
    opts_ref+=(-v /tmp/.X11-unix:/tmp/.X11-unix)

    # Mount the user's runtime directory for Wayland, Pipewire, and D-Bus
    if [[ -d "${XDG_RUNTIME_DIR}" ]]; then
        opts_ref+=(-v "${XDG_RUNTIME_DIR}:${XDG_RUNTIME_DIR}")
        opts_ref+=(--env XDG_RUNTIME_DIR="${XDG_RUNTIME_DIR}")
    else
        echo "Warning: XDG_RUNTIME_DIR is not set. GUI and audio may not work." >&2
    fi

    # Add device access for graphics and sound
    opts_ref+=(--device /dev/dri)
    opts_ref+=(--device /dev/snd)
}

start_dev_container() {
  if ! image_exists "${DEV_IMAGE_NAME}"; then
    build_dev_image
  fi

  if container_is_running; then
    echo "Development container is already running."
    return
  fi

  echo "Starting persistent development container..."
  mkdir -p "${PWD}/.go/pkg/mod"
  local opts=( -d --name "${DEV_CONTAINER_NAME}" )
  setup_gui_options opts
  opts+=(
    --env GOMODCACHE=/go/pkg/mod
    -v "${PWD}:/app:z"
    -v "${PWD}/.go:/go:z"
    -w /app
  )
  
  $CONTAINER_CMD run "${opts[@]}" "${DEV_IMAGE_NAME}" sleep infinity > /dev/null
  echo "Container started."
}

stop_dev_container() {
  if ! container_is_running; then
    echo "Development container is not running."
    return
  fi
  echo "Stopping and removing development container..."
  $CONTAINER_CMD rm -f "${DEV_CONTAINER_NAME}" > /dev/null
  echo "Container stopped."
}

shell_dev_container() {
  if ! container_is_running; then
    echo "Development container is not running. Starting it first..."
    start_dev_container
  fi
  echo "Attaching to development container shell..."
  $CONTAINER_CMD exec -it "${DEV_CONTAINER_NAME}" bash
}

run_dev_build_and_run() {
  if ! container_is_running; then
    echo "Development container is not running. Starting it first..."
    start_dev_container
  fi
  
  echo "Compiling and running in development container..."
  # Construct a command that exports the necessary GUI variables before building and running
  local build_version
  build_version=$(git describe --tags --always --dirty 2>/dev/null || echo "0.0.1")
  local ldflags="-X pilo/internal/cli.Version=${build_version}"
  local run_cmd="export DISPLAY=${DISPLAY}; export XAUTHORITY=${XAUTHORITY}; go build -ldflags='${ldflags}' -o /app/build/${APP_NAME} ${MAIN_GO_PATH} && /app/build/${APP_NAME} $@"
  $CONTAINER_CMD exec -it "${DEV_CONTAINER_NAME}" bash -c "${run_cmd}"
}

run_app() {
  echo "Running application..."
  local opts=( --rm -it --network=host ) # Added --network=host for X11 forwarding
  setup_gui_options opts
  $CONTAINER_CMD run "${opts[@]}" "${APP_IMAGE_NAME}" "$@"
}

build_artifact_binary() {
    local tool_image_name="${APP_IMAGE_NAME}-builder-tool"
    local build_version
    build_version=$(git describe --tags --always --dirty 2>/dev/null || echo "0.0.1")
    local ldflags="-X pilo/internal/cli.Version=${build_version}"

    echo "Building builder tool image..."
    $CONTAINER_CMD build \
        --target builder \
        --build-arg TOOL=true \
        --build-arg APP_NAME="${FINAL_BINARY_NAME}" \
        --build-arg MAIN_GO_PATH="${MAIN_GO_PATH}" \
        --build-arg LDFLAGS_STRING="${ldflags}" \
        -t "${tool_image_name}" \
        -f Containerfile .

    echo "Creating and starting temporary container..."
    local container_id
    container_id=$($CONTAINER_CMD run -d "${tool_image_name}" sleep infinity)

    echo "Building binary in container..."
    # The go build command is now run inside the running container
    $CONTAINER_CMD exec "${container_id}" go build -ldflags="${ldflags}" -tags osusergo,netgo -extldflags "-static" -o "/app/build/${FINAL_BINARY_NAME}" "${MAIN_GO_PATH}"


    echo "Copying binary from container..."
    mkdir -p artifacts
    $CONTAINER_CMD cp "${container_id}:/app/build/${FINAL_BINARY_NAME}" "./artifacts/${FINAL_BINARY_NAME}-linux-amd64"

    echo "Stopping and removing temporary container..."
    $CONTAINER_CMD rm -f "${container_id}"

    echo "Binary created at ./artifacts/${FINAL_BINARY_NAME}-linux-amd64"
}

# --- Main Logic ---
COMMAND="shell-dev"
if [[ $# -gt 0 ]]; then
  COMMAND="$1"
  shift
fi

case "$COMMAND" in
  start-dev)
    start_dev_container
    ;;
  stop-dev)
    stop_dev_container
    ;;
  shell-dev)
    shell_dev_container
    ;;
  run-dev)
    run_dev_build_and_run "$@"
    ;;
  rebuild-dev)
    stop_dev_container
    build_dev_image
    start_dev_container
    ;;
  build)
    build_version="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "0.0.1")}"
    build_app_image "" "${build_version}" "${APP_IMAGE_NAME}:${build_version}"
    ;;
  rebuild-app)
    build_app_image "--no-cache"
    ;;
  run)
    if ! image_exists "${APP_IMAGE_NAME}"; then
      echo "Application image not found. Building it first."
      build_app_image "" "" "pilo-app:latest"
    fi
    run_app "$@"
    ;;
  build-artifact)
    build_artifact_binary
    ;;
  *)
    usage
    ;;
esac