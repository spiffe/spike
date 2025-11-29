//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"regexp"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

// compileRegexPatterns compiles the SPIFFE ID and path patterns from the
// policy into regular expressions, storing them in the policy's IDRegex and
// PathRegex fields. This function modifies the policy in place.
//
// Parameters:
//   - policy: The policy containing SPIFFEIDPattern and PathPattern strings
//     to compile. The compiled regexes are stored in the IDRegex and
//     PathRegex fields.
//
// Returns:
//   - *sdkErrors.SDKError: nil on success, or ErrEntityInvalid if either
//     pattern fails to compile as a valid regular expression
func compileRegexPatterns(
	policy *data.Policy,
) *sdkErrors.SDKError {
	var err error

	policy.IDRegex, err = regexp.Compile(policy.SPIFFEIDPattern)
	if err != nil {
		failErr := sdkErrors.ErrEntityInvalid.Wrap(err)
		failErr.Msg = "invalid SPIFFE ID pattern " + policy.SPIFFEIDPattern
		return failErr
	}

	policy.PathRegex, err = regexp.Compile(policy.PathPattern)
	if err != nil {
		failErr := sdkErrors.ErrEntityInvalid.Wrap(err)
		failErr.Msg = "invalid path pattern " + policy.PathPattern
		return failErr
	}

	return nil
}
