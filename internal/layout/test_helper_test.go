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
	"strings"
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

// spireVersions reads the KEY="value" assignments of hack/lib/versions.sh,
// the single source of truth for the SPIRE versions SPIKE pins.
//
// Parameters:
//   - t: The test, failed when the file cannot be read or holds no pins.
//   - root: The module root.
//
// Returns:
//   - map[string]string: The pinned versions keyed by variable name.
func spireVersions(t *testing.T, root string) map[string]string {
	t.Helper()

	path := filepath.Join(root, "hack", "lib", "versions.sh")
	content, readErr := os.ReadFile(filepath.Clean(path))
	if readErr != nil {
		t.Fatalf("failed to read %s: %v", path, readErr)
	}

	versions := map[string]string{}
	for _, line := range strings.Split(string(content), "\n") {
		key, value, ok := strings.Cut(line, "=")
		if !ok || strings.HasPrefix(line, "#") || strings.ContainsAny(
			key, " \t",
		) {
			continue
		}
		versions[key] = strings.Trim(value, `"`)
	}
	if len(versions) == 0 {
		t.Fatalf("%s declares no versions", path)
	}
	return versions
}

// fileContains fails the test when a file does not contain every expected
// substring, naming the file and the missing text so a doc page that shows
// an old version is easy to fix.
//
// Parameters:
//   - t: The test.
//   - root: The module root.
//   - rel: The file, relative to the module root.
//   - expected: The substrings the file must contain.
func fileContains(t *testing.T, root, rel string, expected ...string) {
	t.Helper()

	content, readErr := os.ReadFile(filepath.Clean(filepath.Join(root, rel)))
	if readErr != nil {
		t.Fatalf("failed to read %s: %v", rel, readErr)
	}
	for _, want := range expected {
		if !strings.Contains(string(content), want) {
			t.Errorf("%s does not mention %q; update it to match "+
				"hack/lib/versions.sh", rel, want)
		}
	}
}
