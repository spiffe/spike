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
	./hack/build-spike.sh

# Registry an entry to the SPIRE server for the demo app.
demo-register-entry:
	./examples/consume-secrets/demo-register-entry.sh

# Create necessary access policies for the demo app.
demo-create-policy:
	./examples/consume-secrets/demo-create-policy.sh

# Put a sample secret to SPIKE Nexus for the demo app.
demo-put-secret:
	./examples/consume-secrets/demo-put-secret.sh

# TODO:
# x. reset microk8s
# x. setup deps & metallb
# 3. deploy SPIRE from charts
# x. build dockerfiles
# x. push images to a local registry
# 6. deploy SPIKE from local registry + manifests
# 7. test if you can create/read secrets and policies
# 8. verify that all files in ./hack has a accompanying make target
#    and also they are references properly across code and docs.

# Reset docker.
docker-cleanup:
	./hack/docker/cleanup.sh

# Build container images.
docker-build:
	./hack/docker/build.sh

docker-push:
	./hack/docker/push.sh

k8s-reset:
	./hack/k8s/microk8s-reset.sh

deploy-spire:
	./hack/k8s/spire-install.sh

.PHONY: lint-go
lint-go:
	./hack/lint-go.sh
