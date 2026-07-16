#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Live recovery/restore drill for the bare-metal dev environment.
#
# Run this from the repository root, in a second terminal, after
# `make start` has completed cleanly (SPIRE server and agent, three
# Keepers, and Nexus with a persistent backend are all running). Run it
# from the same shell environment as `make start` so a restarted Nexus
# inherits the same configuration.
#
# The drill proves the break-the-glass runbook end to end
# (see https://spike.ist/operations/recovery/):
#
#   1. Writes a marker secret while the system is healthy.
#   2. Exports recovery shards via `spike operator recover`.
#   3. Simulates a crash: kills Nexus and all Keepers (the Keepers lose
#      their in-memory shards, so auto-recovery is impossible).
#   4. Restarts Nexus alone and feeds the shards back one by one via
#      `spike operator restore`, using its non-interactive stdin mode.
#   5. Verifies the marker secret reads back after the restore.
#
# The Pilot's SPIRE entry is rotated through the recover and restore
# roles along the way; a trap reverts it to the superuser role on every
# exit path. The Keepers stay down when the drill ends: run `make kill`
# and `make start` afterward to return to a pristine environment.

set -u

DRILL_MARKER_PATH="drill/recovery-marker"
DRILL_MARKER_VALUE="drill-$(date +%s)-$$"
RECOVERY_DIR="${SPIKE_PILOT_RECOVERY_DIR:-$HOME/.spike/recover}"
NEXUS_LOG="$(mktemp -t spike-drill-nexus.XXXXXX.log)"

say() {
  echo ""
  echo "drill: $*"
}

fail() {
  echo "" >&2
  echo "drill: FAIL: $*" >&2
  exit 1
}

# Returns the Pilot's current role by asking the SPIRE server which of
# the known role SPIFFE IDs has a registration entry.
current_pilot_role() {
  for role in superuser recover restore; do
    if spire-server entry show \
      -spiffeID "spiffe://spike.ist/spike/pilot/role/$role" 2>/dev/null |
      grep -q "Entry ID"; then
      echo "$role"
      return 0
    fi
  done
  echo "unknown"
}

# Reverts the Pilot's entry to the superuser role no matter which role
# the drill died in. Runs on every exit path via the trap below.
ensure_superuser_role() {
  case "$(current_pilot_role)" in
  recover)
    say "reverting the Pilot entry from recover to superuser..."
    ./hack/bare-metal/entry/spire-server-entry-recover-revert.sh
    ;;
  restore)
    say "reverting the Pilot entry from restore to superuser..."
    ./hack/bare-metal/entry/spire-server-entry-restore-revert.sh
    ;;
  superuser)
    :
    ;;
  *)
    echo "drill: WARNING: could not determine the Pilot role;" >&2
    echo "drill: restore it manually with" >&2
    echo "drill:   ./hack/bare-metal/entry/spire-server-entry-su-register.sh" >&2
    ;;
  esac
}

trap ensure_superuser_role EXIT

# Retries a command until it succeeds or the attempts run out.
retry() {
  local attempts="$1"
  local delay="$2"
  shift 2

  local i
  for ((i = 1; i <= attempts; i++)); do
    if "$@"; then
      return 0
    fi
    sleep "$delay"
  done
  return 1
}

# --- Preflight ------------------------------------------------------------

[ -x ./bin/spike ] || fail "run this from the repository root" \
  "(./bin/spike not found; did make start build the binaries?)"

for b in spike spire-server; do
  command -v "$b" >/dev/null 2>&1 || fail "'$b' is not on PATH"
done

command -v spike | xargs test "$(pwd)/bin/spike" -ef ||
  fail "PATH resolves 'spike' outside $(pwd)/bin"

pgrep -x nexus >/dev/null || fail "Nexus is not running (run make start)"
pgrep -x keeper >/dev/null || fail "no Keepers running (run make start)"
pgrep -x spire-server >/dev/null || fail "SPIRE server is not running"

if [ "${SPIKE_NEXUS_BACKEND_STORE:-}" = "memory" ]; then
  fail "the drill needs a persistent backend;" \
    "unset SPIKE_NEXUS_BACKEND_STORE"
fi

if [ "$(current_pilot_role)" != "superuser" ]; then
  fail "the Pilot entry is not in the superuser role;" \
    "revert it before running the drill"
fi

# --- 1. Write a marker secret while healthy -------------------------------

say "writing the marker secret ($DRILL_MARKER_PATH)..."
spike secret put "$DRILL_MARKER_PATH" value="$DRILL_MARKER_VALUE" ||
  fail "could not write the marker secret"

# Note: the Pilot prints its output on stderr (cobra's Print* default),
# so merge the streams before grepping, like the other harness scripts.
spike secret get "$DRILL_MARKER_PATH" 2>&1 |
  grep -q "$DRILL_MARKER_VALUE" ||
  fail "could not read the marker secret back while healthy"

# --- 2. Export recovery shards ---------------------------------------------

say "rotating the Pilot entry to the recover role..."
./hack/bare-metal/entry/spire-server-entry-recover-register.sh ||
  fail "could not register the recover role"

recover_once() {
  spike operator recover 2>&1 | grep -q "recovery directory"
}

say "exporting recovery shards (retrying while the SVID rotates)..."
retry 5 3 recover_once || fail "spike operator recover did not succeed"

shard_count=$(ls "$RECOVERY_DIR"/spike.recovery.*.txt 2>/dev/null | wc -l)
[ "$shard_count" -ge 2 ] ||
  fail "expected at least 2 shard files in $RECOVERY_DIR," \
    "found $shard_count"
say "exported $shard_count shards to $RECOVERY_DIR"

say "rotating the Pilot entry back to superuser..."
./hack/bare-metal/entry/spire-server-entry-recover-revert.sh ||
  fail "could not revert the recover role"

# --- 3. Simulate the crash --------------------------------------------------

say "simulating a crash: killing Nexus and all Keepers..."
pkill -x nexus
pkill -x keeper

crash_complete() {
  ! pgrep -x nexus >/dev/null && ! pgrep -x keeper >/dev/null
}
retry 10 1 crash_complete || fail "Nexus/Keeper processes did not exit"
say "Nexus and Keepers are down; Keeper shards are lost"

# --- 4. Restart Nexus alone and restore ------------------------------------

say "restarting Nexus alone (log: $NEXUS_LOG)..."
nohup ./hack/bare-metal/startup/start-nexus.sh \
  >"$NEXUS_LOG" 2>&1 &

say "rotating the Pilot entry to the restore role..."
./hack/bare-metal/entry/spire-server-entry-restore-register.sh ||
  fail "could not register the restore role"

restored=0
for shard_file in "$RECOVERY_DIR"/spike.recovery.*.txt; do
  say "feeding $(basename "$shard_file")..."

  feed_shard() {
    output=$(spike operator restore <"$shard_file" 2>&1)
    # Two conditions are transient and worth retrying: Nexus is still
    # starting (comms errors), and the SPIRE agent has not yet rotated
    # the Pilot SVID to the restore role (access_unauthorized).
    if echo "$output" | grep -q "Failed to communicate"; then
      return 1
    fi
    if echo "$output" | grep -q "access_unauthorized"; then
      return 1
    fi
    return 0
  }

  retry 20 2 feed_shard ||
    fail "could not feed $(basename "$shard_file"):" \
      "Nexus unreachable or the restore SVID never arrived"

  echo "$output"
  if echo "$output" | grep -q "restored and ready"; then
    restored=1
    break
  fi
done

[ "$restored" -eq 1 ] ||
  fail "fed all shards but Nexus did not report itself restored"

say "rotating the Pilot entry back to superuser..."
./hack/bare-metal/entry/spire-server-entry-restore-revert.sh ||
  fail "could not revert the restore role"

# --- 5. Verify the pre-crash secret ----------------------------------------

verify_marker() {
  spike secret get "$DRILL_MARKER_PATH" 2>&1 |
    grep -q "$DRILL_MARKER_VALUE"
}

say "verifying the pre-crash marker secret..."
retry 10 2 verify_marker ||
  fail "the marker secret did not read back after the restore"

# --- Cleanup ----------------------------------------------------------------

say "cleaning up: deleting the marker secret and the shard files..."
spike secret delete "$DRILL_MARKER_PATH" >/dev/null 2>&1
rm -f "$RECOVERY_DIR"/spike.recovery.*.txt

say "PASS: recovery/restore drill completed successfully."
echo ""
echo "  A secret written before the crash survived the loss of Nexus"
echo "  and every Keeper, and was readable after a shard-based restore."
echo ""
echo "  Note: the Keepers are still down. Run 'make kill' and then"
echo "  'make start' to return to a pristine environment."
echo ""
