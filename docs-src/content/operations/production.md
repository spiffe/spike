+++
# //    \\ SPIKE: Secure your secrets with SPIFFE.
# //  \\\\\ Copyright 2024-present SPIKE contributors.
# // \\\\\\\ SPDX-License-Identifier: Apache-2.

title = "SPIKE Production Setup"
weight = 2
sort_by = "weight"
+++

# SPIKE Production Setup Guide

This guide involves configuring the necessary environment, deploying the
application with optimized settings, and ensuring scalability, reliability, and
security for a seamless production experience.

## Baseline recommendations

### Do Not Run as Root

For **SPIKE** components use an unprivileged service account, rather than 
running as the root or Administrator account. **SPIKE** is designed to run as 
an unprivileged user, and doing so adds significant defense against various 
privilege-escalation attacks.

### Allow Minimal Write Privilege

**SPIKE Nexus** only needs writes access to its backing store. It's a good
practice to limit what is writable by the **SPIKE Nexus** process to just
the directories and files of the backing store.

### Disable Swap

**SPIKE** encrypts data in transit and at rest; however, it must still have 
sensitive data in memory to function. Risk of exposure should be minimized by 
disabling swap to prevent the operating system from paging sensitive data to 
disk. 

### Disable Core Dumps

A user or administrator that can force a core dump and has access to the 
resulting file can potentially access **SPIKE**'s root key and other 
cryptographically sensitive material encryption keys. Preventing core dumps is 
a platform-specific process; on Linux setting the resource limit `RLIMIT_CORE` 
to `0` disables core dumps. In the `systemd` service unit file, setting 
`LimitCORE=0` will enforce this setting for the Vault service.

### Network Security

Although **SPIKE** relies on Zero Trust networking principles and establishes
mTLS everywhere, that does not mean perimeter defense is unimportant.

Use a local firewall for **SPIRE Server**, **SPIKE Nexus**, and **SPIKE Keeper**
instances, or relevant features of your cloud provider to restrict incoming and 
outgoing traffic to the bare minimum that you need.

### Disable Shell Command History

You may want the `spike` commands themselves not appear in history at all.

### Keep a Frequent Upgrade Cadence

**SPIKE** is actively developed, hardened, and patched against vulnerabilities.
You should upgrade **SPIKE** often to incorporate security fixes and any 
changes in default settings such as key lengths or cipher suites. 

### Restrict Backing Store Access

**SPIKE** encrypts data at rest, regardless of the kind of backing store it
uses. Although **SPIKE** encrypts the data, an attacker with arbitrary 
control can cause data corruption or loss by modifying or deleting keys. 

You should restrict storage access outside **SPIKE Nexus** to avoid 
unauthorized access or operations.

Also, when using an external data store, although **SPIKE** assumes the store is 
untrusted, yet, still, considering the following is important:

* If this is a shared database with other services, who else has access to it
  and manages it?
* How will **SPIKE** authenticate to the database?
* Does the database connection allow TLS-protected secure communication?

### Configure SELinux / AppArmor

Using mechanisms like `SELinux` and `AppArmor` can help you gain layers of 
security when using **SPIKE**. While **SPIKE** can run on several popular 
operating systems, Linux is recommended due to the various security primitives
and memory governance.

### Container Considerations

**SPIKE** uses memory locking when possible. To use memory locking (`mlock`) 
inside a **SPIKE** container, you need to use the `overlayfs2` or another 
supporting driver.

### Logging Considerations

Like all systems, logging is an essential part of **SPIKE**. However, logs 
produced by **SPIKE** components also function as evidence for audits and 
security incidents.

Currently, we don't separate audit logs from event logs. Audit logs are clearly 
identified by the prefix `[AUDIT]:` at the beginning of each entry.

> **Future Goals**
>
> We have action items to separate audit logs from regular logs and redirect 
> them to a configurable list of audit targets. For now, they remain part of 
> the standard output stream of the application.

Since logs may serve as evidence, consider these important factors when 
implementing a logging solution:

* Retention periods should comply with your organization's legal requirements
* The logging system should maintain high availability for both log intake and 
  storage
* Logs should be tamper-proof with verifiable integrity
* The system should maintain and document a proper chain of custody

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

### Isolate SPIRE Server

The **SPIRE Server** can run completely in Kubernetes, alongside other pods and
applications.

However, it is a good security practice to run the **SPIRE Server** on a 
separate dedicated Kubernetes cluster, or on standalone hardware. This way, if 
the primary cluster is compromised, the SPIRE private keys are not at risk.

To protect **SPIRE** private keys even further, you can use one of the
supported [SPIRE KMS plugins][spire-docs].

[spire-docs]: https://github.com/spiffe/spire/tree/main/doc "SPIRE Docs on GitHub"

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

* `./hack/bare-metal/entry/spire-server-entry-recover-register.sh`
* `./hack/bare-metal/entry/spire-server-entry-spike-register.sh`
* `./hack/bare-metal/entry/spire-server-entry-restore-register.sh`

### Harden SPIRE Deployment on Kubernetes

If you have deployed **SPIRE** on Kubernetes:

* Use Kubernetes [**Pod Security Standards**][pod-security], Network Policies, 
  and RBAC to restrict **SPIRE Server** and **SPIRE Agent** access.
* Limit SPIRE components to trusted namespaces and nodes.

[pod-security]: https://kubernetes.io/docs/concepts/security/pod-security-standards/ "Pod Security Standards"

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
the **SPIKE Pilot superadmin**, will not be able to create SPIRE Server
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

The guidelines covered in this section apply to all **SPIKE** components, 
including **SPIKE Nexus**, **SPIKE Keeper**, and **SPIKE Pilot**.

### Single Tenancy

**SPIKE Nexus** is recommended to be the only main process running on a machine.
This reduces the risk that another process running on the same
machine is compromised and can interact with **SPIKE Nexus**.

In a Kubernetes deployment, you can achieve this with setting up appropriate 
Node affinity rules.

### User Privileges

* For **bare-metal** deployments:
  * Run **SPIKE Nexus** and **SPIKE Keeper** processes as **non-root** users.
  * Configure them to have minimal permissions.
  * Keep OS and security packages up to date.
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
    security context---**Do not** run the container as root.
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

* Set up a comprehensive process logging mechanism
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
* Have comprehensive security documentation.

## How the Root Key Is Protected in SPIKE

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

Let me complete this text for you:

## Turn Swap and Core Dumps Off

Both **SPIKE Nexus** and **SPIKE Keeper** maintain sensitive cryptographic 
material of varying degrees of sensitivity in memory. 

Although **SPIKE** uses secure memory erasing and memory locking practices to
as a defense mechanism against memory-based attacks, it's a good practice to
establish defense-in-depth practices, especially when an exposed root key
provides the possibility to reveal encrypted secrets.

If the memory is swapped, an attacker could potentially extract this 
cryptographic key material from the swap file on the disk. This would compromise 
the security of the system, as swap files are stored unencrypted on disk and 
may persist even after the system is powered down.

Similarly, core dumps can contain a complete copy of the process memory at the 
time of a crash, including any cryptographic keys, passwords, or other 
sensitive data that was in memory. An attacker with access to these core dump 
files could analyze them to extract the sensitive information.

Although **SPIKE** considers the machine as the trust boundary and assumes the
system is breached if the machine is breached, it does not mean we should relax 
security if the machine is compromised. Defense in depth is still important, and
minimizing the exposure of sensitive cryptographic material provides additional 
layers of protection against sophisticated attacks.

To mitigate these risks:

1. Disable swap entirely on systems handling sensitive cryptographic operations
2. If swap cannot be disabled, configure an encrypted swap
3. Disable core dumps for security-critical applications
4. Ensure proper permissions on any diagnostic files that might be generated
5. Consider using memory allocation techniques that minimize exposure of 
   sensitive data

These precautions help prevent attacks where adversaries might attempt to 
retrieve cryptographic keys or other sensitive information from persistent 
storage after it has been paged out from memory or dumped during a crash.

## Hardening SPIKE Keeper for Production

**SPIKE Keeper**s play a critical role in managing sensitive cryptographic
material, specifically handling **shards** that are used to generate the
**root key** that **SPIKE Nexus** uses to encrypt its backing store.

As described in the [**SPIKE Security Model**][security], protecting your system
against memory analysis is important, not only for **SPIKE**, but for any 
application you may be running in your system.

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

> **Memory Security of SPIKE Keepers**
> 
> Note that these memory protection measures are also applicable for 
> **SPIKE Keeper**s where we secure shards of the root key. While a single
> shard does not expose as much risk as an exposed root key, it's still 
> good defense in depth to securing the memory of **SPIKE Keeper** instances.

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
  * Configure memory limits based on an expected load
  * Set CPU quotas to prevent resource exhaustion
  * Implement disk I/O limits
* Monitor resource usage:
  * Track memory utilization
  * Monitor CPU usage
  * Alert on resource threshold violations

### Access Control

* Implement the least privilege access:
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

Although **SPIKE** is designed with security best-practices in mind, a 
multi-layer approach focusing on system, process, and network security is 
important when configuring **SPIKE** for production.

The combination of **mTLS API protection**, **SPIFFE attestation**, and proper
**system-level security controls** will provide robust protection against
unauthorized access to sensitive cryptographic material.

Remember that **security is an ongoing process**, and every system's security
posture and requirements are different. Thus, these measures outlined in this
guide shall be taken as starting recommendations and adjusted to meet your
organization's security requirements.

----

{{ toc_operations() }}

----

{{ toc_top() }}
