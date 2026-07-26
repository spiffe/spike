# Learnings

<!--
UPDATE WHEN:
- Discover a gotcha, bug, or unexpected behavior
- Debugging reveals non-obvious root cause
- External dependency has quirks worth documenting
- "I wish I knew this earlier" moments
- Production incidents reveal gaps

DO NOT UPDATE FOR:
- Well-documented behavior (link to docs instead)
- Temporary workarounds (use TASKS.md for follow-up)
- Opinions without evidence
-->

<!-- INDEX:START -->
| Date | Learning |
|----|--------|
| 2026-06-13 | SPIKE k8s integration test was missing keeper bootstrap; plus a verify-path deadlock |
| 2026-06-13 | Zola 0.19+/0.22 moved syntax highlighting config and renamed themes |
<!-- INDEX:END -->

<!-- Add gotchas, tips, and lessons learned here -->
## [2026-07-25-194538] CI's Go - Lint job runs make audit, not make lint-go

**Context**: CI on main failed at the 'Go - Lint' job while local 'make lint-go' was green throughout. The job name suggests golangci-lint, but .github/workflows/ci.yaml runs 'make audit'.

**Lesson**: 'make audit' is a superset: go mod tidy -diff, go mod verify, gofmt check, go vet, staticcheck, govulncheck, and a CGO_ENABLED=0 golangci-lint run. 'make lint-go' is only the plain golangci-lint. A green 'make lint-go' therefore proves nothing about the CI lint job. Because govulncheck is in there, that job can also fail on a timer when a new advisory drops against an unchanged dependency graph, with no code change involved.

**Application**: Run 'make audit' (not just 'make lint-go') before pushing anything that could touch the module graph, and when CI's lint job fails unexpectedly, check govulncheck first and compare go.mod between the passing and failing runs before assuming a code change caused it.

---

## [2026-07-25-133240] Policy regex patterns are compiled at two sites, not one

**Context**: Evaluating a proposed security patch that anchored policy patterns inside UpsertPolicy. It passed its own test but changed nothing in production.

**Lesson**: UpsertPolicy in app/nexus/internal/state/base/policy.go compiles patterns once at authoring time, but app/nexus/internal/state/backend/sqlite/persist/regex.go recompiles them from the stored strings on EVERY load. CheckPolicyAccess -> ListPolicies -> LoadAllPolicies goes through that second site on every access check. The memory backend retains the struct UpsertPolicy built, so memory-backed tests hide the difference entirely.

**Application**: Any change to how policy patterns are compiled, anchored, or validated must cover both compile sites, or be enforced at CheckPolicyAccess where all paths converge. Always add a SQLite-backed test alongside the memory-backed one; a green memory test proves nothing about production.

---

## [2026-07-18-110741] spiffe.Source without a SPIRE agent hangs forever; a malformed endpoint fails fast

**Context**: The lifecycle integration test hung 148s at RestoreBackingStoreFromPilotShards' source creation: context.Background() with no dial timeout (the open Phase 5 SVID-timeout task, met in the wild).

**Lesson**: SPIFFE_ENDPOINT_SOCKET=bogus://fail-fast makes source creation fail at address validation, instantly and deterministically.

**Application**: Use the malformed endpoint in any test driving code that reaches for a SPIFFE source; pair with SPIKE_STACK_TRACES_ON_LOG_FATAL=true to recover the fatal.

---

## [2026-07-18-110741] SDK config/fs path resolvers are sync.Once-memoized; env overrides must happen in TestMain

**Context**: The test-isolation work found per-test t.Setenv(SPIKE_NEXUS_DATA_DIR, ...) silently ineffective after the first resolution.

**Lesson**: fs.NexusDataFolder and siblings memoize on first call for the process lifetime.

**Application**: Any package whose tests touch SPIKE data or recovery paths needs the env override in TestMain before m.Run(), one temporary directory per package run.

---

## [2026-07-18-110741] make audit lint runs with CGO_ENABLED=0, so typed sqlite error inspection breaks it

**Context**: The sqlite retry work used sqlite3.Error/ErrBusy from mattn/go-sqlite3; build and tests passed, then the gate failed: golangci-lint typechecks with CGO off, where mattn's typed API does not exist.

**Lesson**: Detect transient sqlite failures by driver error strings (database is locked / database table is locked), never by mattn types.

**Application**: Keep mattn/go-sqlite3 symbol references out of files that are not cgo-gated; anything else fails make audit.

---

## [2026-06-13-170816] SPIKE k8s integration test was missing keeper bootstrap; plus a verify-path deadlock

**Context**: minio-rolearn integration test (CI red on main, pre-existing) hangs because keepers are never seeded with root-key shares; SPIKE Nexus InitializeBackingStoreFromKeepers waits forever (retry.Forever, by design until keepers are hydrated). The spire helm chart registers the spike/bootstrap identity but ships no bootstrap workload, and hack/k8s/Bootstrap.yaml does not exist.

**Lesson**: Reproduced locally (kind on colima: needed fs.inotify.max_user_instances bump 128->8192, buildx via arch -arm64, cel CredentialComposer disabled on arm64). Fix: add a spike-bootstrap Job (component=spike-bootstrap label -> spiffe://<td>/spike/bootstrap SVID) + SA/RBAC for the spike-bootstrap-state idempotency ConfigMap, set SPIKE_NEXUS_KEEPER_PEERS, SPIKE_NEXUS_SHAMIR_SHARES/THRESHOLD, SPIKE_NEXUS_API_URL (FQDN .svc.cluster.local; short svc.ns is NXDOMAIN), and ALL trust roots incl SPIKE_TRUST_ROOT_NEXUS (AllowNexus->IsNexus reads it). Also found a real deadlock: app/bootstrap/internal/net/broadcast.go VerifyInitialization took write lock (LockRootKeySeed) then read lock (RootKeySeed) on the same RWMutex -> use RootKeySeedNoLock. RESOLVED: the post-seed api.Verify 400 had two causes. (1) The probe was encrypted with the raw RootKeySeed, but Keepers/Nexus key their cipher with the canonical root key = ComputeShares(seed) scalar MarshalBinary (the seed is reduced mod the P256 order, so seed != canonical key); fix: encrypt with crypto.ComputeShares(seed).MarshalBinary(). (2) The real blocker: Nexus registers /v1/bootstrap/verify ONLY in routeWithBackingStore; in lite mode (routeWithNoBackingStore) it fell through to net.Fallback -> 400, so RouteVerify never ran (audit logs "enter/exit success" even on the fallback 400, which is misleading). Fix: register NexusBootstrapVerify->bootstrap.RouteVerify in routeWithNoBackingStore too (lite mode holds the root key and already exposes cipher/operator routes). Verified end to end: Job succeeded, idempotency ConfigMap written, Nexus Ready in ~30s. Debugging gotcha: kind+imagePullPolicy:Never reuses cached digests for the SAME tag even after kind load; use a UNIQUE image tag and patch the statefulset to force a fresh pull.

**Application**: bootstrap seeding makes Nexus reach Ready (integration test passes; setup.sh gates on Nexus rollout not the Job). Files: ci/integration/minio-rolearn/bootstrap.yaml (new), setup.sh (drop --wait + apply bootstrap), app/bootstrap/internal/net/broadcast.go (deadlock fix). Verify-completion still open.

---

## [2026-06-13-123540] Zola 0.19+/0.22 moved syntax highlighting config and renamed themes

**Context**: make docs failed on Zola 0.22.1: 'unknown field highlight_code' then 'Theme base16-ocean-dark does not exist'. docs-src/config.toml used pre-0.19 highlighting keys.

**Lesson**: Zola 0.19 moved highlight_code/highlight_theme out of [markdown] into a [markdown.highlighting] table (theme/light_theme/dark_theme/style). Zola 0.22 swapped Syntect for the Giallo highlighter, so Syntect theme names like base16-ocean-dark are gone; valid names come from the Giallo bundle (getzola/giallo, sourced from shikijs/textmate-grammars-themes), e.g. material-theme-ocean, one-dark-pro, github-dark, nord.

**Application**: When a Zola upgrade breaks 'make docs', migrate config.toml highlighting into [markdown.highlighting] and map old theme names to Giallo identifiers. List valid themes via: gh api repos/getzola/giallo/readme --jq .content | base64 -d. base16-ocean-dark -> material-theme-ocean is the closest match for SPIKE.
