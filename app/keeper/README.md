```text
  \\ 
 \\\\ SPIKE: Secure your secrets with SPIFFE.
\\\\\\
```

## SPIKE Keeper

**SPIKE Keeper** caches the **SPIKE Nexus**' root encryption key. This is useful 
when you need to recover state without requiring manual intervention if 
**SPIKE Nexus** crashes.

In cases where both **SPIKE Nexus** and **SPIKE Keeper** crash, an admin will 
need to manually re-key the system. To reduce the possibility of this, multiple 
**SPIKE Keeper** instances can be installed as a federated mesh for redundancy.