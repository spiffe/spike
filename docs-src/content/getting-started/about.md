+++
# //    \\ SPIKE: Secure your secrets with SPIFFE.
# //  \\\\\ Copyright 2024-present SPIKE contributors.
# // \\\\\\\ SPDX-License-Identifier: Apache-2.

title = "About SPIKE"
weight = 1
sort_by = "weight"
+++

# A Brief Introduction to SPIKE

## About SPIKE

**SPIKE** is a lightweight secrets store that uses [**SPIFFE**][spiffe] as its
identity control plane.

Using **SPIFFE** as the identity layer ensures strong, verifiable workload
identities without relying on static credentials like API keys or passwords.
This enables **SPIKE** to authenticate and authorize workloads dynamically,
reducing the attack surface and preventing key sprawl. Additionally,
**SPIFFE**'s automatic identity rotation and short-lived cryptographic
identities **enhance security and resilience**, making it an ideal foundation
for a **zero-trust secrets management system**.

[spiffe]: https://spiffe.io/ "SPIFFE"

## Why Not Use Kubernetes Secrets

One of the most misunderstood aspects of Kubernetes secrets is that
**Kubernetes secrets are not truly secrets**. While Kubernetes provides a
mechanism to store and manage sensitive information, it is not a dedicated
secrets management solution.

Here’s why relying on Kubernetes secrets can be problematic:

### Limited Scope & Portability

If a service outside Kubernetes—or in another cluster—needs access to a
Kubernetes secret, it introduces significant authentication and authorization
challenges.

Kubernetes Secrets are designed to work within a single cluster, making it
difficult to share them securely across different environments
(*e.g., multiple clusters, bare-metal infrastructure, or cloud-based services*).
This limitation can lead to inconsistent security policies and fragmented secret
management practices.

A robust secrets management strategy should consider secrets' scope beyond a
single cluster.

### Static Nature of the Secrets

Kubernetes secrets are static once created. They are stored in `etcd` and
injected into pods at startup, meaning updates require manual
intervention—modifying the secret, redeploying affected applications, and
ensuring no stale or expired secrets remain in use.

Updating a Kubernetes Secret does not automatically notify or reload the
workloads consuming it. Applications typically need to be restarted or
re-deployed to pick up the new secret, adding operational complexity and
potential downtime if not managed carefully.

This lack of flexibility introduces security risks and operational overhead.

### Security and Governance Limitations

Kubernetes Secrets are governed by Kubernetes RBAC. Using Kubernetes Secrets, it
can be tricky to enforce a platform-agnostic security policy that spans
multiple environments. This often leads to a fragmented governance and
potential misconfigurations.

A dedicated secrets manager offers dynamic cross-environment compatibility,
and stronger security controls—making it a better choice for modern,
distributed architectures.

### Kubernetes Secrets Are Not Encrypted By Default

Kubernetes Secrets are stored in `etcd`, and unless encryption at rest is
explicitly enabled, they are stored in plaintext. This means that anyone with
access to etcd (*including certain privileged users or attackers who compromise
the cluster*) can retrieve sensitive data without needing Kubernetes API access.

Moreover, even with encryption at rest, the security model of Kubernetes Secrets
remains weaker than a dedicated secrets store. While encryption prevents direct
retrieval of plaintext secrets from `etcd` storage, an attacker with right
privileges can get the encryption key. Additionally, the Kubernetes API must
decrypt secrets when serving them to workloads, meaning any user or process
with sufficient API permissions can still retrieve secrets in plaintext.

So not only `etcd` itself, but also API-layer access is also a risk factor
in enforcing the security of Kubernetes Secrets.

### When Are Kubernetes Secrets Useful?

Despite these challenges, Kubernetes Secrets can still be useful in simple,
cluster-contained workloads where:

* Secrets do not need frequent rotation.
* All applications consuming the secrets reside in the same cluster.
* RBAC policies are well-configured to prevent accidental exposure.

However, for any multi-cluster, dynamic, or **zero-trust** architecture, a
dedicated secrets management solution is a better approach—providing
fine-grained access control, cross-environment compatibility, and stronger
security guarantees.

That part taken care of, [we can get our hands dirty with **SPIKE** in the
**SPIKE Quickstart Guide**](@/getting-started/guide.md).

----

{{ toc_getting_started() }}

----

{{ toc_top() }}