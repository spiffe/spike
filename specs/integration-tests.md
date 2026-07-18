# Spec: Integration Test Suite

## Status

Proposed (2026-07-17). Backs the TASKS.md Phase 3 item "Add integration
tests" and draws on `ideas/research-cli-testing.md`.

## Problem

The behaviors that broke this month — the Nexus boot-order deadlock,
the policy get-by-name regression, the restore flow — were all caught
by hand or by the live drill, not by tests. The suite exercises units
well but nothing verifies the seams: root-key lifecycle against a real
backing store, CRUD through the real persistence stack, and the
Pilot's behavior when Nexus is uninitialized or unreachable.

## Approach: three slices, in increasing coupling order

### Slice A: state-layer integration (no SPIRE, runs in `make test`)

A new test-only package exercising the real state + sqlite persistence
stack in-process, using the per-run temp-dir isolation that already
exists (`SPIKE_NEXUS_DATA_DIR` via `TestMain`):

- Root key: `Initialize` caches it; a second initialization does not
  regenerate or re-encrypt (the "not re-initialized twice" invariant);
  `RestoreBackingStoreFromPilotShards` recovers the same key from
  threshold shards and the data written before the "crash" reads back
  (an in-process mirror of the live drill).
- Secret CRUD: put/get/delete/undelete/list through state + sqlite,
  including versioning metadata.
- Policy CRUD: create/get/delete/list by name through state + sqlite,
  including the pattern-regex compilation invariants.

These run as part of the normal suite; the concurrent gate keeps them
cheap.

### Slice B: live integration (`//go:build integration`, opt-in)

Build-tagged tests gated on `SPIKE_INTEGRATION_TEST=1`, assuming a
healthy `make start` environment, in the spirit of the recovery drill:

- Pilot denies operations when Nexus is uninitialized.
- Pilot warns (does not hang or panic) when Nexus is unreachable.
- A CLI-level smoke pass: secret put/get/delete, policy create/get by
  name, cipher round trip; asserts on stdout now that data goes there.

Run manually or in a dedicated CI job:
`SPIKE_INTEGRATION_TEST=1 go test -tags=integration ./...`.

### Slice C: HTTP-mock helpers for CLI coverage (separate task)

Mocking the SDK's mTLS transport feeds the "CLI coverage to 60%+"
task, not this one; it is out of scope here and tracked separately.

## Open questions (decide before Slice B)

1. Should Slice B live in this repository now, or wait until there is
   a CI runner with a SPIRE environment? A tagged suite nobody runs
   rots quietly, which is how the original "CI integration test is
   broken" task was born.
2. Does Slice B subsume the recovery drill, or stay complementary?
   Recommendation: complementary. The drill kills real processes; the
   tagged tests only observe a healthy environment.

## Acceptance Criteria

- [ ] Slice A package passes in the normal `make test` run, isolated
      from `~/.spike`, leaving no artifacts.
- [ ] The "not re-initialized twice" invariant has an explicit test.
- [ ] The shard-restore round trip has an in-process test mirroring
      the live drill's semantics.
- [ ] Secret and policy CRUD paths are covered end to end at the state
      layer, including deletion and undeletion.
- [ ] Slice B lands only after the open questions are answered.
