//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package layout

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

// moduleRoot returns the directory that holds the module's go.mod, found by
// walking up from the test's working directory.
//
// Parameters:
//   - t: The test, failed when no go.mod is found.
//
// Returns:
//   - string: The absolute path of the module root.
func moduleRoot(t *testing.T) string {
	t.Helper()

	dir, wdErr := os.Getwd()
	if wdErr != nil {
		t.Fatalf("failed to resolve the working directory: %v", wdErr)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found above the working directory")
			return ""
		}
		dir = parent
	}
}

// skippedDir reports whether a directory is outside the scope of the
// layout rules: version control, vendored or generated content, and the
// documentation site.
//
// Parameters:
//   - name: The base name of the directory.
//
// Returns:
//   - bool: True when the directory must not be walked.
func skippedDir(name string) bool {
	switch name {
	case ".git", "vendor", "node_modules", "docs-src", "docs", "bin":
		return true
	}
	return false
}

// classify parses one Go file and counts its top-level functions and
// methods by export status.
//
// Parameters:
//   - t: The test, failed when the file does not parse.
//   - path: The file to parse.
//
// Returns:
//   - int: The number of exported functions and methods.
//   - int: The number of unexported functions and methods.
//   - []string: The names of the unexported ones, for the failure message.
func classify(t *testing.T, path string) (int, int, []string) {
	t.Helper()

	fset := token.NewFileSet()
	file, parseErr := parser.ParseFile(
		fset, path, nil, parser.SkipObjectResolution,
	)
	if parseErr != nil {
		t.Fatalf("failed to parse %s: %v", path, parseErr)
		return 0, 0, nil
	}

	var exported, unexported int
	var names []string
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if ast.IsExported(fn.Name.Name) {
			exported++
			continue
		}
		unexported++
		names = append(names, fn.Name.Name)
	}
	return exported, unexported, names
}
