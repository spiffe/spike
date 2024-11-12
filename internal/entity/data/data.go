//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package data

type InitState string

var AlreadyInitialized InitState = "AlreadyInitialized"
var NotInitialized InitState = "NotInitialized"
