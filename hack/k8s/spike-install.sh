#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

cd ./k8s/local/0.4.0/ || exit 1

create_namespace_if_not_exists() {
    local ns=$1
    if kubectl get namespace "$ns" &> /dev/null; then
        echo "Namespace '$ns' already exists, skipping..."
    else
        echo "Creating namespace '$ns'..."
        kubectl create namespace "$ns"
    fi
}

create_namespace_if_not_exists "spike-control"
create_namespace_if_not_exists "spike-system"
create_namespace_if_not_exists "spike-edge"

kubectl apply -f .
