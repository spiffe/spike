//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"regexp"
	"strings"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/config/auth"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

// arbitraryProbe is a value no real SPIFFE ID or secret path would equal. It
// is used to tell a deliberate catch-all pattern (which accepts it) from a
// specific pattern that only reaches a reserved path through substring
// matching (which does not).
const arbitraryProbe = "spike-arbitrary-probe-value"

// reservedSystemPaths are the internal namespaces that gate SPIKE's own
// privileged operations. Reaching spike/system/acl with write permission
// confers control over every policy in the system, so these paths are held
// to a stricter matching rule than ordinary secret paths.
var reservedSystemPaths = []string{
	auth.PathSystemPolicyAccess,
	auth.PathSystemSecretAccess,
	auth.PathSystemCipherExecute,
}

// isReservedSystemPath reports whether path is one of SPIKE's internal
// privileged namespaces.
func isReservedSystemPath(path string) bool {
	for _, reserved := range reservedSystemPaths {
		if path == reserved {
			return true
		}
	}
	return false
}

// asFullMatch compiles pattern so that it must match an entire subject rather
// than any substring of it. It returns nil when pattern does not compile.
func asFullMatch(pattern string) *regexp.Regexp {
	re, err := regexp.Compile("^(?:" + pattern + ")$")
	if err != nil {
		return nil
	}
	return re
}

// patternIsAnchored reports whether pattern is anchored at both ends.
func patternIsAnchored(pattern string) bool {
	return strings.HasPrefix(pattern, "^") && strings.HasSuffix(pattern, "$")
}

// patternDescribesPath reports whether pattern actually describes path, as
// opposed to merely appearing somewhere inside it.
//
// This is the distinction that matters for reserved namespaces. The pattern
// "acl" matches spike/system/acl only because Go regular expressions match on
// substrings; read as a description of a path it means the path "acl" and
// nothing else. By contrast ".*" and "^spike/system/.*$" genuinely describe
// the reserved path, and their authors plainly meant to include it.
func patternDescribesPath(pattern, path string) bool {
	full := asFullMatch(pattern)
	return full != nil && full.MatchString(path)
}

// patternIsPreciseIdentity reports whether pattern picks out identities
// precisely enough to be trusted with a reserved system path.
//
// An unanchored identity pattern is a substring matcher, so a policy written
// for "spiffe://example\.org/admin" also admits
// "spiffe://example.org/admin-attacker". An anchored pattern cannot do that.
// A deliberate catch-all such as ".*" is also accepted: it is unambiguous
// about covering every identity, and narrowing it is the operator's call.
func patternIsPreciseIdentity(pattern string) bool {
	if patternIsAnchored(pattern) {
		return true
	}
	full := asFullMatch(pattern)
	return full != nil && full.MatchString(arbitraryProbe)
}

// policyMayReachReservedPath reports whether policy is permitted to grant
// access to the given reserved system path.
func policyMayReachReservedPath(policy data.Policy, reserved string) bool {
	return patternDescribesPath(policy.PathPattern, reserved) &&
		patternIsPreciseIdentity(policy.SPIFFEIDPattern)
}

// reservedPathReachedBy returns the first reserved system path that re
// matches, and whether any was matched.
func reservedPathReachedBy(re *regexp.Regexp) (string, bool) {
	if re == nil {
		return "", false
	}
	for _, reserved := range reservedSystemPaths {
		if re.MatchString(reserved) {
			return reserved, true
		}
	}
	return "", false
}

// guardReservedSystemPaths rejects a policy that would reach one of SPIKE's
// reserved system namespaces without describing it deliberately.
//
// Ordinary secret paths are unaffected: they keep plain, unanchored regular
// expression semantics, where matching on a substring is intended behavior.
//
// Parameters:
//   - policy: The policy to check. Its PathRegex must already be compiled.
//
// Returns:
//   - *sdkErrors.SDKError: ErrEntityInvalid naming the reserved path reached
//     and what is required to reach it, or nil when the policy is acceptable.
func guardReservedSystemPaths(policy data.Policy) *sdkErrors.SDKError {
	reserved, reaches := reservedPathReachedBy(policy.PathRegex)
	if !reaches {
		return nil
	}

	if policyMayReachReservedPath(policy, reserved) {
		return nil
	}

	guardErr := sdkErrors.ErrEntityInvalid.Clone()
	guardErr.Msg = "policy " + policy.Name + " reaches the reserved system " +
		"path " + reserved + " only by substring match; anchor both patterns " +
		"with ^ and $ to grant access there deliberately (path pattern: " +
		policy.PathPattern + ", SPIFFE ID pattern: " +
		policy.SPIFFEIDPattern + ")"
	return guardErr
}
