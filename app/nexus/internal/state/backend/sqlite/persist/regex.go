//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"fmt"
	"regexp"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
)

// compileRegexPatterns compiles the SPIFFE ID and path patterns into regular
// expressions. Returns an error if either pattern is invalid.
func compileRegexPatterns(
	policy *data.Policy,
) error {
	var err error
	policy.IDRegex, err = regexp.Compile(policy.SPIFFEIDPattern)
	if err != nil {
		return fmt.Errorf("invalid spiffeid pattern: %w", err)
	}
	policy.PathRegex, err = regexp.Compile(policy.PathPattern)
	if err != nil {
		return fmt.Errorf("invalid path pattern: %w", err)
	}
	return nil
}
