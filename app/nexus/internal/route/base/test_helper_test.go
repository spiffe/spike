//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// scanRouteTree walks every package directory under root, except the
// router package itself, and reports every route handler that reaches no
// guard.
//
// Nothing is skipped by file name. A route handler is recognized by its
// signature, not by its name or its file, so a handler cannot escape the
// scan by being renamed or by living in a differently named file.
//
// Parameters:
//   - t: The test, used for helper marking and parse failures.
//   - root: The directory that holds the route packages.
//   - routerDir: The router package's own directory. Its Route function
//     shares the handler signature but dispatches to handlers instead of
//     guarding a request itself.
//
// Returns:
//   - []string: One entry per route handler that invokes no guard.
func scanRouteTree(t *testing.T, root, routerDir string) []string {
	t.Helper()

	var violations []string
	walkErr := filepath.WalkDir(root, func(
		path string, d fs.DirEntry, err error,
	) error {
		if err != nil {
			return err
		}
		if !d.IsDir() || path == routerDir {
			return nil
		}
		violations = append(violations, checkDirectory(t, path)...)
		return nil
	})
	if walkErr != nil {
		t.Fatalf("failed to walk %s: %v", root, walkErr)
	}
	return violations
}

// checkDirectory parses every non-test Go file of one package directory
// and checks each route handler in it for a guard invocation, following
// same-package delegates transitively.
//
// Parameters:
//   - t: The test, used for helper marking and parse failures.
//   - dir: The package directory to scan.
//
// Returns:
//   - []string: One entry per route handler that invokes no guard.
func checkDirectory(t *testing.T, dir string) []string {
	t.Helper()

	pkgFuncs := parsePackageFuncs(t, dir)

	var handlers []string
	for name, fn := range pkgFuncs {
		if isRouteHandler(fn) {
			handlers = append(handlers, name)
		}
	}

	var violations []string
	for _, name := range handlers {
		if !invokesGuard(pkgFuncs[name], pkgFuncs, map[string]bool{}) {
			violations = append(violations,
				"  - "+filepath.Join(dir, name),
			)
		}
	}
	return violations
}

// parsePackageFuncs parses every non-test Go file in a directory and
// indexes its top-level function declarations by name.
//
// Parameters:
//   - t: The test, failed when a file does not parse.
//   - dir: The package directory to parse.
//
// Returns:
//   - map[string]*ast.FuncDecl: The functions declared in the package.
func parsePackageFuncs(
	t *testing.T, dir string,
) map[string]*ast.FuncDecl {
	t.Helper()

	files, globErr := filepath.Glob(filepath.Join(dir, "*.go"))
	if globErr != nil {
		t.Fatalf("failed to list %s: %v", dir, globErr)
	}

	funcs := map[string]*ast.FuncDecl{}
	fset := token.NewFileSet()
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		node, parseErr := parser.ParseFile(
			fset, file, nil, parser.SkipObjectResolution,
		)
		if parseErr != nil {
			t.Fatalf("failed to parse %s: %v", file, parseErr)
		}
		for _, decl := range node.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil {
				funcs[fn.Name.Name] = fn
			}
		}
	}
	return funcs
}

// isRouteHandler reports whether a function has the route handler
// signature: exactly the parameters (http.ResponseWriter, *http.Request,
// *journal.AuditEntry).
//
// Parameters:
//   - fn: The function declaration to inspect.
//
// Returns:
//   - bool: True when the parameter types match the handler contract.
func isRouteHandler(fn *ast.FuncDecl) bool {
	want := []string{
		"http.ResponseWriter", "*http.Request", "*journal.AuditEntry",
	}

	var got []string
	for _, field := range fn.Type.Params.List {
		name := typeName(field.Type)
		// A field may declare several names of one type (w, x http...).
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		for i := 0; i < count; i++ {
			got = append(got, name)
		}
	}
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// typeName renders a parameter type expression the way it is written in
// source, for the small set of shapes a handler signature uses.
//
// Parameters:
//   - expr: The type expression.
//
// Returns:
//   - string: "pkg.Name", "*pkg.Name", or "Name"; empty for other shapes.
func typeName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + typeName(e.X)
	case *ast.SelectorExpr:
		if pkg, ok := e.X.(*ast.Ident); ok {
			return pkg.Name + "." + e.Sel.Name
		}
	}
	return ""
}

// invokesGuard reports whether a function reaches a guard, either directly
// or through same-package functions it calls or passes as values (for
// example the handlers handed to a content-type dispatcher). Recursion is
// bounded by the visited set.
//
// Parameters:
//   - fn: The function declaration to inspect.
//   - pkgFuncs: The package's functions, for following delegates.
//   - visited: The functions already inspected on this path.
//
// Returns:
//   - bool: True when a guard invocation is reachable from the body.
func invokesGuard(
	fn *ast.FuncDecl, pkgFuncs map[string]*ast.FuncDecl,
	visited map[string]bool,
) bool {
	if fn == nil || fn.Body == nil || visited[fn.Name.Name] {
		return false
	}
	visited[fn.Name.Name] = true

	found := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if found {
			return false
		}
		switch node := n.(type) {
		case *ast.CallExpr:
			name := extractFunctionName(node)
			if isGuardName(name) {
				found = true
				return false
			}
			if delegate, ok := pkgFuncs[name]; ok &&
				invokesGuard(delegate, pkgFuncs, visited) {
				found = true
				return false
			}
		case *ast.Ident:
			if delegate, ok := pkgFuncs[node.Name]; ok &&
				invokesGuard(delegate, pkgFuncs, visited) {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

// isGuardName reports whether a called function is a guard by the
// project's naming contract: net.ReadParseAndGuard, a function starting
// with "guard", or a function whose name contains "Guard".
//
// Parameters:
//   - name: The called function's name.
//
// Returns:
//   - bool: True when the name denotes a guard.
func isGuardName(name string) bool {
	return name == "ReadParseAndGuard" ||
		strings.HasPrefix(name, "guard") ||
		strings.Contains(name, "Guard")
}

// extractFunctionName extracts the function name from a call expression.
//
// Parameters:
//   - callExpr: The call expression to inspect.
//
// Returns:
//   - string: The called function's name, or an empty string when it cannot
//     be determined.
func extractFunctionName(callExpr *ast.CallExpr) string {
	switch fn := callExpr.Fun.(type) {
	case *ast.Ident:
		return fn.Name
	case *ast.SelectorExpr:
		return fn.Sel.Name
	case *ast.IndexListExpr:
		if sel, ok := fn.X.(*ast.SelectorExpr); ok {
			return sel.Sel.Name
		}
		if ident, ok := fn.X.(*ast.Ident); ok {
			return ident.Name
		}
	}
	return ""
}

// writeScanFixture writes a Go source file into a temporary package
// directory for exercising the scan against known-good and known-bad
// handlers. The files are parsed, never compiled, so imports need not
// resolve.
//
// Parameters:
//   - t: The test, failed when the file cannot be written.
//   - dir: The temporary package directory.
//   - name: The file name.
//   - body: The Go source.
func writeScanFixture(t *testing.T, dir, name, body string) {
	t.Helper()

	if writeErr := os.WriteFile(
		filepath.Join(dir, name), []byte(body), 0600,
	); writeErr != nil {
		t.Fatalf("failed to write %s: %v", name, writeErr)
	}
}

// countRouteHandlers counts the route handlers under root, excluding the
// router package, so a test can prove the scan is not passing vacuously.
//
// Parameters:
//   - t: The test, used for helper marking and parse failures.
//   - root: The directory that holds the route packages.
//   - routerDir: The router package's own directory, skipped.
//
// Returns:
//   - int: The number of functions with the route handler signature.
func countRouteHandlers(t *testing.T, root, routerDir string) int {
	t.Helper()

	count := 0
	walkErr := filepath.WalkDir(root, func(
		path string, d fs.DirEntry, err error,
	) error {
		if err != nil {
			return err
		}
		if !d.IsDir() || path == routerDir {
			return nil
		}
		for _, fn := range parsePackageFuncs(t, path) {
			if isRouteHandler(fn) {
				count++
			}
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("failed to walk %s: %v", root, walkErr)
	}
	return count
}
