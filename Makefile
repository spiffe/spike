#   \\
#  \\\\ SPIKE: Secure your secrets with SPIFFE.
# \\\\\\

start:
	./hack/start.sh

demo-registry-entry:
	./examples/consume-secrets/demo-register-entry.sh

demo-create-policy:
	./examples/consume-secrets/demo-create-policy.sh

demo-put-secret:
	./examples/consume-secrets/demo-put-secret.sh