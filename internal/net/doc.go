//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package net provides HTTP utilities for SPIKE components.
//
// This package contains shared networking code used by SPIKE Nexus, SPIKE
// Keeper, and other components for handling HTTP requests and responses.
//
// Request handling:
//
//   - HandleRoute: Wraps HTTP handlers with audit logging, generating trail
//     IDs, and recording request lifecycle events.
//   - Handler: Function type for request handlers with audit support.
//
// Response utilities:
//
//   - Respond: Writes JSON responses with proper headers and caching controls.
//   - Fail: Sends error responses with appropriate HTTP status codes.
//   - HandleError: Processes SDK errors and sends 404 or 500 responses.
//   - HandleInternalError: Sends 500 responses for internal errors.
//   - MarshalBodyAndRespondOnMarshalFail: Safely marshals JSON responses.
//
// HTTP client:
//
//   - Post: Performs HTTP POST requests with error handling and response
//     body management.
//
// Request parsing:
//
//   - ReadBody: Reads and unmarshals JSON request bodies.
//   - ValidateSpiffeId: Extracts and validates SPIFFE IDs from requests.
//
// This package is internal to SPIKE and should not be imported by external
// code.
package net
