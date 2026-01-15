//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package route contains HTTP route handlers for SPIKE Nexus API endpoints.
//
// This package is organized into sub-packages by functional domain:
//
//   - base: Core routing logic, request dispatching, and guard utilities
//   - bootstrap: Bootstrap verification endpoints (proof-of-possession protocol)
//   - cipher: Encryption and decryption endpoints (encryption-as-a-service)
//   - operator: Disaster recovery endpoints (recover, restore)
//   - secret: Secret management endpoints (CRUD operations with versioning)
//   - acl/policy: Access control policy management endpoints
//
// # Authentication Model
//
// All routes require mTLS authentication via SPIFFE. The caller's SPIFFE ID is
// extracted from the client certificate and validated before any operations are
// performed. This provides strong cryptographic identity verification at the
// transport layer.
//
// # Authorization Model
//
// Routes fall into two authorization categories:
//
// 1. Policy-based routes (secret, cipher, acl/policy):
// These routes use CheckAccess to evaluate the caller's SPIFFE ID against
// defined policies. Workloads with matching policies can access resources
// based on their granted permissions (read, write, list, super).
//
// 2. Identity-restricted routes (bootstrap, operator):
// These routes require exact SPIFFE ID matches and do not honor policy
// overrides. Recovery and restore operations, for example, are restricted
// to specific SPIKE Pilot identities to prevent unauthorized access to
// sensitive key material.
//
// # Error Response Design
//
// Route handlers return distinct HTTP status codes for authentication failures
// (401 Unauthorized) versus input validation failures (400 Bad Request). While
// security purists might prefer uniform error responses to prevent information
// leakage, this asymmetry is acceptable in SPIKE's threat model for the
// following reasons:
//
//   - No enumeration attack surface: Unlike username/password authentication
//     where distinct errors help enumerate valid accounts, here the caller
//     already knows their own SPIFFE ID. They cannot probe for "valid"
//     identities since identity is cryptographically bound to their SVID.
//
//   - Binary authorization: Identity checks are "you are X or you are not."
//     An attacker with a non-matching SVID learns nothing useful from an
//     Unauthorized response since they already know their own identity.
//
//   - mTLS is the real gate: Attackers without a valid SVID from the trust
//     domain cannot even establish a connection. HTTP response codes are
//     only visible to entities that have already passed a strong
//     authentication boundary.
//
//   - Operational benefit: Distinct error codes help legitimate workloads
//     debug issues. "Unauthorized" indicates an identity mismatch while
//     "Bad Request" indicates malformed input. Conflating these would
//     trade marginal theoretical security for real operational pain.
//
// # Request Handling Patterns
//
// Most routes use the embedded guard pattern via net.ReadParseAndGuard, which
// combines request reading, JSON parsing, and authorization in a single
// operation. The cipher package uses a separate guard pattern due to its
// unique streaming mode requirements; see the cipher package documentation
// for details.
//
// # Interceptors
//
// Each route handler is typically paired with an interceptor file (e.g.,
// get.go and get_intercept.go). Interceptors contain guard functions that
// perform authentication and authorization checks before the main handler
// logic executes. This separation keeps security logic distinct from
// business logic.
package route
