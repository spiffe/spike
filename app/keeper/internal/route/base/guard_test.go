//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestAllRouteHandlersUseGuardFunctions scans every route package for
// handlers that reach no guard. Handlers are recognized by signature and
// delegates are followed, so there is no list of exempt files, names, or
// directories to grow.
func TestAllRouteHandlersUseGuardFunctions(t *testing.T) {
	routerDir, absErr := filepath.Abs(".")
	if absErr != nil {
		t.Fatalf("failed to resolve the router directory: %v", absErr)
	}
	routeDir := filepath.Dir(routerDir)

	violations := scanRouteTree(t, routeDir, routerDir)
	if len(violations) > 0 {
		t.Errorf(
			"Route handlers missing guard function invocation:\n%s\n\n"+
				"All route handlers must either:\n"+
				"  1. Call net.ReadParseAndGuard with a guard function, or\n"+
				"  2. Call a guard function (e.g., guardXxxRequest) directly,\n"+
				"  3. Or delegate to a same-package function that does.\n\n"+
				"This ensures authorization checks are performed for every "+
				"request.",
			strings.Join(violations, "\n"),
		)
	}
}

// TestGuardScan_FlagsUnguardedHandler proves the scan reports a handler
// that reaches no guard, whatever its name or file.
func TestGuardScan_FlagsUnguardedHandler(t *testing.T) {
	dir := t.TempDir()
	writeScanFixture(t, dir, "anything.go", `package fixture

func Serve(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) *sdkErrors.SDKError {
	return nil
}
`)

	violations := checkDirectory(t, dir)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d: %v",
			len(violations), violations)
	}
	if !strings.HasSuffix(violations[0], "Serve") {
		t.Errorf("violation does not name the handler: %q", violations[0])
	}
}

// TestGuardScan_FollowsDelegates proves that a handler which hands the
// request to a same-package function, by call or by value, is guarded
// when that function reaches a guard.
func TestGuardScan_FollowsDelegates(t *testing.T) {
	dir := t.TempDir()
	writeScanFixture(t, dir, "route.go", `package fixture

func RouteByValue(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) *sdkErrors.SDKError {
	return net.DispatchByContentType(w, r, audit, handleJSON, handleStream)
}

func RouteByCall(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) *sdkErrors.SDKError {
	return handleJSON(w, r, audit)
}
`)
	writeScanFixture(t, dir, "handle.go", `package fixture

func handleJSON(w, r, audit any) *sdkErrors.SDKError {
	return guardRequest(w, r, audit)
}

func handleStream(w, r, audit any) *sdkErrors.SDKError {
	return handleJSON(w, r, audit)
}
`)

	if violations := checkDirectory(t, dir); len(violations) != 0 {
		t.Errorf("expected no violations, got %v", violations)
	}
}

// TestGuardScan_SeesHandlers fails when the scan finds no route handlers
// at all, which would mean the signature match or the walk is broken and
// TestAllRouteHandlersUseGuardFunctions is passing vacuously.
func TestGuardScan_SeesHandlers(t *testing.T) {
	routerDir, absErr := filepath.Abs(".")
	if absErr != nil {
		t.Fatalf("failed to resolve the router directory: %v", absErr)
	}

	if n := countRouteHandlers(t, filepath.Dir(routerDir), routerDir); n == 0 {
		t.Fatal("the guard scan found no route handlers; " +
			"the signature match or the walk is broken")
	}
}
