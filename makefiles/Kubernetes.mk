#   \\
#  \\\\ SPIKE: Secure your secrets with SPIFFE.
# \\\\\\

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

# 4. Deploy SPIRE and SPIKE to the cluster.
deploy-spire:
	./hack/k8s/spire-install.sh

tail-nexus:
	kubectl logs spike-nexus-0 -n spike -f

tail-keeper-0:
	kubectl logs spike-keeper-0 -n spike -f

tail-keeper-1:
	kubectl logs spike-keeper-1 -n spike -f

tail-keeper-2:
	kubectl logs spike-keeper-2 -n spike -f

exec-spike:
	./hack/k8s/spike-sh.sh