#!/usr/bin/env bash

#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

set -e  # Exit on any error

sudo microk8s reset --destroy-storage
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
  --ignore-not-found
#
kubectl delete clusterrolebinding \
   spire-agent \
   spire-controller-manager \
   spire-mgmt-spire-controller-manager \
   spire-mgmt-spire-server \
   spire-server-pre-upgrade \
   spire-system-spire-controller-manager \
   spire-system-spire-server \
   --ignore-not-found
#
kubectl delete csidriver csi.spiffe.io --ignore-not-found
#
kubectl delete validatingwebhookconfiguration \
  spire-controller-manager-webhook \
  spire-mgmt-spire-controller-manager-webhook \
  --ignore-not-found

#kubectl get clusterrole -o name | grep spire | xargs -r kubectl delete --ignore-not-found
#kubectl get clusterrolebinding -o name | grep spire | xargs -r kubectl delete --ignore-not-found
#kubectl delete csidriver csi.spiffe.io --ignore-not-found
#kubectl get validatingwebhookconfiguration -o name | grep spire | xargs -r kubectl delete --ignore-not-found
#kubectl get mutatingwebhookconfiguration -o name | grep spire | xargs -r kubectl delete --ignore-not-found


echo "Everything is awesome!"