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
| 2026-06-20 | Bitnami deleted docker.io/bitnami/* images; charts need bitnamilegacy/* + allowInsecureImages |
| 2026-06-13 | SPIKE k8s integration test was missing keeper bootstrap; plus a verify-path deadlock |
| 2026-06-13 | Zola 0.19+/0.22 moved syntax highlighting config and renamed themes |
<!-- INDEX:END -->

<!-- Add gotchas, tips, and lessons learned here -->
## [2026-06-20-121612] Bitnami deleted docker.io/bitnami/* images; charts need bitnamilegacy/* + allowInsecureImages

**Context**: The minio-rolearn integration test (CI) was red even though the whole SPIKE stack was healthy (keepers, Nexus, Pilot, bootstrap all up). Root cause: every MinIO pod hit ImagePullBackOff and Helm's post-install --wait timed out. Bitnami removed its free docker.io/bitnami/* images from Docker Hub and relocated them to docker.io/bitnamilegacy/*, so the chart's default image tags now 404.

**Lesson**: Any Helm chart that pulls default Bitnami images is now broken. The fix has three parts that must go together: (1) redirect every image the chart pulls to bitnamilegacy/* (for the minio chart: image, clientImage, console.image, defaultInitContainers.volumePermissions.image), (2) set global.security.allowInsecureImages=true because the chart's image-verification gate rejects non-bitnami/ repositories, and (3) pin the chart --version so the image tags stay aligned with the frozen, no-longer-updated legacy repo.

**Application**: When a Bitnami-backed chart fails with ImagePullBackOff, check the registry: docker.io/bitnami/<img>:<tag> returns 404 while docker.io/bitnamilegacy/<img>:<tag> returns 200. Override all image repos to bitnamilegacy/*, add allowInsecureImages=true, and pin the chart version. bitnamilegacy is frozen; a future migration off Bitnami images is the durable fix. See specs/minio-bitnami-image-relocation.md and ci/integration/minio-rolearn/minio-values.yaml.

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
