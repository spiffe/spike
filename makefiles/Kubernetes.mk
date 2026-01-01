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
#
# ! Make sure you read https://spike.ist/development/quickstart/ and          !
# ! https://spike.ist/getting-started/bare-metal/ before you run this command !
# ! to have a smooth developer experience.                                    !
k8s-start:
	./hack/k8s/minikube-start.sh

# Deletes and re-installs the Minikube cluster.
k8s-reset:
	k8s-delete
	k8s-start

# 3. Build container images.
docker-build:
	./hack/docker/build-local.sh

# For Minikube, instead of forwarding the registry, you can directly load
# the container images to the cluster's internal local registry.
#
# This is especially helpful when you are using Docker Desktop for Windows' WSL
# Integration and `make docker-push` hangs up regardless of the
# `insecure-registries` settings in `Settings > Docker Engine` of Docker
# for Windows or `/etc/docker/daemon.json` of Docker on WSL.
#
# This can happen because in a typical WSL-Docker-for-Windows integration,
# your Docker CLI is in WSL, but the daemon is Docker Desktop: So when you run
# `docker push localhost:5000/...`, the Windows-side daemon tries to reach
# Windows' `localhost:5000`; meanwhile your WSL `docker push` will hit WSL's
# `localhost:5000` (*where you likely did the `kubectl port-forward`). Those
# are different network stacks. The result will: the push sits on "Waiting".
k8s-load-images:
	./hack/k8s/minikube-load-images.sh

# For Multi-Cluster Federation Demo, DO NOT run `deploy-local`
# Instead, see FederationDemo.mk for the remaining steps.

# 6. Deploy SPIKE locally.
deploy-dev-local:
	./hack/k8s/spike-dev-install.sh
	#./hack/k8s/spike-job-install.sh

# Shell into SPIKE Pilot.
exec-spike:
	./hack/k8s/spike-sh.sh

# Builds and deploys SPIKE to Minikube from scratch. This target orchestrates
# a complete deployment pipeline: builds binaries, cleans up Docker resources,
# builds container images, resets the Minikube cluster, loads images into the
# cluster, and deploys SPIKE in development mode.
deploy-minikube: \
	build \
	docker-cleanup \
	docker-build \
	k8s-delete \
	k8s-start \
	k8s-load-images \
	deploy-dev-local

tail-nexus:
	kubectl logs spike-nexus-0 -n spike -f

tail-keeper-0:
	kubectl logs spike-keeper-0 -n spike -f

tail-keeper-1:
	kubectl logs spike-keeper-1 -n spike -f

tail-keeper-2:
	kubectl logs spike-keeper-2 -n spike -f
