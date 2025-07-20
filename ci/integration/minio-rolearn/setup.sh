#!/bin/bash

SCRIPT="$(readlink -f "$0")"
SCRIPTPATH="$(dirname "${SCRIPT}")"
BASEPATH="${SCRIPTPATH}/../../../"

helm upgrade --install -n spire-server spire-crds spire-crds --repo https://spiffe.github.io/helm-charts-hardened/ --create-namespace
helm upgrade --install -n spire-server spire /home/kfox/git/helm-charts-hardened4/charts/spire -f "${SCRIPTPATH}/spire-values.yaml" --wait
kubectl apply -f "${SCRIPTPATH}/test.yaml"
helm upgrade --install minio -n minio --create-namespace oci://registry-1.docker.io/bitnamicharts/minio -f "${SCRIPTPATH}/minio-values.yaml"
kubectl rollout restart -n minio deployment/minio
kubectl rollout status -n minio deployment/minio
