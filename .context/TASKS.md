# Tasks

<!--
UPDATE WHEN:
- New work is identified → add task with #added timestamp
- Starting work → add #in-progress or #started timestamp
- Work completes → mark [x]
- Work is blocked → add to Blocked section with reason
- Scope changes → update task description inline

DO NOT UPDATE FOR:
- Reorganizing or moving tasks (violates CONSTITUTION)
- Removing completed tasks (use ctx task archive instead)

STRUCTURE RULES (see CONSTITUTION.md):
- Tasks stay in their Phase section permanently: never move them
- Use inline labels: #in-progress, #blocked, #priority:high
- Mark completed: [x], skipped: [-] (with reason)
- Never delete tasks, never remove Phase headers

TASK STATUS LABELS:
  `[ ]`: pending
  `[x]`: completed
  `[-]`: skipped (with reason)
  `#in-progress`: currently being worked on (add inline, don't move task)
-->

<!--
Phases below seeded 2026-07-14 from a triage of the old jira.xml sandbox.
Summaries only; detail lives in gitignored ideas/research-*.md. Broader,
non-task ideas live in ideas/*.md for serendipity. Stale/already-done sandbox
items (contributeWithContext, policy-ID duplicates, validUUID, several test
fixes) were dropped, having been superseded by the v0.19.9 SDK migration and
the name-based policy work.
-->

### Phase 1: Correctness & Broken Things `#priority:high`
- [x] Fix broken CI integration test #source:jira.xml #added:2026-07-14 #done:2026-07-16 (stale: CI has been green for several weeks; closed on user confirmation)
- [x] Fix broken recovery/restore flow #source:jira.xml #added:2026-07-14 #done:2026-07-16 (the drill exposed a real boot-order deadlock: an un-initialized Nexus never started its listener, so the emergency restore route was unreachable when every Keeper died; fixed by recovering from Keepers in the background, and the recovery loop now stands down once an operator restore supplies the root key. Verified live by make drill-recovery)
- [x] Fix `spike cipher` stream mode (broken; owner: Murat); JSON mode fix unblocks encryption-as-a-service demo/docs #source:jira.xml #added:2026-07-14 #done:2026-07-15 (stale: both cipher streaming and file modes verified passing via the make start checks on 2026-07-15)
- [x] Retry sqlite operations with exponential backoff on transient locks → ideas/research-db-resilience.md #source:jira.xml #added:2026-07-14 #done:2026-07-16 (withSerializableTx retries SQLITE_BUSY/SQLITE_LOCKED with exponential backoff at the single choke point every write flows through; reads rely on WAL plus the busy_timeout DSN parameter and honor the operation deadline; note the DB ops live under state/backend/sqlite/persist these days, not state/persist)
- [x] Bound the Bootstrap keeper-wait loop with a configurable timeout/max-attempts instead of looping forever → ideas/research-db-resilience.md #source:jira.xml #added:2026-07-14 #done:2026-07-16 (stale: superseded by the SDK retry migration; broadcastToKeeper bounds each keeper with retry.WithMaxAttempts, a per-keeper context timeout, and configurable backoff intervals — app/bootstrap/internal/net/dispatch.go — and broadcast.go bounds init verification with WithMaxElapsedTime)
- [-] Make `env` accessors return sentinel errors instead of calling `log.FatalLn` → ideas/research-env-error-handling.md #source:jira.xml #added:2026-07-14 #skipped:2026-07-17 (working as intended: the crash-fast behavior on missing critical config such as SPIKE_NEXUS_KEEPER_PEERS is deliberate, since SPIKE cannot operate reliably without it; the accessors ARE testable via the SPIKE_STACK_TRACES_ON_LOG_FATAL panic-recover pattern; and the env-to-log circular dependency the jira item cited dissolved when both packages moved into spike-sdk-go)
- [x] Fix Pilot printing normal output to stderr: cobra Print* writes to OutOrStderr and the root command never calls SetOut, so data output (secrets included) lands on stderr and `spike secret get x > file.txt` yields an empty file; every harness script compensates with 2>&1. #added:2026-07-16 #done:2026-07-16 (rootCmd.SetOut(os.Stdout) in cmd.Initialize; PrintErr still goes to stderr, and the harness scripts that merge streams keep working)

### Phase 2: SDK Extraction `#priority:medium`
- [ ] Graduate generic internal helpers to spike-sdk-go (nonce/crypto, Shamir verify, permission (de)serialize, validation, trust/spiffeid, URL builders, `GCMNonceSize`, `Id()`, canonical permission set) → ideas/research-sdk-extraction.md #source:jira.xml #added:2026-07-14

### Phase 3: Testing `#priority:medium`
- [x] Make `make test` concurrent again → ideas/research-cli-testing.md #source:jira.xml #added:2026-07-14 #done:2026-07-17 (removed -p 1 once the data-dir isolation landed; nothing else shared state across packages — no fixed ports, no t.Parallel, env vars are per-process. Full -race suite: 29.7s serialized to 2.9s concurrent, roughly 10x; two consecutive concurrent runs clean)
- [x] Move the sqlite state tests off the real ~/.spike/data/spike.db: make test deleted the live dev environment database mid-run (bit us three times this week). #added:2026-07-16 #done:2026-07-16 (fs.NexusDataFolder is sync.Once-memoized, so per-test t.Setenv cannot work; instead each affected package sets SPIKE_NEXUS_DATA_DIR to a per-run temp dir in TestMain before the first resolution — state/base, state/persist, and backend/sqlite/persist — verified: a full package run leaves ~/.spike untouched)
- [x] Add integration tests: root key cached/recovered/not-re-initialized; secret & policy CRUD; Pilot denies when Nexus uninitialized / warns when unreachable → ideas/research-cli-testing.md #source:jira.xml #added:2026-07-14 #done:2026-07-18 (2026-07-17: Slice A shipped — specs/integration-tests.md; app/nexus/internal/state/integration covers the root-key lifecycle, the not-re-initialized-twice invariant, CRUD, and an in-process shard-restore round trip inside the normal suite. 2026-07-18: Slice B shipped — app/spike/internal/cmd/integration, build tag integration + SPIKE_INTEGRATION_TEST=1, drives the spike binary end to end. Non-destructive CRUD+cipher smoke pass (asserts secret data on stdout) and the Nexus-unreachable warning both verified live against make start; the uninitialized-Nexus denial is a destructive, double-gated test (SPIKE_INTEGRATION_DESTRUCTIVE=1) reusing the drill's kill/restart-Nexus-alone machinery; it cleans up the Nexus it spawns, so a Ctrl+C on the make start terminal (or make kill) then make start resets the env. Spec open questions resolved: land now opt-in, complementary to the drill. make targets: integration-test / integration-test-destructive)
- [ ] Raise CLI command coverage to 60%+ via unit + HTTP-mock tests; fix `t.Skip()`ed tests; DI-refactor `sendShardsToKeepers` → ideas/research-cli-testing.md #source:jira.xml #added:2026-07-14
- [x] `start.sh` should exercise recovery/restore and encryption/decryption #source:jira.xml #added:2026-07-14 #done:2026-07-16 (encryption/decryption checks live in start.sh since the policy-validation rework; recovery/restore is exercised by make drill-recovery, kept as a separate second-terminal script deliberately so the crash simulation never runs inside the normal startup path)
- [x] Scripted live recovery/restore drill: once `make start` completes cleanly, run `spike operator recover`, kill Nexus and the Keepers, restart Nexus alone, feed the shards back via `spike operator restore` (scriptable via stdin since fix/operator-restore), and verify a pre-crash secret reads back. Rationale: the 2026-07-16 code review found no live breakage (shard-index fidelity intact end to end; guards use exact SPIFFE role matching, unaffected by the policy-name migration), so only a drill can prove the Phase 1 "recovery/restore is broken" claim stale and close both tasks. Needs the recover/restore role entries (spire-server-entry-recover-register.sh / -restore-register.sh), which make start does not register by default. #added:2026-07-16 #done:2026-07-16 (implemented as hack/bare-metal/drill/recovery-drill.sh behind make drill-recovery; the drill first exposed the Nexus boot-order deadlock, then passed end to end once it was fixed)

### Phase 4: Policy & Secrets `#priority:medium`
- [ ] Add a `list` permission type; scope `spike secret list` to the caller's allowed path patterns → ideas/research-list-permission.md #source:jira.xml #added:2026-07-14
- [x] Add `LoadPolicyByName` for indexed O(1) lookup (UpsertPolicy currently scans all policies) #source:jira.xml #added:2026-07-14 #done:2026-07-14 #branch:fix/upsert-policy-indexed-lookup (used the existing `LoadPolicy(ctx, name)` indexed lookup; no new method needed)
- [ ] Reject empty id/path for both policies and secrets #source:jira.xml #added:2026-07-14
- [ ] Substring/partial-match filtering for `policy list` and `secret list` (today requires the exact regex) #source:jira.xml #added:2026-07-14
- [ ] Refactor `app/spike/internal/cmd/secret/get.go` to extract the repeated marshal-and-print helpers #source:jira.xml #added:2026-07-14

### Phase 5: Ops & Config `#priority:low`
- [x] Wire the SDK's `env.DatabaseOperationTimeoutVal()` (spike-sdk-go config/env/database.go) into the Nexus DB path → ideas/research-db-resilience.md #source:jira.xml #added:2026-07-14 #done:2026-07-16 (wired as a context deadline via operationContext in the sqlite persist layer: writes through withSerializableTx and all four read entry points honor SPIKE_NEXUS_DB_OPERATION_TIMEOUT, default 15s)
- [-] Move `const maxShardID = 1000` to env-var configuration #source:jira.xml #added:2026-07-14 #skipped:2026-07-14 (obsolete: the constant no longer exists anywhere in the repo or SDK)
- [ ] Add a configurable SVID-source acquisition timeout (go-spiffe context) so lookups don't wait forever #source:jira.xml #added:2026-07-14
- [ ] Add a Nexus `/status` endpoint; Pilot warns when Nexus not ready; Bootstrap uses it instead of the k8s Jobs API #source:jira.xml #added:2026-07-14

## Blocked

### Maintenance

- [x] Clear remaining uncalled govulncheck advisories (10 import + 4 module) via a dependency-bump pass #priority:low #session:6d40ae08 #branch:chore/introduce-ctx #commit:ca9b541 #added:2026-06-13-130622 #done:2026-06-13 (x/crypto v0.52.0, grpc v1.79.3; govulncheck now 0 total)
- [ ] Wire `SPIKE_NEXUS_PBKDF2_ITERATION_COUNT` into the crypto path: the constant `NexusPBKDF2IterationCount` is defined (spike-sdk-go config/env/env.go:45) and documented (docs-src/content/usage/configuration.md, default 600000) but no code reads it (no `…Val()` accessor, no os.Getenv). Add the accessor + consumer so the documented option takes effect, or escalate upstream to the SDK. #priority:low #session:00ce042d #branch:chore/introduce-ctx #added:2026-06-15 (found during docs-vs-code config audit)
- [x] Fix Docs Link Check so PRs validate internal `https://spike.ist/...` links against the checked-out `docs/` artifact while pushes to `main` keep production-origin behavior. Spec: `specs/docs-link-check-local-artifacts.md`. Issue: #288. #priority:high #branch:fix/docs-link-check-local-artifacts #added:2026-06-21 #done:2026-06-21
- [x] Fix Docs Link Check fork PR sticky-comment permission failure by granting explicit read/comment permissions for trusted PRs and skipping comment writes for fork PRs. Spec: `specs/docs-link-check-local-artifacts.md`. Issue: #288. #priority:high #branch:fix/docs-link-check-local-artifacts #added:2026-06-21 #done:2026-06-21
