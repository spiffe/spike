//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAllRouteHandlersUseGuardFunctions verifies that all route handler
// functions invoke guard functions for authorization checks.
//
// The test scans all Go files in the route subdirectories (excluding _test.go,
// doc.go, and _intercept.go files) looking for functions that match the route
// handler signature pattern (functions starting with "Route").
//
// For each route handler found, it verifies that one of the following patterns
// is used:
//  1. net.ReadParseAndGuard - the standard pattern for JSON routes
//  2. A function call containing "guard" (case-insensitive) - for routes with
//     custom guard patterns (e.g., cipher routes that support streaming)
//
// If any route handler is missing guard invocation, the test fails with a
// descriptive error message.
func TestAllRouteHandlersUseGuardFunctions(t *testing.T) {
	relPath := filepath.Join("..", "..", "route")
	routeDir, err := filepath.Abs(relPath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Subdirectories containing route handlers
	subDirs := []string{
		"acl/policy",
		"bootstrap",
		"cipher",
		"operator",
		"secret",
	}

	var violations []string

	for _, subDir := range subDirs {
		dir := filepath.Join(routeDir, subDir)
		violations = append(violations, checkDirectory(t, dir)...)
	}

	if len(violations) > 0 {
		t.Errorf(
			"Route handlers missing guard function invocation:\n%s\n\n"+
				"All route handlers must either:\n"+
				"  1. Call net.ReadParseAndGuard with a guard function, or\n"+
				"  2. Call a guard function (e.g., guardXxxRequest) directly\n\n"+
				"This ensures authorization checks are performed for every "+
				"request.",
			strings.Join(violations, "\n"),
		)
	}
}

// checkDirectory scans a directory for route handler files and checks each
// handler for ReadParseAndGuard usage.
func checkDirectory(t *testing.T, dir string) []string {
	t.Helper()

	var violations []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Logf("Warning: could not read directory %s: %v", dir, err)
		return violations
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Skip non-Go files
		if !strings.HasSuffix(name, ".go") {
			continue
		}

		// Skip test files, doc files, and intercept files
		if strings.HasSuffix(name, "_test.go") ||
			name == "doc.go" ||
			strings.HasSuffix(name, "_intercept.go") {
			continue
		}

		// Skip helper/utility files that don't contain route handlers
		if isUtilityFile(name) {
			continue
		}

		filePath := filepath.Join(dir, name)
		fileViolations := checkFile(t, filePath)
		violations = append(violations, fileViolations...)
	}

	return violations
}

// isUtilityFile returns true if the file is a utility/helper file that
// doesn't contain route handlers.
func isUtilityFile(name string) bool {
	utilityFiles := []string{
		"errors.go",
		"guard.go",
		"map.go",
		"config.go",
		"crypto.go",
		"handle.go",
		"net.go",
		"state.go",
		"validation.go",
		"test_helper.go",
	}

	for _, util := range utilityFiles {
		if name == util {
			return true
		}
	}
	return false
}

// checkFile parses a Go file and checks all route handler functions for
// guard function usage.
func checkFile(t *testing.T, filePath string) []string {
	t.Helper()

	var violations []string

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		t.Logf("Warning: could not parse file %s: %v", filePath, err)
		return violations
	}

	for _, decl := range node.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		// Check if this is a route handler function (starts with "Route")
		if !strings.HasPrefix(funcDecl.Name.Name, "Route") {
			continue
		}

		// Check if the function invokes a guard
		if !invokesGuard(funcDecl) {
			violations = append(violations,
				"  - "+filePath+": "+funcDecl.Name.Name,
			)
		}
	}

	return violations
}

// invokesGuard checks if a function declaration contains a guard invocation.
// This includes:
//   - Calls to net.ReadParseAndGuard (standard JSON route pattern)
//   - Calls to functions starting with "guard" (custom guard patterns)
//   - Calls to functions containing "Guard" in the name
//   - Calls to helper functions that delegate to guarded handlers
//     (e.g., handleJSONDecrypt, handleStreamingEncrypt)
func invokesGuard(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Body == nil {
		return false
	}

	found := false
	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		if found {
			return false
		}

		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		funcName := extractFunctionName(callExpr)
		if funcName == "" {
			return true
		}

		// Check for ReadParseAndGuard
		if funcName == "ReadParseAndGuard" {
			found = true
			return false
		}

		// Check for guard functions (e.g., guardPolicyDeleteRequest)
		if strings.HasPrefix(funcName, "guard") {
			found = true
			return false
		}

		// Check for functions containing "Guard" (e.g., ReadParseAndGuard)
		if strings.Contains(funcName, "Guard") {
			found = true
			return false
		}

		// Check for handler functions that internally call guards
		// These are the cipher route helpers that have guards built-in
		guardedHandlers := []string{
			"handleJSONDecrypt",
			"handleStreamingDecrypt",
			"handleJSONEncrypt",
			"handleStreamingEncrypt",
		}
		for _, handler := range guardedHandlers {
			if funcName == handler {
				found = true
				return false
			}
		}

		return true
	})

	return found
}

// extractFunctionName extracts the function name from a call expression.
// Returns empty string if the name cannot be determined.
func extractFunctionName(callExpr *ast.CallExpr) string {
	switch fn := callExpr.Fun.(type) {
	case *ast.Ident:
		// Direct function call: guardFoo()
		return fn.Name
	case *ast.SelectorExpr:
		// Method or package call: net.ReadParseAndGuard()
		return fn.Sel.Name
	case *ast.IndexListExpr:
		// Generic function call: net.ReadParseAndGuard[T, U]()
		if sel, ok := fn.X.(*ast.SelectorExpr); ok {
			return sel.Sel.Name
		}
		if ident, ok := fn.X.(*ast.Ident); ok {
			return ident.Name
		}
	}
	return ""
}
