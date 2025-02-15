+++
# //    \\ SPIKE: Secure your secrets with SPIFFE.
# //  \\\\\ Copyright 2024-present SPIKE contributors.
# // \\\\\\\ SPDX-License-Identifier: Apache-2.

title = "SPIKE Production Setup"
weight = 2
sort_by = "weight"
+++

{{ star() }}

## Production Setup Guide

This guide involves configuring the necessary environment, deploying the
application with optimized settings, and ensuring scalability, reliability, and
security for a seamless production experience.

## Hardening SPIRE

**SPIKE** leverages **SPIFFE** and **SPIRE** as its identity control plane to
manage cryptographic workload identities securely and efficiently. **SPIRE**
is an implementation of the **SPIFFE** specification, providing a robust
framework for workload attestation and cryptographic identity issuance within
distributed systems.

Configuring **SPIRE** for production is critical to ensure the security and
reliability of **SPIKE**. An improperly configured SPIRE deployment can leave
gaps in the identity management process, potentially exposing sensitive
cryptographic operations to unauthorized access.

Here are some key steps to harden **SPIRE** for production:

### Secure SPIRE Server and Agent Communication

* Use **mutual TLS (mTLS)** for all communication between SPIRE Server, SPIRE
  Agents, and workloads.
* Configure **SPIRE** Server and Agents to only accept connections from trusted
  sources.

### Set Up Attestation Policies

* Define strict attestation policies to ensure that only trusted workloads are
  issued SPIFFE IDs. // TODO: explain what that means in the context of SPIKE.
* Utilize the **node attestation plugins** (e.g., AWS IID, Kubernetes) to verify
  the identity of nodes running SPIRE Agents.

### Limit Permissions

* Run SPIRE Server and Agents with the minimum required permissions.
* Use dedicated **non-root** users for running SPIRE processes.

### Secure SPIRE Database

If the SPIRE Server is configured to use an external database for
persistence, ensure that the database is:
* Encrypted at rest and in transit.
* Restricted to access only from SPIRE Server.

### Configure Registration Entries

* Create granular SPIFFE ID **registration entries** for specific workloads.
* Avoid using wildcard matching in selectors to reduce the risk of impersonation
  attacks.

You can find sample scripts that creates registration entries under the 
[`./hack`][hack] folder:

[hack]: https://github.com/spiffe/spike/tree/main/hack

* `./hack/spire-server-entry-recover-register.sh`
* `./hack/spire-server-entry-spike-register.sh`
* `./hack/spire-server-entry-restore-register.sh`

### Harden SPIRE Deployment on Kubernetes

If you have deployed **SPIRE** on Kubernetes:

* Use Kubernetes **Pod Security Policies**, Network Policies, and RBAC to
  restrict SPIRE Server and Agent access.
* Limit SPIRE components to trusted namespaces and nodes.

### Regularly Rotate Certificates

* Configure SPIRE to rotate workload certificates and keys frequently.
* Automate the process to ensure timely certificate renewal without manual
  intervention.

### Enable Logging and Monitoring

* Configure logging for SPIRE Server and Agents to capture suspicious
  activity.
* Monitor logs for failed authentication attempts, unauthorized access, or
  other anomalies.

### Perform Regular Audits

* Conduct regular security audits and penetration tests on the SPIRE deployment.
* Review registration entries and attestation policies to ensure they align
  with security best practices.

### Update SPIRE Regularly

* Keep SPIRE updated to the latest stable version to benefit from security
  patches and new features.

By carefully configuring and hardening **SPIRE**, you ensure that
**SPIKE**'s **SPIFFE-based identity control plane** is robust, reliable, and
secured against potential threats, forming the foundation for **SPIKE's** secure
operations in production environments.

### Isolate SPIRE Server

You are encouraged to isolate the **SPIRE Server** from other SPIKE components.

By doing this, a separate administrator can access the **SPIRE Server** and
create **SPIKE registration entries**, whereas other **SPIKE** users, including
the **SPIKE Pilot superadmin**, will not have the ability to create SPIRE Server
registration entries.

This approach aligns with **zero-trust**best practices by enforcing separation
of privileges and reducing the risk of privilege misuse or escalation.

For **bare-metal** or **VM deployments**, it is recommended to run the
**SPIRE Server** on its own dedicated machine, separate from **SPIKE Keeper**
and **SPIKE Nexus** machines. This ensures that any compromise of those
machines does not directly impact the **SPIRE Server**.

For **Kubernetes deployments**, it is a good practice to run the **SPIRE
Server** outside the Kubernetes cluster on an ultra-hardened system. This
protects the **SPIRE Server** from potential security breaches or privilege
escalations within the Kubernetes cluster.

## SPIKE General Hardening Guidelines

The guidelines covered in this section apply too all **SPIKE** components
including **SPIKE Nexus**, **SPIKE Keeper**, and **SPIKE Pilot**.

### User Privileges

* For **bare-metal** deployments:
  * Run **SPIKE Nexus** and **SPIKE Keeper** processes as **non-root** users.
  * Configure them to have minimal permissions.
  * Keep OS and security packages up-to-date.
* For **Kubernetes** deployments:
  * Disable privilege escalation for containers by setting
    `allowPrivilegeEscalation: false` in your PodSecurity configuration.
  * Use Kubernetes **Pod Security Admission** or equivalent policies to enforce
    security constraints.
  * Limit the use of privileged containers (`privileged: false`) wherever
    feasible.
  * Configure strict NetworkPolicies to restrict communication between Pods.
  * Always use read-only root filesystems for the containers
    (`readOnlyRootFilesystem: true`).
  * Specify **non-root** `runAsUser` and `runAsGroup` in the container's 
    security context.
* For **Docker** deployments:
  * Prevent containers from running in privileged mode using the
    `--privileged=false` option.
  * Use `--read-only` to enforce read-only filesystem access for the container.
  * Limit container capabilities by setting the `--cap-drop` option to drop all
    unnecessary capabilities.
  * Avoid mapping the Docker socket into containers for security-sensitive
    workloads.
  * Implement user namespaces with `--userns-remap` to isolate containers from
    the host's root user.

### Security Modules

* For Linux bare-metal **SPIKE** installations, consider enabling and
  configuring **AppArmor** and **SELinux**.
* Set up mandatory access control.
* Enforce strict process isolation.

### Network Security

* Restrict network access to essential ports/protocols.
* Implement network segmentation.
* Configure strict firewall rules.
* Conduct regular network security audits.

### Logging and Monitoring

* Set up comprehensive process logging
* Monitor for unauthorized access attempts
* Implement real-time alerting
* Regular log analysis and review

### Security Auditing

* Regular system configuration audits
* Security control effectiveness reviews
* Periodic penetration testing
* Configuration compliance checks

### Binary Integrity

Official **SPIKE** binaries are published with SHA-256 checksums. Make sure
you implement SHA hash verification when using **SPIKE** distributions to
ensure that you are using original, tested, validated, and approved binaries.

In addition, it's useful to have regular binary integrity checks too, to ensure
that binaries are not replaced with malicious code.

One more thing you are encouraged to do is to include **SPIKE Nexus**,
**SPIKE Keeper**, and **SPIKE Pilot**'s binary SHA hashes while registering
them to **SPIRE Server**. Here's an example:

```bash 
# Register SPIKE Keeper
spire-server entry create \
    -spiffeID spiffe://spike.ist/spike/keeper \
    -parentID "spiffe://spike.ist/spire-agent" \
    -selector unix:uid:"$KEEPER_UID" \
    -selector unix:path:"$KEEPER_PATH" \
    -selector unix:sha256:"$KEEPER_SHA"
```

This way, if the binary changes, **SPIRE Server** will not assign it an SVID,
and the rest of the system will not trust it and stop communicating with it,
effectively securing the **SPIKE** components by totally isolating and
keeping out the untrusted binary.

### Defense in Depth

* Implement multiple layers of security controls.
* Have regular security control reviews.
* Have a comprehensive security documentation.

## How th Root Key Is Protected in SPIKE

In **SPIKE**, the **root key** is essential for encrypting secrets within the 
central store, **SPIKE Nexus**. To prevent any single entity from having full 
access to this key, SPIKE uses [Shamir's Secret Sharing][shamir] to divide the 
root key into multiple shares. These shares are distributed among 
**SPIKE Keeper**s, ensuring that the root key can only be reconstructed when a 
sufficient number of shares are combined.

This approach enhances security by requiring collaboration among multiple
trusted components to access the root key.

[Shamir's Secret Sharing (SSS)][shamir] is a cryptographic method that divides a
secret into parts, distributing them among participants. The secret can only be
reconstructed when a minimum number of parts (the **threshold**) are combined. 
This ensures that partial knowledge of the secret does not compromise its 
security.

[shamir]: https://en.wikipedia.org/wiki/Shamir%27s_secret_sharing "Shamir's Secret Sharing"

## Hardening SPIKE Keeper for Production

**SPIKE Keeper**s play a critical role in managing sensitive cryptographic
material, specifically handling **shards** that are use to generate the
**root key** that **SPIKE Nexus** uses to encrypt its backing store.

As described in the [**SPIKE Security Model**][security], protecting your system
against memory analysis is important, not only for **SPIKE**, but for any 
application you may be running n your system.

System administrators should implement the following security measures to
prevent memory analysis:

* Set `/proc/sys/kernel/yama/ptrace_scope` to `2` or `3`:
    * Value `2` restricts `ptrace` to `root-only` access
    * Value `3` disables `ptrace` completely, offering maximum security
* Make this setting permanent by adding `kernel.yama.ptrace_scope = 2` to
  `/etc/sysctl.d/10-ptrace.conf`
* Consider using **SELinux** or **AppArmor** profiles to further restrict
  process debugging capabilities
* If running in a container, ensure the container runtime is configured to
  disable ptrace capabilities (*e.g.,
  using `--security-opt=no-new-privileges` in Docker*)
* Regular audit of processes with `CAP_SYS_PTRACE` capability, as this can
  bypass ptrace restrictions

[security]: @/architecture/security-model.md "SPIKE security model"

By default, **SPIKE Keeper**s are protected by multiple layers of security:

* **mTLS API Protection**: All **SPIKE Keeper** APIs are protected by mutual
  TLS (mTLS), preventing direct access to the shards through the API interface.
* **SPIFFE Attestation**: **SPIKE Keeper**s implement **SPIFFE attestation**
  which verifies the authenticity of **SPIKE Keeper** binaries by validating
  attributes like the SHA hash, unix user id, and path. This prevents attackers
  from running malicious keeper processes, as they would fail the attestation
  check.

Although these protections are in place, they need to be properly configured to
take effect. For example, a misconfigured **SPIRE Server** registration entry or
using a user with elevated privileges to run the **SPIKE Keeper** binaries may
result in a security breach (see the "*hardening SPIRE for production*" section
before for details)

## Hardening SPIKE Nexus for Production

**SPIKE Nexus** serves as the central secrets store, maintaining sensitive data 
in memory and using encrypted storage for persistence. Due to its critical role 
in managing secrets, special attention must be paid to its security 
configuration.

### Memory Protection

* The `ptrace` and `yama` recommendations for **SPIKE Keeper**s covered in
  the previous section also applies to **SPIKE Nexus**. Protect **SPIKE 
  Nexus**'s memory against external analysis.
* Configure memory restrictions to prevent swapping:
  * Set `vm.swappiness=0` in sysctl configuration
  * Use `mlock` to lock memory pages and prevent them from being swapped
  * If using systemd, set `LimitMEMLOCK=infinity` in the service file
* Enable Address Space Layout Randomization (ASLR):
  * Ensure `/proc/sys/kernel/randomize_va_space` is set to 2
* Implement memory scrubbing:
  * Configure automatic memory wiping for deallocated memory
  * Use secure memory allocation practices for sensitive data

### Backing Store Security

* Configure secure backup procedures:
  * Encrypt all backups
  * Implement strict access controls on backup storage
  * Regular backup integrity verification
* Monitor backing store access:
  * Log all access attempts
  * Implement alerting for unusual access patterns
  * Regular audit of access logs

### Resource Management

* Set appropriate resource limits:
  * Configure memory limits based on expected load
  * Set CPU quotas to prevent resource exhaustion
  * Implement disk I/O limits
* Monitor resource usage:
  * Track memory utilization
  * Monitor CPU usage
  * Alert on resource threshold violations

### Access Control

* Implement least privilege access:
  * Create dedicated service accounts
  * Restrict file system permissions
  * Use **SELinux** or **AppArmor** profiles

### Disaster Recovery

* Document recovery procedures:
  * Clear steps for various failure scenarios
  * Regular testing of recovery procedures
  * Maintain updated recovery documentation
* Configure backup systems:
  * Regular backup testing
  * Secure offsite storage
  * Automated recovery validation

### Container-Specific Hardening

When deploying SPIKE Nexus in containers:

* Use minimal base images:
  * Build from scratch or distroless images
    * Regular security updates
* Configure container security:
  * Enable `seccomp` profiles
  * Set appropriate `ulimit`s
  * Implement container isolation

Remember to regularly review and update these security measures based on new 
threats and security best practices. Security configuration should be treated 
as a continuous process rather than a one-time setup.

## Conclusion

Although **SPIKE** is designed security best-practices in mind, a multi-layer
approach focusing on system, process, and network security is important when
configuring **SPIKE** for production.

The combination of **mTLS API protection**, **SPIFFE attestation**, and proper
**system-level security controls** will provide robust protection against
unauthorized access to sensitive cryptographic material.

Remember that **security is an ongoing process**, and every system's security
posture and requirements is different. Thus, these measures outlined in this
guide shall be taken as starting recommendations and adjusted to meet your
organization's security requirements.

----

{{ toc_operations() }}

----

{{ toc_top() }}
