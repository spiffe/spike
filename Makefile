#   \\
#  \\\\ SPIKE: Secure your secrets with SPIFFE.
# \\\\\\

# Start a local development environment to test SPIKE.
start:
	./hack/start.sh

# Builds SPIKE binaries.
build:
	./hack/build-spike.sh

# Registry an entry to the SPIRE server for the demo app.
demo-registry-entry:
	./examples/consume-secrets/demo-register-entry.sh

# Create necessary access policies for the demo app.
demo-create-policy:
	./examples/consume-secrets/demo-create-policy.sh

# Put a sample secret to SPIKE Nexus for the demo app.
demo-put-secret:
	./examples/consume-secrets/demo-put-secret.sh