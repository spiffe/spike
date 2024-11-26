//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package data

import (
	"regexp"
	"time"
)

type PolicyPermission string

// PermissionRead gives permission to read secrets.
// This DOES NOT include listing secrets.
const PermissionRead PolicyPermission = "read"

// PermissionWrite gives permission to write (including
// create, update and delete) secrets.
const PermissionWrite PolicyPermission = "write"

const PermissionList PolicyPermission = "list"

// PermissionSuper gives superuser permissions.
// The user is the alpha and the omega.
const PermissionSuper PolicyPermission = "super"

type Policy struct {
	Id              string             `json:"id"`
	Name            string             `json:"name"`
	SpiffeIdPattern string             `json:"spiffe_id_pattern"`
	PathPattern     string             `json:"path_pattern"`
	Permissions     []PolicyPermission `json:"permissions"`
	CreatedAt       time.Time          `json:"created_at"`
	CreatedBy       string             `json:"created_by"`

	// Unexported fields won't be serialized to JSON
	IdRegex   *regexp.Regexp `json:"-"`
	PathRegex *regexp.Regexp `json:"-"`
}
