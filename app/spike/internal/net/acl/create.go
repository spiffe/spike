//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package acl

import "github.com/spiffe/go-spiffe/v2/workloadapi"

func CreatePolicy(source *workloadapi.X509Source,
	name string, pattern string, pattern2 string, permissions []string,
) error {
	return nil
}
