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
	./hack/docker/build-local.sh

# 4. Forward registry.
docker-forward-registry:
	./hack/docker/minikube-forward-registry.sh

# 5. Push to the container registry.
docker-push:
	./hack/docker/push-local.sh

# 4. Deploy SPIRE and SPIKE to the cluster.
deploy-local:
	./hack/k8s/spike-install.sh

deploy-demo:
	./hack/k8s/spike-install-demo.sh

# Port forward the bundle endpoint for the demo.
spire-bundle-port-forward:
	./hack/k8s/spire-server-bundle-endpoint-port-forward.sh

tail-nexus:
	kubectl logs spike-nexus-0 -n spike -f

tail-keeper-0:
	kubectl logs spike-keeper-0 -n spike -f

tail-keeper-1:
	kubectl logs spike-keeper-1 -n spike -f

tail-keeper-2:
	kubectl logs spike-keeper-2 -n spike -f

# Shell into SPIKE Pilot.
exec-spike:
	./hack/k8s/spike-sh.sh
