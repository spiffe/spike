//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"context"
	"regexp"
	"testing"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/config/auth"

	"github.com/spiffe/spike/app/nexus/internal/state/persist"
)

func TestPatternIsAnchored(t *testing.T) {
	tests := []struct {
		pattern string
		want    bool
	}{
		{"^spike/system/acl$", true},
		{"^.*$", true},
		{"^spike/system/.*$", true},
		{"acl", false},
		{"^acl", false},
		{"acl$", false},
		{"", false},
		{"^a$|acl", false},
	}

	for _, tt := range tests {
		if got := patternIsAnchored(tt.pattern); got != tt.want {
			t.Errorf("patternIsAnchored(%q) = %v, want %v",
				tt.pattern, got, tt.want)
		}
	}
}

// A pattern must describe a reserved path, not merely appear inside it.
func TestPatternDescribesPath(t *testing.T) {
	const acl = "spike/system/acl"

	tests := []struct {
		pattern string
		want    bool
	}{
		// Deliberate: these genuinely describe the reserved path.
		{"^spike/system/acl$", true},
		{"spike/system/acl", true},
		{"^spike/system/.*$", true},
		{".*", true},
		{"^.*$", true},
		// Accidental: these reach it only through substring matching.
		{"acl", false},
		{"system", false},
		{"spike", false},
		{"^spike", false},
		{"acl$", false},
	}

	for _, tt := range tests {
		if got := patternDescribesPath(tt.pattern, acl); got != tt.want {
			t.Errorf("patternDescribesPath(%q, %q) = %v, want %v",
				tt.pattern, acl, got, tt.want)
		}
	}
}

// A deliberate catch-all identity is acceptable; a loose specific one is not.
func TestPatternIsPreciseIdentity(t *testing.T) {
	tests := []struct {
		pattern string
		want    bool
	}{
		{"^spiffe://example\\.org/admin$", true},
		{".*", true},
		{"^.*$", true},
		{"^spiffe://example\\.org/.*$", true},
		{"spiffe://example\\.org/admin", false},
		{"admin", false},
	}

	for _, tt := range tests {
		if got := patternIsPreciseIdentity(tt.pattern); got != tt.want {
			t.Errorf("patternIsPreciseIdentity(%q) = %v, want %v",
				tt.pattern, got, tt.want)
		}
	}
}

// A wildcard policy is unambiguous about covering everything, including the
// reserved namespaces, and must keep working.
func TestUpsertPolicy_AllowsDeliberateWildcard(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		persist.InitializeBackend(nil)

		policy := data.Policy{
			Name:            "wildcard-admin",
			SPIFFEIDPattern: ".*",
			PathPattern:     ".*",
			Permissions:     []data.PolicyPermission{data.PermissionSuper},
		}
		if _, err := UpsertPolicy(policy); err != nil {
			t.Fatalf("UpsertPolicy rejected a deliberate wildcard: %v", err)
			return
		}

		wants := []data.PolicyPermission{data.PermissionWrite}
		if !CheckPolicyAccess(
			"spiffe://example.org/anything",
			auth.PathSystemPolicyAccess, wants,
		) {
			t.Error("wildcard policy stopped granting reserved-path access")
		}
	})
}

func TestIsReservedSystemPath(t *testing.T) {
	reserved := []string{
		auth.PathSystemPolicyAccess,
		auth.PathSystemSecretAccess,
		auth.PathSystemCipherExecute,
	}
	for _, path := range reserved {
		if !isReservedSystemPath(path) {
			t.Errorf("isReservedSystemPath(%q) = false, want true", path)
		}
	}

	ordinary := []string{"secrets/db", "spike/system", "tenants/acme/db"}
	for _, path := range ordinary {
		if isReservedSystemPath(path) {
			t.Errorf("isReservedSystemPath(%q) = true, want false", path)
		}
	}
}

// UpsertPolicy must refuse a policy whose unanchored path pattern reaches a
// reserved system namespace by substring.
func TestUpsertPolicy_RejectsUnanchoredReservedPath(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		persist.InitializeBackend(nil)

		// Each of these reaches a reserved path by substring alone.
		patterns := []string{"acl", "system", "spike", "secret", "exec"}

		for _, pattern := range patterns {
			_, err := UpsertPolicy(data.Policy{
				Name:            "escalation-" + pattern,
				SPIFFEIDPattern: "^spiffe://example\\.org/audit$",
				PathPattern:     pattern,
				Permissions:     []data.PolicyPermission{data.PermissionWrite},
			})
			if err == nil {
				t.Errorf("UpsertPolicy accepted unanchored reserved-path "+
					"pattern %q", pattern)
			}
		}
	})
}

// An unanchored SPIFFE ID pattern is equally unacceptable for a reserved
// path, even when the path pattern itself is anchored correctly.
func TestUpsertPolicy_RejectsUnanchoredIDForReservedPath(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		persist.InitializeBackend(nil)

		_, err := UpsertPolicy(data.Policy{
			Name:            "acl-manager-loose-id",
			SPIFFEIDPattern: "spiffe://example\\.org/admin",
			PathPattern:     "^spike/system/acl$",
			Permissions:     []data.PolicyPermission{data.PermissionWrite},
		})
		if err == nil {
			t.Error("UpsertPolicy accepted an unanchored SPIFFE ID pattern " +
				"for a reserved system path")
		}
	})
}

// A deliberate, fully anchored delegation of policy management is still
// allowed: the guard constrains accidents, not intent.
func TestUpsertPolicy_AllowsAnchoredReservedPath(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		persist.InitializeBackend(nil)

		policy := data.Policy{
			Name:            "acl-manager",
			SPIFFEIDPattern: "^spiffe://example\\.org/admin$",
			PathPattern:     "^spike/system/acl$",
			Permissions:     []data.PolicyPermission{data.PermissionWrite},
		}
		if _, err := UpsertPolicy(policy); err != nil {
			t.Fatalf("UpsertPolicy rejected an anchored delegation: %v", err)
			return
		}

		wants := []data.PolicyPermission{data.PermissionWrite}
		if !CheckPolicyAccess(
			"spiffe://example.org/admin", auth.PathSystemPolicyAccess, wants,
		) {
			t.Error("anchored delegation did not grant policy write")
		}
		// The anchored identity pattern must not admit a lookalike.
		if CheckPolicyAccess(
			"spiffe://example.org/admin-attacker",
			auth.PathSystemPolicyAccess, wants,
		) {
			t.Error("anchored delegation admitted a lookalike identity")
		}
	})
}

// Ordinary paths keep plain, unanchored regular expression semantics. This
// is intended behavior and the guard must not alter it.
func TestUpsertPolicy_OrdinaryPathsUnaffected(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		persist.InitializeBackend(nil)

		policy := data.Policy{
			Name:            "prefix-reader",
			SPIFFEIDPattern: "^spiffe://example\\.org/web$",
			PathPattern:     "^tenants/acme/",
			Permissions:     []data.PolicyPermission{data.PermissionRead},
		}
		if _, err := UpsertPolicy(policy); err != nil {
			t.Fatalf("UpsertPolicy rejected an ordinary prefix policy: %v",
				err)
			return
		}

		wants := []data.PolicyPermission{data.PermissionRead}
		if !CheckPolicyAccess(
			"spiffe://example.org/web", "tenants/acme/db/creds", wants,
		) {
			t.Error("prefix policy stopped matching; anchoring was applied " +
				"to an ordinary path")
		}
	})
}

// CheckPolicyAccess must independently refuse a policy that predates the
// guard. StorePolicy is used directly to bypass UpsertPolicy's rejection,
// simulating a policy already in the backing store.
func TestCheckPolicyAccess_RefusesPreexistingUnanchoredPolicy(t *testing.T) {
	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		persist.InitializeBackend(nil)

		ctx := context.Background()
		legacy := data.Policy{
			Name:            "legacy-acl",
			SPIFFEIDPattern: "spiffe://example\\.org/audit",
			PathPattern:     "acl",
			Permissions:     []data.PolicyPermission{data.PermissionWrite},
		}
		legacy.IDRegex = regexp.MustCompile(legacy.SPIFFEIDPattern)
		legacy.PathRegex = regexp.MustCompile(legacy.PathPattern)

		if storeErr := persist.Backend().StorePolicy(ctx, legacy); storeErr != nil {
			t.Fatalf("StorePolicy: %v", storeErr)
			return
		}

		wants := []data.PolicyPermission{data.PermissionWrite}
		if CheckPolicyAccess(
			"spiffe://example.org/audit", auth.PathSystemPolicyAccess, wants,
		) {
			t.Error("a pre-existing unanchored policy escalated to policy " +
				"write on the reserved ACL path")
		}
	})
}

// The guard must survive the SQLite load path, which recompiles both
// patterns from the stored strings on every access check.
func TestCheckPolicyAccess_ReservedGuardHoldsUnderSQLite(t *testing.T) {
	withSQLiteEnvironment(t, func() {
		ctx := context.Background()
		cleanupSQLiteDatabase(t)

		rootKey := createTestRootKey(t)
		resetRootKey()
		Initialize(rootKey)

		defer func() {
			_ = persist.Backend().Close(ctx)
		}()

		// Stored directly so it reaches the database despite the guard in
		// UpsertPolicy, as a policy written before the guard existed would.
		legacy := data.Policy{
			Name:            "legacy-acl",
			SPIFFEIDPattern: "spiffe://example\\.org/audit",
			PathPattern:     "acl",
			Permissions:     []data.PolicyPermission{data.PermissionWrite},
		}
		if storeErr := persist.Backend().StorePolicy(ctx, legacy); storeErr != nil {
			t.Fatalf("StorePolicy: %v", storeErr)
			return
		}

		wants := []data.PolicyPermission{data.PermissionWrite}
		if CheckPolicyAccess(
			"spiffe://example.org/audit", auth.PathSystemPolicyAccess, wants,
		) {
			t.Error("SQLite recompile path escalated to policy write on " +
				"the reserved ACL path")
		}
	})
}
