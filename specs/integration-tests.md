# Spec: Integration Test Suite

## Status

Implemented (2026-07-18). Slice A shipped 2026-07-17
(`app/nexus/internal/state/integration`); Slice B shipped 2026-07-18
(`app/spike/internal/cmd/integration`, build tag `integration`). Backs the
TASKS.md Phase 3 item "Add integration tests" and draws on
`ideas/research-cli-testing.md`.

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
healthy `make start` environment, in the spirit of the recovery drill.
They drive the built `spike` binary end to end (resolved from PATH, per
the harness convention) rather than importing command internals. Note
that both error paths funnel through `stdout.HandleAPIError` and exit 0
(the subcommands use cobra `Run`, not `RunE`), so the assertions key on
the stderr message and the no-hang property, not the exit code.

- A CLI-level smoke pass (`TestPilotSmokePass`, non-destructive): secret
  put/get/delete, policy create/get by name, cipher round trip; asserts
  on stdout now that data goes there. Cleans up its own artifacts.
- Pilot warns, without hanging or panicking, when Nexus is unreachable
  (`TestPilotWarnsWhenNexusUnreachable`, non-destructive): a single
  invocation is pointed at a closed port via `SPIKE_NEXUS_API_URL`,
  leaving the running Nexus untouched. The no-hang assertion doubles as
  a guard on the still-open SVID-acquisition-timeout task.
- Pilot denies operations when Nexus is uninitialized
  (`TestPilotDeniesWhenNexusUninitialized`, destructive): the only way
  to reach a reachable-but-uninitialized Nexus is to kill Nexus and every
  Keeper (losing the in-memory shards) and restart Nexus alone. This case
  is therefore gated behind a second flag, `SPIKE_INTEGRATION_DESTRUCTIVE=1`.
  The restarted Nexus is spawned outside `make start`'s process table, so
  the test kills it on cleanup; resetting is then a Ctrl+C on the make start
  terminal (or `make kill`) followed by `make start`.

Run the non-destructive cases with `make integration-test`; add the
destructive case with `make integration-test-destructive`.

### Slice C: HTTP-mock helpers for CLI coverage (separate task)

Mocking the SDK's mTLS transport feeds the "CLI coverage to 60%+"
task, not this one; it is out of scope here and tracked separately.

## Open questions (resolved 2026-07-18)

1. In-repo now, or wait for a SPIRE CI runner? **Land now, opt-in.** The
   suite lives in the repository behind the `integration` build tag and
   `SPIKE_INTEGRATION_TEST=1`, run manually against `make start` (the same
   footing as the recovery drill) until a SPIRE-capable CI job exists. The
   rot risk is accepted for now; `make integration-test` keeps it one
   command away.
2. Subsume the recovery drill, or stay complementary? **Complementary.**
   The drill kills real processes and proves restore end to end; the
   tagged tests assert the Pilot's observable behavior. The one
   destructive test (uninitialized Nexus) reuses the drill's kill and
   restart-Nexus-alone machinery but stops short of a restore.

## Acceptance Criteria

- [x] Slice A package passes in the normal `make test` run, isolated
      from `~/.spike`, leaving no artifacts.
- [x] The "not re-initialized twice" invariant has an explicit test.
- [x] The shard-restore round trip has an in-process test mirroring
      the live drill's semantics.
- [x] Secret and policy CRUD paths are covered end to end at the state
      layer, including deletion and undeletion.
- [x] Slice B lands only after the open questions are answered.
- [x] Slice B smoke pass and Nexus-unreachable warning pass live against
      `make start` (verified 2026-07-18), asserting secret data on stdout.
- [x] Slice B uninitialized-Nexus denial is covered by a destructive,
      double-gated test that cleans up the Nexus it spawns and leaves the
      environment resettable via a Ctrl+C on the make start terminal (or
      `make kill`), then `make start`.
