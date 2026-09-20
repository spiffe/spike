//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package layout

import (
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoFileMixesExportedAndUnexportedFunctions walks every Go file in the
// module and fails when a single file declares both exported and unexported
// functions or methods. Test files count their Test* and Benchmark*
// functions as exported, so their unexported helpers must live in the
// package's test helper file.
func TestNoFileMixesExportedAndUnexportedFunctions(t *testing.T) {
	root := moduleRoot(t)

	var offenders []string
	walkErr := filepath.WalkDir(root, func(
		path string, d fs.DirEntry, err error,
	) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skippedDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		exported, unexported, names := classify(t, path)
		if exported > 0 && unexported > 0 {
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				rel = path
			}
			offenders = append(offenders,
				rel+" mixes exported and unexported functions; move "+
					strings.Join(names, ", ")+" to a sibling file")
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("failed to walk %s: %v", root, walkErr)
	}

	for _, o := range offenders {
		t.Error(o)
	}
}
