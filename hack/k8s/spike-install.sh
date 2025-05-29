#    \\ SPIKE: Secure your secrets with SPIFFE.
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

cd ./k8s/local/0.4.0/ || exit 1

echo "SPIKE install..."
echo "Current context: $(kubectl config current-context)"

create_namespace_if_not_exists() {
    local ns=$1
    echo "Checking namespace '$ns'..."

    # More explicit check
    if kubectl get namespace "$ns" 2>/dev/null | grep -q "$ns"; then
        echo "Namespace '$ns' already exists, skipping..."
    else
        echo "Creating namespace '$ns'..."
        kubectl create namespace "$ns"
        # shellcheck disable=SC2181
        if [ $? -eq 0 ]; then
            echo "Successfully created namespace '$ns'"
        else
            echo "Failed to create namespace '$ns'"
            return 1
        fi
    fi
}

create_namespace_if_not_exists "spike-control" # Pilot
create_namespace_if_not_exists "spike-system"  # Nexus
create_namespace_if_not_exists "spike-edge"    # Keepers

# List all namespaces after creation
echo "SPIKE namespaces:"
kubectl get namespaces | grep spike || echo "No spike namespaces found"

echo "Deploying SPIKE."
kubectl apply -f .

echo "Everything is awesome!"
