# Spec: Repair minio-rolearn integration test after Bitnami image relocation

## Problem Statement

The `Integration Test` CI job (`ci/integration/minio-rolearn`) is red.
Unlike the earlier keeper-bootstrap failure
([integration-bootstrap-fix.md](integration-bootstrap-fix.md)), the SPIKE
stack itself is healthy: keepers (x3), Nexus, Pilot all `Running`, the
bootstrap Job `Completed`. The job dies in `setup.sh` at the MinIO Helm
install:

```
Error: failed post-install: 1 error occurred:
        * timed out waiting for the condition
```

Root cause: every MinIO pod is `ImagePullBackOff`. Bitnami removed its
free `docker.io/bitnami/*` images from Docker Hub and relocated them to
`docker.io/bitnamilegacy/*`. Verified against the live registry:

- `bitnami/minio:2025.7.23-debian-12-r3` -> HTTP 404
- `bitnamilegacy/minio:2025.7.23-debian-12-r3` -> HTTP 200

`setup.sh` installs the chart with **no pinned version**, so it floats to
the latest chart (17.0.21) whose default image tags are the deleted ones.
Helm's post-install `--wait` then times out after 5 minutes and the job
exits 1.

The chart pulls four images, all now 404 under `bitnami/` and present
under `bitnamilegacy/`: `minio`, `minio-client`, `minio-object-browser`
(console), and `os-shell` (provisioning init container).

## Proposed Solution

1. **Redirect every chart image to the legacy repository** in
   `minio-values.yaml`: `image`, `clientImage`, `console.image`, and
   `defaultInitContainers.volumePermissions.image` -> `bitnamilegacy/*`.
   Set `global.security.allowInsecureImages: true`; the chart's
   image-verification gate rejects non-`bitnami/` repositories otherwise.
2. **Pin the chart version** to `17.0.21` in `setup.sh` so the image tags
   stay aligned with the tags mirrored in the frozen legacy repository.
   An unpinned install would float to a newer chart whose tags may not
   exist under `bitnamilegacy/`.

## File Surface

- `ci/integration/minio-rolearn/minio-values.yaml`: legacy image repos +
  `allowInsecureImages`.
- `ci/integration/minio-rolearn/setup.sh`: pin `--version 17.0.21`.

## Non-Goals

- No SPIKE code changes; the stack itself was healthy.
- `bitnamilegacy/*` is a frozen, unmaintained repository. A future
  migration off Bitnami images (or to a self-hosted mirror) is out of
  scope here; this spec only restores a green CI job.

## Verification

- Live-registry manifest checks: all four `bitnamilegacy/*` tags return
  HTTP 200; the `bitnami/*` originals return 404.
- `helm template` with the pinned version and new values renders cleanly
  with no image-verification error; every rendered image resolves to
  `docker.io/bitnamilegacy/*` with zero leftover `docker.io/bitnami/*`
  references.
- `make lint-go` + `make test` pass on the branch.
