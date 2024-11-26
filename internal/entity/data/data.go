//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package data

import "time"

type InitState string

const AlreadyInitialized InitState = "AlreadyInitialized"
const NotInitialized InitState = "NotInitialized"

type Secret struct {
	Data map[string]string `json:"data"`
}

type PolicyPermission string

// PermissionRead gives permission to read secrets.
// This includes listing secrets.
const PermissionRead PolicyPermission = "read"

// PermissionWrite gives permission to write (including
// create, update and delete) secrets.
const PermissionWrite PolicyPermission = "write"

type Policy struct {
	Id              string             `json:"id"`
	Name            string             `json:"name"`
	SpiffeIdPattern string             `json:"spiffe_id_pattern"`
	PathPattern     string             `json:"path_pattern"`
	Permissions     []PolicyPermission `json:"permissions"`
	CreatedAt       time.Time          `json:"created_at"`
	CreatedBy       string             `json:"created_by"`
}
