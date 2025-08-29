//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

// Spec represents the YAML structure for policy configuration
type Spec struct {
	Name        string   `yaml:"name"`
	SpiffeID    string   `yaml:"spiffeid"`
	Path        string   `yaml:"path"`
	Permissions []string `yaml:"permissions"`
}
