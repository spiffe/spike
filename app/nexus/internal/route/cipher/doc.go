//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package cipher provides HTTP route handlers for encryption and decryption
// operations in SPIKE Nexus.
//
// This package implements encryption-as-a-service endpoints that allow
// workloads to encrypt and decrypt data using the Nexus cipher without
// persisting the data. It supports both JSON and streaming modes for
// efficient handling of different data types and sizes.
//
// # Request Handling Patterns
//
// This package uses a specific request handling pattern that differs from
// other SPIKE routes due to unique security and technical requirements.
//
// ## Two Approaches to Request Validation
//
// SPIKE routes generally use one of two patterns:
//
// 1. Embedded Guard Pattern (used in most routes):
//   - Read, parse, and guard in a single operation
//   - Example: net.ReadParseAndGuard combines all steps
//   - Simpler flow, fewer function calls
//   - Guard validates SPIFFE ID only, not request fields
//
// 2. Separate Guard Pattern (used in cipher routes):
//   - Extract SPIFFE ID separately
//   - Parse request data
//   - Call guard with complete request object
//   - Guard validates both SPIFFE ID and request fields
//
// ## Why Cipher Routes Use Separate Guard Pattern
//
// The cipher routes require the separate guard pattern for three reasons:
//
//  1. Request Field Validation Requirement:
//     The guard functions need to validate request fields (ciphertext size,
//     plaintext limits, etc.), not just perform authentication. This requires
//     passing the complete request object to the guard.
//
//  2. Streaming Mode Constraint:
//     Streaming mode has a chicken-and-egg problem: we need the cipher to
//     parse the request (to know the nonce size), but we must perform
//     authentication before accessing the cipher (principle of least
//     privilege). This requires splitting authentication from parsing.
//
//  3. Consistency:
//     Both streaming and JSON modes use the same pattern for consistency and
//     maintainability, even though JSON mode doesn't have the technical
//     constraint.
//
// ## Request Flow
//
// ### Streaming Mode Flow
//
//  1. Extract and validate SPIFFE ID (lightweight auth, no cipher access)
//  2. Get cipher (only after SPIFFE ID validation passes)
//  3. Read binary data (now that we have cipher for nonce size)
//  4. Construct request object from binary data
//  5. Call guard with request + SPIFFE ID (full validation)
//  6. Perform cryptographic operation
//  7. Send response
//
// ### JSON Mode Flow
//
//  1. Extract and validate SPIFFE ID (lightweight auth)
//  2. Parse JSON request (doesn't need cipher)
//  3. Call guard with request + SPIFFE ID (full validation)
//  4. Get cipher (only after validation passes)
//  5. Perform cryptographic operation
//  6. Send response
//
// ## Security Properties
//
// Both flows ensure:
//   - Cipher (sensitive resource) is accessed only after SPIFFE ID validation
//   - Guard receives complete request for field validation
//   - Principle of least privilege is maintained
//   - Authentication happens before authorization
//
// ## Function Organization
//
// Request handling is organized into focused functions:
//   - decrypt_intercept.go: Authentication and authorization guards
//   - encrypt_intercept.go: Authentication and authorization guards
//   - read.go: Request parsing (WithoutGuard variants)
//   - handle.go: Request orchestration (implements flows above)
//   - crypto.go: Cryptographic operations
//   - validation.go: Request field validation
//   - net.go: Response formatting
//
// ## Adding Request Validation
//
// To add request field validation, modify the guard functions:
//
//	func guardDecryptCipherRequest(
//	    request reqres.CipherDecryptRequest,
//	    peerSPIFFEID *spiffeid.ID,
//	    w http.ResponseWriter,
//	    r *http.Request,
//	) error {
//	    // Add validation here:
//	    if len(request.Ciphertext) > maxCiphertextSize {
//	        return net.Fail(...)
//	    }
//	    // ... rest of guard logic
//	}
package cipher
