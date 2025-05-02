```text
  \\ 
 \\\\ SPIKE: Secure your secrets with SPIFFE.
\\\\\\
```

## SPIKE Nexus

**SPIKE Nexus** is the Secrets Store of SPIKE. It stores secrets securely in
memory and also can optionally back them up (*in encrypted form*) to a Postgres
database for production deployments.

The decryption of the secrets will be done by a **root key** that is automatically
generated during **SPIKE Nexus**'s bootstrapping sequence. This **root key**
is also securely share with **SPIKE Keeper** instances for redundancy and 
automatic recovery.

The administrator installing **SPIKE** for the first time is encouraged to take
an encrypted backup of the **root key** to manually recover the system if the
system locks itself.

If you are initializing **SPIKE** from the **SPIKE Pilot** CLI, you will be 
prompted to enter a password to back up the root key. It is **crucial** that 
you don't forget the password, and you don't lose the encrypted backup of
the **root key**.
