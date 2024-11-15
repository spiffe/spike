#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Array to store background PIDs
declare -a BG_PIDS

# Function to cleanup background processes
cleanup() {
    echo "Cleaning up background processes..."
    for pid in "${BG_PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            echo "Killing process $pid"
            kill "$pid"
            wait "$pid" 2>/dev/null
        fi
    done
    exit
}

# Set trap for script termination
trap cleanup EXIT INT TERM

# Function to run a script in background and store its PID
run_background() {
    local script="$1"
    # shellcheck disable=SC2155
    local logfile="./$(basename "$script").log"

    if [ ! -x "$script" ]; then
        echo "Error: $script is not executable"
        return 1
    fi

    "$script" > "$logfile" 2>&1 &
    local pid=$!
    BG_PIDS+=("$pid")
    echo "Started $script with PID $pid, logging to $logfile"
}

run_background "./hack/spire-server-starter.sh"
sleep 10

run_background "./hack/spire-agent-starter.sh"
run_background "./keeper"
run_background "./nexus"

echo "Waiting for background processes to start..."
sleep 10

echo "Starting tests..."
go build -o ci-test ./ci/test/main.go
./ci-test

echo "Tests completed successfully!"
echo "Cleaning up background processes..."

cleanup

echo "Everything is awesome!"
