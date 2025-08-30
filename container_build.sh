#!/usr/bin/env bash

set -euo pipefail

# --- Configuration ---
APP_IMAGE_NAME="pilo-app"
DEV_IMAGE_NAME="go-dev-env"
DEV_CONTAINER_NAME="pilo-dev-container"
APP_NAME="pilo"
MAIN_GO_PATH="./"
BUILDER_BINARY_PATH="bin/pilo"
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
  echo "Usage: $0 [start-dev|stop-dev|shell-dev|run-dev|build|run|rebuild-dev|rebuild-app] [app_args...]"
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
  echo "Building final application image: ${APP_IMAGE_NAME}"
  $CONTAINER_CMD build \
    ${extra_args} \
    --build-arg APP_NAME="${FINAL_BINARY_NAME}" \
    --build-arg BINARY_PATH="${BUILDER_BINARY_PATH}" \
    --build-arg MAIN_GO_PATH="${MAIN_GO_PATH}" \
    -t "${APP_IMAGE_NAME}" \
    -f Containerfile .
}

setup_gui_options() {
    local HOST_UID
    HOST_UID=$(id -u)
    local RUNTIME_DIR="/run/user/${HOST_UID}"
    local opts=()

    opts+=(--privileged --cap-add SYS_ADMIN --security-opt seccomp=unconfined --device /dev/dri --device /dev/snd --device /dev/fuse)
    if [[ "$CONTAINER_CMD" == "podman" ]]; then
        opts+=(--security-opt label=disable)
    fi

    local WAYLAND_ENV_VAR_VAL="${WAYLAND_DISPLAY:-wayland-0}"
    local DISPLAY_ENV_VAR_VAL="${DISPLAY:-:0}"
    opts+=(--env DISPLAY="$DISPLAY_ENV_VAR_VAL" --env WAYLAND_DISPLAY="$WAYLAND_ENV_VAR_VAL" --env WINIT_UNIX_BACKEND="x11" -v /tmp/.X11-unix:/tmp/.X11-unix:ro)

    if [[ -n "$XAUTHORITY" ]]; then
        opts+=(--env XAUTHORITY="$XAUTHORITY")
    elif [[ -d "$RUNTIME_DIR" ]]; then
        local xauth_file
        xauth_file=$(find "$RUNTIME_DIR" -name ".mutter-Xwaylandauth.*" 2>/dev/null | head -n 1)
        if [[ -n "$xauth_file" ]]; then
            opts+=(--env XAUTHORITY="$xauth_file")
        fi
    fi

    if [[ -d "$RUNTIME_DIR" ]]; then
        local mount_opts="z"
        if [[ "$CONTAINER_CMD" == "docker" ]]; then
            mount_opts="ro"
        fi
        opts+=(--env XDG_RUNTIME_DIR="$RUNTIME_DIR" --env DBUS_SESSION_BUS_ADDRESS="unix:path=${RUNTIME_DIR}/bus" -v "$RUNTIME_DIR:$RUNTIME_DIR:$mount_opts")
    else
        echo "Warning: XDG_RUNTIME_DIR ($RUNTIME_DIR) not found on host. GUI and session features might fail." >&2
    fi

    local PULSE_SOCKET_HOST_PATH="${RUNTIME_DIR}/pulse/native"
    local PIPEWIRE_SOCKET_HOST_PATH="${RUNTIME_DIR}/pipewire-0"
    if [ -S "$PULSE_SOCKET_HOST_PATH" ]; then
        opts+=(--env PULSE_SERVER="unix:$PULSE_SOCKET_HOST_PATH")
    elif [ -S "$PIPEWIRE_SOCKET_HOST_PATH" ]; then
        opts+=(--env PULSE_SERVER="unix:${RUNTIME_DIR}/pulse/native")
    fi
    
    echo "${opts[@]}"
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
  opts+=( $(setup_gui_options) )
  opts+=(
    --env GOMODCACHE=/go/pkg/mod
    -v "${PWD}:/app"
    -v "${PWD}/.go:/go"
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
  local build_and_run_cmd="go build -o ${BUILDER_BINARY_PATH} ${MAIN_GO_PATH} && ./${BUILDER_BINARY_PATH} $@"
  $CONTAINER_CMD exec -it "${DEV_CONTAINER_NAME}" bash -c "${build_and_run_cmd}"
}

run_app() {
  echo "Running application..."
  local opts=( --rm -it )
  opts+=( $(setup_gui_options) )
  $CONTAINER_CMD run "${opts[@]}" "${APP_IMAGE_NAME}" /usr/local/bin/pilo "$@"
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
    build_app_image ""
    ;;
  rebuild-app)
    build_app_image "--no-cache"
    ;;
  run)
    if ! image_exists "${APP_IMAGE_NAME}"; then
      echo "Application image not found. Building it first."
      build_app_image ""
    fi
    run_app "$@"
    ;;
  *)
    usage
    ;;
esac