## Production Setup Guide

> **WORK IN PROGRESS**
> 
> Note that this is a work in progress.
>
> We will add more SPIKE production deployment best practices
> in time. 
> 
> Also, based on the `TODO: //` remarks in the text, expect things to change
> **a lot**.
> 
> We will remove this notice once the document stabilizes.


This guide involves configuring the necessary environment, deploying the
application with optimized settings, and ensuring scalability, reliability, and
security for a seamless production experience.

// TODO: re-edit this document after #74 is done, as the security posture will
// change significantly:
// https://github.com/spiffe/spike/issues/74

// TODO: some of these are generic recommendations and some of them are more
// SPIKE-specific; organize them into a coherent document before final publishing.

// TODO: once completed let Claude, ChatGPT, and Perplexity review this entire shebang

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

// TODO: add spike-specific recommendations.

// TODO: link to the official docs.

// TODO: carry over some SPIRE best practices from VSecM docs too.

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

// TODO: give examples in the context of SPIKE.

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
  * Specify **non-root** `runAsUser` and `runAsGroup` in the container's security
    context.
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

In SPIKE, the root key is essential for encrypting secrets within the central
store, SPIKE Nexus. To prevent any single entity from having full access to this
key, SPIKE uses Shamir's Secret Sharing to divide the root key into multiple
shares. These shares are distributed among SPIKE Keepers, ensuring that the root
key can only be reconstructed when a sufficient number of shares are combined.

This approach enhances security by requiring collaboration among multiple
trusted components to access the root key.

Shamir's Secret Sharing (SSS): SSS is a cryptographic method that divides a
secret into parts, distributing them among participants. The secret can only be
reconstructed when a minimum number of parts (the threshold) are combined. This
ensures that partial knowledge of the secret does not compromise its security.

## Hardening SPIKE Keeper for Production

**SPIKE Keeper**s play a critical role in managing sensitive cryptographic 
material, specifically handling **shards** that are use to generate the
**root key** that **SPIKE Nexus** uses to encrypt its backing store.

**SPIKE Keeper**s only temporarily generate the root key during their
bootstrapping process, and securely erase it from the memory as soon as it
is no longer needed. Therefore, the possible window of attack for obtaining
the root key is extremely slim.

Attaching to the memory of a running process is generally restricted due 
to security measures in Linux systems. By default, the Linux kernel's Yama 
security module enforces a policy that prevents non-root users from attaching 
debuggers to processes they do not own. This is controlled by the 
`/proc/sys/kernel/yama/ptrace_scope` setting. Which means, for any 
root-key extraction scenario, the attacker will also need to find a way to
elevate privileges.

Although the window of attack is extremely slim, even if an attacker finds a 
chance to attack through that window, it is still very hard to obtain the root 
key even during this short period. An attacker would need to
monster-in-the-middle (MITM) the mTLS connection or have root privileges 
to peek into **SPIKE Keeper**'s memory space. Such an attack vector is extremely
unlikely when **SPIKE Keeper** machines are properly secured---with root access
disabled, privilege escalation prevented, and SSH access restricted to a 
trusted set of IPs and users.

In addition it's worth noting that the attacker cannot use their own counterfeit
SPIKE Keeper binary because SPIFFE Attestation will reject it. So a successful
attack that changes the SPIKE Keeper's binary will also require the attacker
to hack SPIRE Server, which raises the barrier even higher. 

By default, **SPIKE Keeper**s are protected by multiple layers of security:

1. **mTLS API Protection**: All **SPIKE Keeper** APIs are protected by mutual 
   TLS (mTLS), preventing direct access to the shards through the API interface.
2. **SPIFFE Attestation**: **SPIKE Keeper**s implement **SPIFFE attestation** 
   which verifies the authenticity of **SPIKE Keeper** binaries by validating 
  attributes like the SHA hash, unix user id, and path. This prevents attackers 
  from running malicious keeper processes, as they would fail the attestation 
  check.
3. **Memory Access Restrictions**: The only theoretical way to access the root 
  key is through direct memory access, which is heavily restricted by OS-level 
  security controls when properly configured.

Although these protections are in place, they need to be properly configured to
take effect. For example, a misconfigured **SPIRE Server** registration entry or
using a user with elevated privileges to run the **SPIKE Keeper** binaries may
result in a security breach (see the "*hardening SPIRE for production*" section
before for details)

## Hardening SPIKE Nexus for Production

TBD

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

