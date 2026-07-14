+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

title = "Multi-Tenancy Deployment Patterns"
weight = 6
sort_by = "weight"
+++

# Multi-Tenancy Deployment Patterns

This guide describes strategies for deploying **SPIKE** in multi-tenant 
environments where multiple organizational units, teams, or customers need 
isolated secrets management.

## Understanding SPIKE's Single-Tenant Model

By default, SPIKE operates with a single-tenant model:

* **SPIKE Pilot** has a system SPIFFE ID that grants unrestricted access to all
  secrets and policies 
* The operator (*anyone with access to SPIKE Pilot*) has full visibility across
  the entire secret store
* Policies control workload access but do not restrict administrative access

This model works well when:

- A single team or organization manages all secrets
- The operator is trusted with visibility into all data
- There is no requirement for administrative isolation between groups

## When You Need Multi-Tenancy

Consider multi-tenancy when:

- **Organizational boundaries**: Different business units should not see each
  other's secrets
- **Customer isolation**: You are building a platform where customers expect
  data isolation
- **Compliance requirements**: Regulations mandate administrative separation
- **Blast radius reduction**: You want to limit the impact of credential
  compromise
- **Delegated administration**: Tenant admins should manage their own scope
  without central operator involvement

## Current Multi-Tenancy Options

SPIKE currently supports two approaches to multi-tenancy, each with different
trade-offs.

### Option 1: Separate SPIKE Deployments (*Strong Isolation*)

Deploy independent SPIKE instances for each tenant, typically in separate
Kubernetes namespaces or on separate infrastructure.

```
Tenant: PepsiCola                    Tenant: CocaCola
namespace: pepsi-secrets             namespace: coca-secrets
├── spike-nexus                      ├── spike-nexus
├── spike-keeper (x3)                ├── spike-keeper (x3)
├── spike-pilot                      ├── spike-pilot
└── spike-bootstrap                  └── spike-bootstrap
```

#### Kubernetes Example

```yaml
# pepsi-secrets namespace
apiVersion: v1
kind: Namespace
metadata:
  name: pepsi-secrets
  labels:
    tenant: pepsi
---
# Deploy SPIKE components in this namespace
# (Helm chart or manifests with namespace: pepsi-secrets)
```

```yaml
# coca-secrets namespace
apiVersion: v1
kind: Namespace
metadata:
  name: coca-secrets
  labels:
    tenant: coca
---
# Deploy SPIKE components in this namespace
# (Helm chart or manifests with namespace: coca-secrets)
```

#### SPIRE Configuration

Each tenant deployment requires its own SPIRE registration entries. You can
use the same SPIRE Server with different SPIFFE ID paths per tenant:

```bash
# Pepsi tenant SPIKE components
spire-server entry create \
  -spiffeID spiffe://example.org/tenant/pepsi/spike/nexus \
  -parentID spiffe://example.org/spire/agent/k8s/pepsi-secrets \
  -selector k8s:ns:pepsi-secrets \
  -selector k8s:sa:spike-nexus

spire-server entry create \
  -spiffeID spiffe://example.org/tenant/pepsi/spike/pilot/role/superuser \
  -parentID spiffe://example.org/spire/agent/k8s/pepsi-secrets \
  -selector k8s:ns:pepsi-secrets \
  -selector k8s:sa:spike-pilot

# Coca tenant SPIKE components
spire-server entry create \
  -spiffeID spiffe://example.org/tenant/coca/spike/nexus \
  -parentID spiffe://example.org/spire/agent/k8s/coca-secrets \
  -selector k8s:ns:coca-secrets \
  -selector k8s:sa:spike-nexus

spire-server entry create \
  -spiffeID spiffe://example.org/tenant/coca/spike/pilot/role/superuser \
  -parentID spiffe://example.org/spire/agent/k8s/coca-secrets \
  -selector k8s:ns:coca-secrets \
  -selector k8s:sa:spike-pilot
```

#### Environment Configuration

Each tenant's SPIKE components need their own trust root configuration:

```bash
# Pepsi tenant
export SPIKE_TRUST_ROOT_NEXUS="tenant/pepsi"
export SPIKE_TRUST_ROOT_PILOT="tenant/pepsi"
export SPIKE_TRUST_ROOT_KEEPER="tenant/pepsi"
export SPIKE_TRUST_ROOT_BOOTSTRAP="tenant/pepsi"

# Coca tenant
export SPIKE_TRUST_ROOT_NEXUS="tenant/coca"
export SPIKE_TRUST_ROOT_PILOT="tenant/coca"
export SPIKE_TRUST_ROOT_KEEPER="tenant/coca"
export SPIKE_TRUST_ROOT_BOOTSTRAP="tenant/coca"
```

#### Characteristics

| Aspect               | Description                         |
|----------------------|-------------------------------------|
| Isolation Level      | Strong (complete separation)        |
| Operational Overhead | High (N deployments to manage)      |
| Resource Usage       | Higher (duplicate components)       |
| Cross-Tenant Access  | Impossible by design                |
| Central Management   | None (each tenant is independent)   |
| Blast Radius         | Limited to single tenant            |

#### When to Use

* Tenants are external customers or competitors
* Regulatory requirements mandate complete isolation
* Tenants have different compliance requirements 
  (*e.g., one needs FIPS, another does not*)
* You need different availability or backup policies per tenant
* Maximum security is more important than operational efficiency

#### Best Practices

1. **Use GitOps**: Template your SPIKE deployment manifests to reduce
   duplication
2. **Centralize SPIRE Server**: You can use a single SPIRE Server with
   per-tenant SPIFFE ID paths
3. **Namespace isolation**: Use Kubernetes NetworkPolicies to prevent
   cross-namespace communication
4. **Separate storage**: Each SPIKE Nexus should have its own database or
   storage path
5. **Monitoring**: Aggregate metrics and logs centrally with tenant labels

### Option 2: Path-Based Tenant Isolation (*Soft Isolation*)

Use a single **SPIKE** deployment with path conventions and policies to isolate
tenant data. The central operator retains visibility across all tenants.

```
Single SPIKE Deployment
├── spike-nexus
├── spike-keeper (x3)
├── spike-pilot          <-- Central operator, sees all tenants
└── spike-bootstrap

Secret Paths:
├── tenants/pepsi/db/credentials
├── tenants/pepsi/api/keys
├── tenants/coca/db/credentials
└── tenants/coca/api/keys
```

#### Policy Configuration

Create policies that restrict each tenant's workloads to their own paths:

```bash
# Pepsi workloads can only access pepsi secrets
spike policy create --name=pepsi-workload-read \
  --path-pattern="^tenants/pepsi/.*$" \
  --spiffeid-pattern="^spiffe://example\.org/tenant/pepsi/workload/.*$" \
  --permissions="read"

spike policy create --name=pepsi-workload-write \
  --path-pattern="^tenants/pepsi/.*$" \
  --spiffeid-pattern="^spiffe://example\.org/tenant/pepsi/workload/.*$" \
  --permissions="write"

# Coca workloads can only access coca secrets
spike policy create --name=coca-workload-read \
  --path-pattern="^tenants/coca/.*$" \
  --spiffeid-pattern="^spiffe://example\.org/tenant/coca/workload/.*$" \
  --permissions="read"

spike policy create --name=coca-workload-write \
  --path-pattern="^tenants/coca/.*$" \
  --spiffeid-pattern="^spiffe://example\.org/tenant/coca/workload/.*$" \
  --permissions="write"
```

#### Tenant Admin Workloads

You can create "tenant admin" workloads that have broader permissions within
their tenant's scope:

```bash
# Pepsi admin workload can manage all pepsi secrets
spike policy create --name=pepsi-admin \
  --path-pattern="^tenants/pepsi/.*$" \
  --spiffeid-pattern="^spiffe://example\.org/tenant/pepsi/admin$" \
  --permissions="super"

# Coca admin workload can manage all coca secrets
spike policy create --name=coca-admin \
  --path-pattern="^tenants/coca/.*$" \
  --spiffeid-pattern="^spiffe://example\.org/tenant/coca/admin$" \
  --permissions="super"
```

**Important**: These tenant admin workloads:

- Can read/write/list secrets within their tenant's path
- Cannot see secrets outside their tenant's path
- Cannot create or modify policies (*that requires SPIKE Pilot access*)
- Cannot perform recovery/restore operations

#### Characteristics

| Aspect               | Description                                  |
|----------------------|----------------------------------------------|
| Isolation Level      | Weak (workloads isolated, operator sees all) |
| Operational Overhead | Low (single deployment)                      |
| Resource Usage       | Lower (shared components)                    |
| Cross-Tenant Access  | Prevented by policy (for workloads)          |
| Central Management   | Yes (single Pilot manages all)               |
| Blast Radius         | Full system if Pilot is compromised          |

#### When to Use

* Tenants are internal teams within the same organization
* A trusted central operator is acceptable
* You want simpler operations with a single deployment
* Workload isolation is sufficient (*admin visibility is acceptable*)
* You need cross-tenant visibility for auditing or compliance

#### Best Practices

1. **Consistent path conventions**: Establish and enforce a standard path
   structure (e.g., `tenants/{tenant-name}/{category}/{secret-name}`)
2. **Policy naming conventions**: Use clear names that indicate tenant scope
   (e.g., `{tenant}-{role}-{permission}`)
3. **Regular policy audits**: Review policies to ensure no accidental
   cross-tenant access
4. **SPIFFE ID conventions**: Structure workload SPIFFE IDs to include tenant
   information for easier policy matching
5. **Documentation**: Document the path structure and policy model for all
   tenant administrators

#### Limitations

- **No administrative isolation**: The central Pilot operator can see all
  tenants' secrets
- **Policy misconfiguration risk**: An incorrect regex pattern could leak
  secrets across tenants
- **No delegated policy management**: Tenants cannot create their own policies
- **Single point of trust**: Pilot credential compromise affects all tenants

## Comparison Matrix

| Capability                   | Separate Deployments    | Path-Based Isolation |
|------------------------------|-------------------------|----------------------|
| Workload isolation           | Yes                     | Yes                  |
| Administrative isolation     | Yes                     | No                   |
| Delegated policy management  | Yes (per-tenant Pilot)  | No                   |
| Cross-tenant visibility      | No                      | Yes (for operator)   |
| Resource efficiency          | Lower                   | Higher               |
| Operational complexity       | Higher                  | Lower                |
| Blast radius                 | Per-tenant              | Full system          |
| Compliance suitability       | Higher                  | Lower                |

## Future: Scoped Pilot Instances

A future enhancement 
**Scoped Pilot Instances** that would provide administrative isolation within
a single SPIKE deployment:

```
# Future capability (not yet implemented)
spiffe://example.org/spike/pilot/scope/tenants/pepsi
spiffe://example.org/spike/pilot/scope/tenants/coca
```

(see
[EPIC: Scoped SPIKE Pilot Instances for Multi-Tenancy](https://github.com/spiffe/spike/issues/281))

Scoped Pilots would:

- Only see secrets and policies within their designated scope
- Only create policies for paths within their scope
- Provide delegated administration without full deployment separation
- Maintain a single Nexus/Keeper infrastructure

This would offer a middle ground between the two current options, providing
administrative isolation with lower operational overhead than separate
deployments.

## Choosing an Approach

Use this decision tree to select the right approach:

```
Do tenants require COMPLETE administrative isolation?
|
+-- YES --> Are tenants external customers or competitors?
|           |
|           +-- YES --> Separate Deployments (Option 1)
|           |
|           +-- NO --> Wait for Scoped Pilots (ADR-0033)
|                      or use Separate Deployments
|
+-- NO --> Is a trusted central operator acceptable?
           |
           +-- YES --> Path-Based Isolation (Option 2)
           |
           +-- NO --> Separate Deployments (Option 1)
```

<p>&nbsp;</p>

----

{{ toc_operations() }}

----

{{ toc_top() }}
