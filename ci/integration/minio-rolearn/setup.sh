#!/bin/bash -e

SCRIPT="$(readlink -f "$0")"
SCRIPTPATH="$(dirname "${SCRIPT}")"
BASEPATH="${SCRIPTPATH}/../../../"

helm upgrade --install -n spire-server spire-crds spire-crds --repo https://spiffe.github.io/helm-charts-hardened/ --create-namespace
helm upgrade --install -n spire-server spire spire --repo https://spiffe.github.io/helm-charts-hardened/ -f "${SCRIPTPATH}/spire-values.yaml" --wait
#FIXME remove once upstream chart supports this
kubectl patch statefulset -n spire-server spire-spike-nexus --type='strategic' -p '
spec:
  template:
    spec:
      containers:
      - name: spire-spike-nexus
        env:
        - name: SPIKE_NEXUS_BACKEND_STORE
          value: lite
        - name: SPIKE_TRUST_ROOT_LITE_WORKLOAD
          value: example.org
'
kubectl rollout status statefulset/spire-spike-nexus -n spire-server --watch --timeout=5m
kubectl apply -f "${SCRIPTPATH}/test.yaml"
helm upgrade --install minio -n minio --create-namespace oci://registry-1.docker.io/bitnamicharts/minio -f "${SCRIPTPATH}/minio-values.yaml"
kubectl rollout restart -n minio deployment/minio
kubectl rollout status -n minio deployment/minio
