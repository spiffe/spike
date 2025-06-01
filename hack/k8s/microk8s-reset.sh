#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

set -e  # Exit on any error

echo "Resetting the MicroK8s Test Cluster... This will take a while; go grab a sandwich."

# helm delete -n spire-mgmt spire

sudo microk8s reset # --destroy-storage
# rm -rf ~/.cache/helm ~/.config/helm ~/.local/share/helm

# Enable alias expansion in non-interactive shells
shopt -s expand_aliases

source ./hack/lib/aliases.sh
source ./hack/lib/env-k8s.sh

microk8s enable metallb:"$METALLB_IP_RANGE_START"-"$METALLB_IP_RANGE_END"
microk8s enable dns
microk8s enable hostpath-storage
microk8s enable registry

kubectl delete clusterrole spire-agent \
  spire-controller-manager \
  spire-mgmt-spire-controller-manager \
  spire-mgmt-spire-server \
  spire-server-pre-upgrade \
  spire-system-spire-controller-manager \
  spire-system-spire-server \
  spire-server-spire-controller-manager \
  spire-server-spire-server \
  --ignore-not-found

kubectl delete clusterrolebinding \
   spire-agent \
   spire-controller-manager \
   spire-mgmt-spire-controller-manager \
   spire-mgmt-spire-server \
   spire-server-pre-upgrade \
   spire-system-spire-controller-manager \
   spire-system-spire-server \
   spire-server-spire-controller-manager \
   spire-server-spire-server \
   --ignore-not-found

kubectl delete csidriver csi.spiffe.io --ignore-not-found

kubectl delete validatingwebhookconfiguration \
  spire-controller-manager-webhook \
  spire-mgmt-spire-controller-manager-webhook \
  spire-server-spire-controller-manager-webhook \
  spire-system-spire-controller-manager-webhook \
  --ignore-not-found



echo "Everything is awesome!"
