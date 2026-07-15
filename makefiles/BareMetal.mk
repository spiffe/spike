#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

# Start a local development environment to test SPIKE.
# By default, SPIRE Agent will have the same privileges as the user.
#
# ! Make sure you read https://spike.ist/development/bare-metal/ and          !
# ! https://spike.ist/getting-started/quickstart/ before you run this command !
# ! to have a smooth developer experience.                                    !
start:
	./hack/bare-metal/startup/start.sh

# Kill any dangling SPIKE-related processes.
# Use this if `make start` was interrupted and left background processes running.
kill:
	./hack/bare-metal/startup/kill.sh

.PHONY: bootstrap
# Initialize the SPIKE setup with a brand new, secure, random root key.
bootstrap:
	./hack/bare-metal/startup/bootstrap.sh

# Start a local SPIKE development environment.
# In this case, SPIRE Agent will use privileged mode.
start-privileged:
	./hack/bare-metal/startup/start.sh --use-sudo

# Builds SPIKE binaries.
build:
	./hack/bare-metal/build/build-spike.sh

.PHONY: build-spire
# Builds and installs the SPIRE server and agent binaries that the
# bare-metal dev environment needs (SPIRE v1.11.2 into /usr/local/bin).
# Requires Go and sudo. Run this once before `make start` if
# spire-server/spire-agent are not already on your PATH.
build-spire:
	./hack/bare-metal/build/build-spire.sh

# Registry an entry to the SPIRE server for the demo app.
demo-register-entry:
	./examples/consume-secrets/demo-register-entry.sh

# Create necessary access policies for the demo app.
demo-create-policy:
	./examples/consume-secrets/demo-create-policy.sh

# Put a sample secret to SPIKE Nexus for the demo app.
demo-put-secret:
	./examples/consume-secrets/demo-put-secret.sh
