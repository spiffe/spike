#!/bin/bash -e

SCRIPT="$(readlink -f "$0")"
SCRIPTPATH="$(dirname "${SCRIPT}")"
BASEPATH="${SCRIPTPATH}/../../../"

# Chart versions are pinned so a CI run tests a known chart, not whatever
# upstream published last. They come from the single source of truth.
# shellcheck source=hack/lib/versions.sh
. "${BASEPATH}/hack/lib/versions.sh"

helm upgrade --install -n spire-server spire-crds spire-crds \
  --repo https://spiffe.github.io/helm-charts-hardened/ \
  --version "${SPIRE_CRDS_HELM_CHART_VERSION}" --create-namespace

# The spike-nexus subchart runs the SPIKE Bootstrap as a post-install Helm
# hook with the dev bootstrap image loaded into kind (see spire-values.yaml).
# The bootstrap waits until every Keeper listens before it produces a single
# share, so a retried hook cannot leave the Keepers holding two keys. Helm
# waits for the hook; the Keepers need their SVIDs first, hence the timeout.
helm upgrade --install -n spire-server spire spire \
  --repo https://spiffe.github.io/helm-charts-hardened/ \
  --version "${SPIRE_HELM_CHART_VERSION}" --timeout 10m \
  -f "${SCRIPTPATH}/spire-values.yaml"

# The chart has no value for the lite-workload trust root yet (upstream ask
# tracked in .context/TASKS.md), so it is patched in. The backend store is
# set through the chart's backendStore value.
kubectl patch statefulset -n spire-server spire-spike-nexus \
  --type='strategic' -p '
spec:
  template:
    spec:
      containers:
      - name: spire-spike-nexus
        env:
        - name: SPIKE_TRUST_ROOT_LITE_WORKLOAD
          value: example.org
'
kubectl rollout status statefulset/spire-spike-nexus -n spire-server \
  --watch --timeout=5m
kubectl apply -f "${SCRIPTPATH}/test.yaml"
# Pin the chart version so the image tags stay aligned with the tags that
# exist under docker.io/bitnamilegacy/* (see minio-values.yaml). An unpinned
# install would float to the latest chart, whose newer image tags may not be
# mirrored in the frozen legacy repository.
helm upgrade --install minio -n minio --create-namespace --version 17.0.21 \
  oci://registry-1.docker.io/bitnamicharts/minio \
  -f "${SCRIPTPATH}/minio-values.yaml"
kubectl rollout restart -n minio deployment/minio
kubectl rollout status -n minio deployment/minio
kubectl wait -l statefulset.kubernetes.io/pod-name=test-0 \
  --for=condition=ready pod --timeout=-360s
