#   \\
#  \\\\ SPIKE: Secure your secrets with SPIFFE.
# \\\\\\

# 6. Patch charts
demo-patch-charts:
	./hack/k8s/patch-charts.sh

# 7. Deploy SPIRE and SPIKE (based on the hostname)
demo-deploy:
	./hack/k8s/spike-install-demo.sh

# 8. Port forward the bundle endpoint for the demo.
demo-spire-bundle-port-forward:
	./hack/k8s/spire-server-bundle-endpoint-port-forward.sh

# 9. Extract bundles.
demo-bundle-extract:
	./hack/spiffe/extract-bundle.sh

# 10. Exchange bundles.
demo-bundle-set:
	./hack/spiffe/set-bundle.sh

# 11. Port forward SPIKE Keeper instances for the demo setup.
demo-spike-keeper-port-forward:
	./hack/k8s/spike-keeper-port-forward.sh
