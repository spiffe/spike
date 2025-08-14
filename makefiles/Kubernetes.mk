#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# 0. Prune docker file system to save resources.
docker-cleanup:
	./hack/docker/cleanup.sh

# 1. Reset the test cluster.
k8s-delete:
	./hack/k8s/minikube-delete.sh

# 2. Start the test cluster.
k8s-start:
	./hack/k8s/minikube-start.sh

# Deletes and re-installs the Minikube cluster.
k8s-reset:
	k8s-delete
	k8s-start

# 3. Build container images.
docker-build:
	./hack/docker/build-local.sh

# 4. Forward registry.
docker-forward-registry:
	./hack/docker/minikube-forward-registry.sh

# 5. Push to the container registry.
docker-push:
	./hack/docker/push-local.sh

# For Multi-Cluster Federation Demo, DO NOT run `deploy-local`
# Instead, see FederationDemo.mk for the remaining steps.

# 6. (Single Cluster) Deploy SPIRE and SPIKE to the cluster.
deploy-local:
	./hack/k8s/spike-install.sh

# 6.x. TODO: new target to test the bootstrapping in k8s.
deploy-dev-local:
	./hack/k8s/spike-dev-install.sh

# Shell into SPIKE Pilot.
exec-spike:
	./hack/k8s/spike-sh.sh

tail-nexus:
	kubectl logs spike-nexus-0 -n spike -f

tail-keeper-0:
	kubectl logs spike-keeper-0 -n spike -f

tail-keeper-1:
	kubectl logs spike-keeper-1 -n spike -f

tail-keeper-2:
	kubectl logs spike-keeper-2 -n spike -f
