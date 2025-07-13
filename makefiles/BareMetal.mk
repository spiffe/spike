#   \\
#  \\\\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
# \\\\\\

# Start a local development environment to test SPIKE.
# By default, SPIRE Agent will have the same privileges as the user.
start:
	./hack/bare-metal/startup/start.sh

# Start a local SPIKE development environment.
# In this case, SPIRE Agent will use privileged mode.
start-privileged:
	./hack/bare-metal/startup/start.sh --use-sudo

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
