#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# This is temporary, until we enhance the upstream charts.

cp ./hack/k8s/patch/spike-nexus/service.yaml \
  ../helm-charts-hardened/charts/spire/charts/spike-nexus/templates/service.yaml

cp ./hack/k8s/patch/spike-nexus/statefulset.yaml \
  ../helm-charts-hardened/charts/spire/charts/spike-nexus/templates/statefulset.yaml

cp ./hack/k8s/patch/spike-keeper/statefulset.yaml \
  ../helm-charts-hardened/charts/spire/charts/spike-keeper/templates/statefulset.yaml