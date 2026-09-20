//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package logfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spiffe/spike-sdk-go/log"
)

// TestPath_UsesHomeThenTemp checks both branches of the path resolution.
func TestPath_UsesHomeThenTemp(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	want := filepath.Join(home, ".spike", "pilot.log")
	if got := Path(); got != want {
		t.Errorf("Path() = %q, want %q", got, want)
	}

	t.Setenv("HOME", "")
	t.Setenv("USER", "tester")
	want = filepath.Join("/tmp/.spike-tester", "pilot.log")
	if got := Path(); got != want {
		t.Errorf("Path() without HOME = %q, want %q", got, want)
	}
}

// TestRoute_BindsLoggerToFile verifies that after Route the SDK logger
// writes to the diagnostics file and that os.Stdout is restored. The SDK
// logger is a process-wide singleton, so this is the only test in the
// package that may call Route.
func TestRoute_BindsLoggerToFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	before := os.Stdout
	if routeErr := Route(); routeErr != nil {
		t.Fatalf("Route() failed: %v", routeErr)
	}
	if os.Stdout != before {
		t.Fatal("Route() did not restore os.Stdout")
	}

	const marker = "logfile-route-marker"
	log.Error("TestRoute_BindsLoggerToFile", "message", marker)

	logPath := filepath.Clean(Path())
	content, readErr := os.ReadFile(logPath)
	if readErr != nil {
		t.Fatalf("failed to read the log file: %v", readErr)
	}
	if !strings.Contains(string(content), marker) {
		t.Errorf("log file does not contain %q: %q", marker, content)
	}

	info, statErr := os.Stat(logPath)
	if statErr != nil {
		t.Fatalf("failed to stat the log file: %v", statErr)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("log file permissions = %o, want 0600", perm)
	}
}
