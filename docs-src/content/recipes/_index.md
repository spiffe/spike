+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

title = "Recipes"
weight = 4
sort_by = "weight"
+++

# Recipes

Task-first guides for SPIKE. Each recipe states the **problem** you're trying
to solve, gives a **TL;DR**, walks the **workflow** step by step, then lists
**tips**, **pitfalls**, and where to go **next**. For exhaustive option lists
see [Configuration](/usage/configuration/) and the
[command reference](/usage/commands/); for the why, see
[Architecture](/architecture/system-overview/).

## Concepts and Decisions

- [Choosing a backend store: memory, lite, or sqlite](/recipes/choosing-a-backend-store/)
- [Bootstrapping a fresh SPIKE](/recipes/bootstrapping-spike/)
- [Where the root key lives: keepers, Shamir, and recovery](/recipes/root-key-keepers-recovery/)

## Day-to-Day Usage

- [Storing and reading secrets](/recipes/storing-and-reading-secrets/)
- [Writing access policies](/recipes/writing-access-policies/)
- [Granting a workload access to secrets](/recipes/granting-a-workload-access/)
- [Using SPIKE as an encryption service](/recipes/encryption-as-a-service/)

## Operations and Lifecycle

- [Break-the-glass disaster recovery](/recipes/break-the-glass-recovery/)
- [Backup and restore](/recipes/backup-and-restore/)
- [Deploying SPIKE (Kubernetes and bare-metal)](/recipes/deploying-spike/)
- [Production hardening](/recipes/production-hardening/)
- [Troubleshooting](/recipes/troubleshooting/)

## Integration and Advanced

- [Integrating the Go SDK](/recipes/go-sdk-integration/)
- [Upgrading SPIKE](/recipes/upgrading-spike/)
