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
- [ ] Fix broken recovery/restore flow #source:jira.xml #added:2026-07-14 (code review 2026-07-16: no live breakage found; awaiting the scripted drill in Phase 3 to confirm and close)
- [x] Fix `spike cipher` stream mode (broken; owner: Murat); JSON mode fix unblocks encryption-as-a-service demo/docs #source:jira.xml #added:2026-07-14 #done:2026-07-15 (stale: both cipher streaming and file modes verified passing via the make start checks on 2026-07-15)
- [ ] Retry sqlite operations with exponential backoff on transient locks (all of `app/nexus/internal/state/persist`) → ideas/research-db-resilience.md #source:jira.xml #added:2026-07-14
- [ ] Bound the Bootstrap keeper-wait loop with a configurable timeout/max-attempts instead of looping forever → ideas/research-db-resilience.md #source:jira.xml #added:2026-07-14
- [ ] Make `env` accessors return sentinel errors instead of calling `log.FatalLn` (removes env→log circular dep, makes them testable) → ideas/research-env-error-handling.md #source:jira.xml #added:2026-07-14

### Phase 2: SDK Extraction `#priority:medium`
- [ ] Graduate generic internal helpers to spike-sdk-go (nonce/crypto, Shamir verify, permission (de)serialize, validation, trust/spiffeid, URL builders, `GCMNonceSize`, `Id()`, canonical permission set) → ideas/research-sdk-extraction.md #source:jira.xml #added:2026-07-14

### Phase 3: Testing `#priority:medium`
- [ ] Make `make test` concurrent again (currently serialized by env setup) → ideas/research-cli-testing.md #source:jira.xml #added:2026-07-14
- [ ] Add integration tests: root key cached/recovered/not-re-initialized; secret & policy CRUD; Pilot denies when Nexus uninitialized / warns when unreachable → ideas/research-cli-testing.md #source:jira.xml #added:2026-07-14
- [ ] Raise CLI command coverage to 60%+ via unit + HTTP-mock tests; fix `t.Skip()`ed tests; DI-refactor `sendShardsToKeepers` → ideas/research-cli-testing.md #source:jira.xml #added:2026-07-14
- [ ] `start.sh` should exercise recovery/restore and encryption/decryption #source:jira.xml #added:2026-07-14
- [ ] Scripted live recovery/restore drill: once `make start` completes cleanly, run `spike operator recover`, kill Nexus and the Keepers, restart Nexus alone, feed the shards back via `spike operator restore` (scriptable via stdin since fix/operator-restore), and verify a pre-crash secret reads back. Rationale: the 2026-07-16 code review found no live breakage (shard-index fidelity intact end to end; guards use exact SPIFFE role matching, unaffected by the policy-name migration), so only a drill can prove the Phase 1 "recovery/restore is broken" claim stale and close both tasks. Needs the recover/restore role entries (spire-server-entry-recover-register.sh / -restore-register.sh), which make start does not register by default. #added:2026-07-16

### Phase 4: Policy & Secrets `#priority:medium`
- [ ] Add a `list` permission type; scope `spike secret list` to the caller's allowed path patterns → ideas/research-list-permission.md #source:jira.xml #added:2026-07-14
- [x] Add `LoadPolicyByName` for indexed O(1) lookup (UpsertPolicy currently scans all policies) #source:jira.xml #added:2026-07-14 #done:2026-07-14 #branch:fix/upsert-policy-indexed-lookup (used the existing `LoadPolicy(ctx, name)` indexed lookup; no new method needed)
- [ ] Reject empty id/path for both policies and secrets #source:jira.xml #added:2026-07-14
- [ ] Substring/partial-match filtering for `policy list` and `secret list` (today requires the exact regex) #source:jira.xml #added:2026-07-14
- [ ] Refactor `app/spike/internal/cmd/secret/get.go` to extract the repeated marshal-and-print helpers #source:jira.xml #added:2026-07-14

### Phase 5: Ops & Config `#priority:low`
- [ ] Wire the SDK's `env.DatabaseOperationTimeoutVal()` (spike-sdk-go config/env/database.go) into the Nexus DB path; SPIKE does not call it yet → ideas/research-db-resilience.md #source:jira.xml #added:2026-07-14 (retargeted 2026-07-14: the old in-repo accessor was removed and the accessor now lives in the SDK)
- [-] Move `const maxShardID = 1000` to env-var configuration #source:jira.xml #added:2026-07-14 #skipped:2026-07-14 (obsolete: the constant no longer exists anywhere in the repo or SDK)
- [ ] Add a configurable SVID-source acquisition timeout (go-spiffe context) so lookups don't wait forever #source:jira.xml #added:2026-07-14
- [ ] Add a Nexus `/status` endpoint; Pilot warns when Nexus not ready; Bootstrap uses it instead of the k8s Jobs API #source:jira.xml #added:2026-07-14

## Blocked

### Maintenance

- [x] Clear remaining uncalled govulncheck advisories (10 import + 4 module) via a dependency-bump pass #priority:low #session:6d40ae08 #branch:chore/introduce-ctx #commit:ca9b541 #added:2026-06-13-130622 #done:2026-06-13 (x/crypto v0.52.0, grpc v1.79.3; govulncheck now 0 total)
- [ ] Wire `SPIKE_NEXUS_PBKDF2_ITERATION_COUNT` into the crypto path: the constant `NexusPBKDF2IterationCount` is defined (spike-sdk-go config/env/env.go:45) and documented (docs-src/content/usage/configuration.md, default 600000) but no code reads it (no `…Val()` accessor, no os.Getenv). Add the accessor + consumer so the documented option takes effect, or escalate upstream to the SDK. #priority:low #session:00ce042d #branch:chore/introduce-ctx #added:2026-06-15 (found during docs-vs-code config audit)
- [x] Fix Docs Link Check so PRs validate internal `https://spike.ist/...` links against the checked-out `docs/` artifact while pushes to `main` keep production-origin behavior. Spec: `specs/docs-link-check-local-artifacts.md`. Issue: #288. #priority:high #branch:fix/docs-link-check-local-artifacts #added:2026-06-21 #done:2026-06-21
- [x] Fix Docs Link Check fork PR sticky-comment permission failure by granting explicit read/comment permissions for trusted PRs and skipping comment writes for fork PRs. Spec: `specs/docs-link-check-local-artifacts.md`. Issue: #288. #priority:high #branch:fix/docs-link-check-local-artifacts #added:2026-06-21 #done:2026-06-21
