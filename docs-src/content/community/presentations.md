+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.

title = "Presentations and Demos"
weight = 3
sort_by = "weight"
+++

# Presentations and Demos

Here you can find a range of presentations and demos that highlight **SPIKE**'s
capabilities and showcase its innovative features.

* [Developing SPIKE on Bare Metal and Kubernetes][spike-dev]:<br>
  This walkthrough demonstrates building and running **SPIKE** both on local
  bare-metal Linux and inside a local Minikube Kubernetes, illustrating how the
  project can be built, developed, and tested on your development environment.
* [Building and Testing SPIKE from Source in ~2 Minutes][spike-in-2]:<br>
  This is a quick demonstration of how to clone, build, and test the **SPIKE**
  system from its codebase in under two minutes, showing rapid developer
  iteration and validating that the core components work end-to-end.
* [Introduction to SPIKE: Secure Production Identity for Key
  Encryption][spike]:<br>
  This is our first **SPIKE** walkthrough, where we introduce the project and 
  its goals.
* [Unlocking SPIKE: A New Era for Secure Identity-Driven
  Secrets][spike-intro]:<br>
  This is a brief introduction to **SPIKE**; what it is, how it works, and why
  it's important.
* [Unlocking SPIKE: A Seamless Token-Based Login Experience][spike-jwt]:<br>
  In this demo, we see **SPIKE**'s new JWT authentication flow.
* [Goodbye Passwords: Secure Secrets Management with SPIFFE
  and SPIKE][spike-passwordless]:<br>
  In this demo, we allow a user to use `spike` just by checking an SVID; we 
  don't use any kind of passwords to identify the user, SPIFFE does it for us.
* [Unveiling SPIKE's New Audit Trail Capabilities: Zero-Trust Meets
  Accountability][spike-audits]:<br>
  This demo explains the new auditing capabilities of **SPIKE** that we will
  continue developing. This is the initial incarnation of the feature, and
  more will come.
* [Introducing Policy-Based Access Control in SPIKE][spike-policy]:<br>
  This demo introduces the new policy-based access control feature of **SPIKE**.
  This is the initial incarnation of the feature. We will create more demos
  as we enhance the feature.
* [Introducing SPIKE Secret Metadata API][spike-metadata-intro]<br>
  This demo introduces **SPIKE**'s new metadata API, which lets you attach
  key/value metadata to secrets to add richer context, governance, or
  classification. It also shows how the SPIKE Go SDK can be used to both set
  and retrieve that metadata in conjunction with policy enforcement.
* [Unlocking Secrets: Policy-Based Access and Metadata in 
  SPIKE][spike-policy-access]<br>
  This demo showcases how **SPIKE** enforces policy-based access control when
  reading or writing secrets, incorporating metadata as a first-class dimension
  of those policies. We observe how policies can conditionally govern operations
  based on the metadata values, enabling fine-grained, context-aware
  authorization.
* [Policy to the Rescue: Secure Secret Access and Metadata with
  SPIKE][spike-metadata]:<br>
  This demo introduces the new **SPIKE** metadata API. We also use the **SPIKE**
  Go SDK to consume secrets.
* [Using Policies to Read and Write Secrets using SPIKE][spike-policies]:<br>
  This demo introduces the new **Makefile**-based development workflow, the
  enhanced starter script, policy-based access control, and metadata support.
* [Secrets Resiliency with SPIKE: Self-Healing and Doomsday
  Recovery][spike-doomsday]:<br>
  Secrets management is critical, but what happens when everything fails? In 
  this video, we explore **SPIKE**’s disaster recovery mechanisms, covering both 
  self-healing capabilities and the manual break-the-glass recovery process.
* [Federating Secrets with SPIFFE and SPIKE][spike-federated]:<br>
  In this demo, we show how you can deploy **SPIRE** and **SPIKE** from SPIFFE
  Helm charts. We then establish a multi-cluster secret federation where 
  the workload clusters can securely access secrets stored in the management
  cluster.
* [SPIKE's Shamir's Secret Sharing with SPIFFE mTLS][spike-shamir]:<br>
  This demo walks through how SPIKE leverages Shamir's Secret Sharing to split
  the root key across multiple SPIKE Keeper nodes such that no single node holds 
  the full key. It also demonstrates how communications between **SPIKE
  Keeper**s and **SPIKE Nexus** are secured using SPIFFE-based mTLS to ensure
  authenticated, encrypted transport.
* [Secure SPIKE Deployment: Integrating SPIRE with an Isolated Management
  Cluster][secure-spike]:<br>
  In this demo, the presenter shows how to deploy SPIKE in a management cluster
  that is isolated from workload clusters, integrating SPIRE to issue identities
  and enforce trust boundaries. They highlight deployment topology, secure
  isolation methods, and how **SPIKE** components interface via SPIFFE
  identities in that setup.
* [Cross-Cluster Secrets Federation with SPIFFE and
  SPIKE][spike-fed-secrets]:<br>
  Here, the focus is on federating secrets across multiple clusters, allowing
  workloads in different clusters to access shared secrets securely. The demo
  shows how **SPIKE** can bridge trust boundaries using SPIFFE identities and
  secret federation.

[spike-shamir]: https://youtu.be/N2uAeFwxf90?si=CfZXPbQtWOKzE6Sd
[secure-spike]: https://youtu.be/BHtl_wGN-KY?si=pf1CZBf6NX4P5U5m
[spike-dev]: https://youtu.be/AdJblx6NLOU?si=y9mZ053mTLHNUQve
[spike-fed-secrets]: https://youtu.be/-AtHyqakbeY?si=eb16L9wb0LhonE_i
[spike-in-2]: https://youtu.be/Rl6pBvxffA0?si=dUkeBkB1yLxML5Yw
[spike]: https://youtu.be/Eeis67-3dd0?si=Z_vM1pOXhQG0ip-o
[spike-intro]: https://youtu.be/NEvQpTeKFp0?si=iuYx9xL_aA6SHECv
[spike-jwt]: https://youtu.be/ZT1f67N8vLA?si=k4a79C40-v3aqIj8
[spike-passwordless]: https://youtu.be/Tk8EERYjATo?si=JE8UR-F16nRE8rVs
[spike-audits]: https://youtu.be/EnIsDbQqUEs?si=WgqNXeUzBVPZdn7w
[spike-policy]: https://youtu.be/KGxHxgtHptI?si=0ljNrKKm0q138pcn
[spike-policy-access]: https://youtu.be/pyi26rIJbnI?si=ZZhGCNYhecc3TCQD
[spike-metadata]: https://youtu.be/OSr5VahEE0E?si=p_JV5IhtwmC8FA3S
[spike-metadata-intro]: https://youtu.be/OSr5VahEE0E?si=7Q4kfKdBU_2atwlC
[spike-policies]: https://youtu.be/cwNMHDzLP5Y?si=eFQcUlm212pOufBF
[spike-doomsday]: https://youtu.be/MX8dIUDC9iI?si=vGInHbBd3Vv0Iion
[spike-federated]: https://youtu.be/xGAg_zBvJrg?si=bEz2uJwQnalSOAMw

<p>&nbsp;</p>

----

{{ toc_community() }}

----

{{ toc_top() }}
