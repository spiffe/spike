#   \\
#  \\\\ SPIKE: Secure your secrets with SPIFFE.
# \\\\\\

# Start a local development environment to test SPIKE.
# By default, SPIRE Agent will have the same privileges as the user.
start:
	./hack/start.sh

# Start a local SPIKE development environment.
# In this case, SPIRE Agent will use privileged mode.
start-privileged:
	./hack/start.sh --use-sudo

# Builds SPIKE binaries.
build:
	./hack/bare-metal/build/build-spike.sh

# Registry an entry to the SPIRE server for the demo app.
demo-register-entry:
	./examples/consume-secrets/demo-register-entry.sh

# Create necessary access policies for the demo app.
demo-create-policy:
	./examples/consume-secrets/demo-create-policy.sh

# Put a sample secret to SPIKE Nexus for the demo app.
demo-put-secret:
	./examples/consume-secrets/demo-put-secret.sh

# 0. Prune docker file system to save resources.
docker-cleanup:
	./hack/docker/cleanup.sh

# 1. Reset the test cluster.
k8s-delete:
	./hack/k8s/minikube-delete.sh

# 2. Start the test cluster.
k8s-start:
	./hack/k8s/minikube-start.sh

# 3. Build container images.
docker-build:
	./hack/docker/build.sh

# 4. Forward registry.
docker-forward-registry:
	./hack/docker/minikube-forward-registry.sh

# 5. Push to the container registry.
docker-push:
	./hack/docker/push.sh

# 4. Deploy SPIRE to the test cluster.
deploy-spire:
	./hack/k8s/spire-install.sh

# 5. Deploy SPIKE
deploy-spike:
	./hack/k8s/spike-install.sh

tail-nexus:
	kubectl logs spike-nexus-0 -n spike-system -f

tail-keeper-0:
	kubectl logs spike-keeper-0 -n spike-system -f

tail-keeper-1:
	kubectl logs spike-keeper-1 -n spike-system -f

tail-keeper-2:
	kubectl logs spike-keeper-2 -n spike-system -f

exec-spike:
	./hack/k8s/spike-sh.sh

.PHONY: lint-go
lint-go:
	./hack/lint-go.sh
