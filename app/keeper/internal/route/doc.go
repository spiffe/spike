//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package route contains HTTP route handlers for SPIKE Keeper API endpoints.
//
// SPIKE Keeper is a shard storage service that holds Shamir secret shares for
// disaster recovery. Unlike SPIKE Nexus, which manages secrets and policies,
// Keeper has a focused responsibility: securely storing and serving shards
// that can be combined to reconstruct the root key.
//
// This package is organized into sub-packages:
//
//   - base: Core routing logic and request dispatching
//   - store: Shard storage endpoints (contribute, retrieve)
//
// # Endpoints
//
// Keeper exposes two endpoints:
//
//   - Contribute: Accepts shard contributions from SPIKE Bootstrap (during
//     initial setup) or SPIKE Nexus (during periodic updates). Validates
//     that the shard is non-nil and non-zero before storing.
//
//   - Shard: Returns the stored shard to SPIKE Nexus during recovery
//     operations. Only Nexus is authorized to retrieve shards.
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
// Unlike SPIKE Nexus, Keeper does not use policy-based authorization. All
// routes are identity-restricted, requiring exact SPIFFE ID matches:
//
//   - Contribute: Accepts requests from SPIKE Bootstrap or SPIKE Nexus only
//   - Shard: Accepts requests from SPIKE Nexus only
//
// This strict identity-based model is appropriate because Keeper handles
// sensitive key material that should never be accessible to arbitrary
// workloads, regardless of any policy configuration.
//
// # Error Response Design
//
// Route handlers return distinct HTTP status codes for authentication failures
// (401 Unauthorized) versus input validation failures (400 Bad Request). This
// follows the same rationale as SPIKE Nexus routes: the asymmetry is acceptable
// because mTLS + SPIFFE ID verification forms a strong authentication boundary,
// and distinct error codes aid operational debugging without meaningful
// information leakage.
//
// See the SPIKE Nexus route package documentation for the full security
// analysis of this design decision.
//
// # Security Properties
//
// Keeper implements additional security measures for handling sensitive shard
// data:
//
//   - Memory clearing: Shard data is zeroed from memory after use to minimize
//     exposure window
//   - Input validation: Shards are validated to be non-nil and non-zero before
//     storage
//   - Audit logging: All operations are logged for security auditing
//
// # Interceptors
//
// Each route handler is paired with an interceptor file (e.g., shard.go and
// shard_intercept.go). Interceptors contain guard functions that perform
// authentication checks before the main handler logic executes.
package route
