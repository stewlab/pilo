#!/bin/sh
set -e

# Build the debug binary
echo "Building debug binary..."
CGO_CFLAGS="-O -g" go build -buildmode=exe -gcflags="all=-N -l" -o ./bin/pilo-debug .

# Run the debug binary in the background
echo "Starting the application..."
./bin/pilo-debug gui &
PID=$!

# Ensure the process is killed on exit
trap "echo 'Stopping application...'; kill $PID" EXIT

# Attach Delve
echo "Attaching Delve to PID: $PID"
dlv attach $PID --allow-non-terminal-interactive=true

# The trap will execute on exit, cleaning up the process.