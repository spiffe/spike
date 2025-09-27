#!/usr/bin/env bash

set -euo pipefail

if [[ -z "${SPIKE_NEXUS_URL:-}" ]]; then
  echo "SPIKE_NEXUS_URL must be set to run this test" >&2
  exit 1
fi

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

PLAIN="$TMPDIR/plain.txt"
ENC="$TMPDIR/plain.enc"
DEC="$TMPDIR/plain.dec.txt"

echo "hello from $(date)" > "$PLAIN"

echo "Encrypting (octet-stream)..."
spike cipher encrypt -f "$PLAIN" -o "$ENC"

echo "Decrypting (octet-stream)..."
spike cipher decrypt -f "$ENC" -o "$DEC"

echo "Verifying..."
diff -u "$PLAIN" "$DEC"

echo "Done. OK"



