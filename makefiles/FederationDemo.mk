#   \\
#  \\\\ SPIKE: Secure your secrets with SPIFFE.
# \\\\\\

# 6. Deploy SPIRE and SPIKE (based on the hostname)
demo-deploy:
	./examples/federation/spike-install-demo.sh

# 7. Port forward the bundle endpoint for the demo.
demo-spire-bundle-port-forward:
	./examples/federations/spire-server-bundle-endpoint-port-forward.sh

# 8. Extract bundles.
demo-bundle-extract:
	./examples/federation/extract-bundle.sh

# 9. Exchange bundles.
demo-bundle-set:
	./examples/federation/set-bundle.sh

# Extract and set in a single step.
demo-bundle-exchange:
	demo-bundle-extract
	demo-bundle-set

# 10. Port forward SPIKE Keeper instances for the demo setup.
demo-spike-keeper-port-forward:
	./examples/federation/spike-keeper-port-forward.sh

# 11. Port forward SPIKE Nexus for the workload to consume it.
demo-spike-nexus-port-forward:
	./examples/federation/spike-nexus-port-forward.sh

# 12. Deploy sample workload
demo-deploy-workload:
	./examples/federation/workload-deploy.sh

# 13. Set policies for the workload to consume secrets.
demo-set-policies:
	./examples/federation/workload-set-policies.sh

# 15. Check whether workload received secrets.
demo-exec:
	./examples/federation/workload-exec.sh