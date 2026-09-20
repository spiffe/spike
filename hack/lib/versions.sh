#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# The single source of truth for the SPIRE versions SPIKE is developed and
# tested against. Every script that installs or builds SPIRE sources this
# file; the documentation pages that show these versions are checked against
# it by internal/layout's tests, so a bump is: change the values here, run
# make test, and fix the doc lines the test names.
#
# Keep SPIRE_VERSION equal to the spire chart's appVersion so that the
# bare-metal build and the Kubernetes deployment run the same SPIRE.

# The spire Helm chart (https://spiffe.github.io/helm-charts-hardened/).
SPIRE_HELM_CHART_VERSION="0.30.2"

# The spire-crds Helm chart from the same repository.
SPIRE_CRDS_HELM_CHART_VERSION="0.6.1"

# The SPIRE release tag built for the bare-metal environment.
SPIRE_VERSION="v1.15.3"
