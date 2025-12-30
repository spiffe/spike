#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
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

# Function to run a script in background and store its PID
run_background() {
    local script="$1"
    mkdir -p "./logs"
    # shellcheck disable=SC2155
    local logfile="./logs/$(basename "$script").log"

    if [ ! -x "$script" ]; then
        echo "Error: $script is not executable"
        return 1
    fi

    "$script" > "$logfile" 2>&1 &
    local pid=$!
    BG_PIDS+=("$pid")
    echo "Started $script with PID $pid, logging to $logfile"
}

# Export functions and variables so they're available to the sourcing script
export BG_PIDS
export -f cleanup
export -f run_background

# Set up trap for cleanup
trap cleanup EXIT INT TERM
