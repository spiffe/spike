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

### Phase 1: [Name] `#priority:high`
- [ ] Task 1
- [ ] Task 2

### Phase 2: [Name] `#priority:medium`
- [ ] Task 1
- [ ] Task 2

## Blocked

### Maintenance

- [ ] test.
- [ ] Replace MinIO in minio-rolearn integration test with a maintained, permissively-licensed S3 store. The current bitnamilegacy/* pin (commit 4decfca) is a stopgap: that repo is frozen, and MinIO Community itself is AGPLv3, entered maintenance mode Dec 2025 / archived early 2026, and stopped publishing official minio/minio images (Docker Hub + Quay) in Oct 2025. Binding constraint: the test exercises S3 + OIDC AssumeRoleWithWebIdentity role-policy (mc idp openid against the SPIRE OIDC discovery provider), so any replacement MUST support OIDC/STS web-identity role mapping. Candidates (researched 2026-06-20): SeaweedFS (Apache-2.0, explicit AssumeRoleWithWebIdentity + OIDC docs, lightweight - evaluate FIRST); RustFS (Apache-2.0, Rust, MinIO drop-in, OIDC/STS unverified); Ceph RGW (strong OIDC STS but GPL + heavyweight, poor kind fit). Garage ruled out (AGPLv3). Action: PoC SeaweedFS S3 + AssumeRoleWithWebIdentity against SPIRE JWT-SVID in kind; if it works, port the minio-values.yaml policy/OIDC wiring and drop the Bitnami dependency entirely. #session:5a7938b8 #branch:chore/introduce-ctx #commit:4decfca #added:2026-06-20-124108

- [x] Clear remaining uncalled govulncheck advisories (10 import + 4 module) via a dependency-bump pass #priority:low #session:6d40ae08 #branch:chore/introduce-ctx #commit:ca9b541 #added:2026-06-13-130622 #done:2026-06-13 (x/crypto v0.52.0, grpc v1.79.3; govulncheck now 0 total)
- [ ] Wire `SPIKE_NEXUS_PBKDF2_ITERATION_COUNT` into the crypto path: the constant `NexusPBKDF2IterationCount` is defined (spike-sdk-go config/env/env.go:45) and documented (docs-src/content/usage/configuration.md, default 600000) but no code reads it (no `…Val()` accessor, no os.Getenv). Add the accessor + consumer so the documented option takes effect, or escalate upstream to the SDK. #priority:low #session:00ce042d #branch:chore/introduce-ctx #added:2026-06-15 (found during docs-vs-code config audit)
