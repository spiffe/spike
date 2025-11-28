//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spiffe/spike-sdk-go/config/env"
)

func TestTryCustomNexusDataDir(t *testing.T) {
	t.Run("empty env var returns empty string", func(t *testing.T) {
		// Ensure the env var is not set for this test
		t.Setenv(env.NexusDataDir, "")
		result := tryCustomNexusDataDir("test")
		if result != "" {
			t.Errorf("expected empty string, got %q", result)
		}
	})

	t.Run("invalid directory returns empty string", func(t *testing.T) {
		// Use a restricted path.
		t.Setenv(env.NexusDataDir, "/etc")
		result := tryCustomNexusDataDir("test")
		if result != "" {
			t.Errorf("expected empty string for restricted path, got %q", result)
		}
	})

	t.Run("non-existent parent returns empty string", func(t *testing.T) {
		t.Setenv(env.NexusDataDir, "/nonexistent/path/to/dir")
		result := tryCustomNexusDataDir("test")
		if result != "" {
			t.Errorf("expected empty string for non-existent path, got %q", result)
		}
	})

	t.Run("valid directory creates data subdirectory", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv(env.NexusDataDir, tempDir)

		result := tryCustomNexusDataDir("test")

		expectedPath := filepath.Join(tempDir, spikeDataFolderName)
		if result != expectedPath {
			t.Errorf("expected %q, got %q", expectedPath, result)
		}

		// Verify the directory was created.
		info, err := os.Stat(result)
		if err != nil {
			t.Errorf("directory was not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("path is not a directory")
		}

		// Verify permissions are 0700.
		if info.Mode().Perm() != 0700 {
			t.Errorf("expected permissions 0700, got %o", info.Mode().Perm())
		}
	})
}

func TestTryCustomPilotRecoveryDir(t *testing.T) {
	t.Run("empty env var returns empty string", func(t *testing.T) {
		t.Setenv(env.PilotRecoveryDir, "")
		result := tryCustomPilotRecoveryDir("test")
		if result != "" {
			t.Errorf("expected empty string, got %q", result)
		}
	})

	t.Run("invalid directory returns empty string", func(t *testing.T) {
		t.Setenv(env.PilotRecoveryDir, "/etc")
		result := tryCustomPilotRecoveryDir("test")
		if result != "" {
			t.Errorf("expected empty string for restricted path, got %q", result)
		}
	})

	t.Run("valid directory creates recover subdirectory", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv(env.PilotRecoveryDir, tempDir)

		result := tryCustomPilotRecoveryDir("test")

		expectedPath := filepath.Join(tempDir, spikeRecoveryFolderName)
		if result != expectedPath {
			t.Errorf("expected %q, got %q", expectedPath, result)
		}

		// Verify the directory was created.
		info, err := os.Stat(result)
		if err != nil {
			t.Errorf("directory was not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("path is not a directory")
		}

		// Verify permissions are 0700.
		if info.Mode().Perm() != 0700 {
			t.Errorf("expected permissions 0700, got %o", info.Mode().Perm())
		}
	})
}

func TestTryHomeNexusDataDir(t *testing.T) {
	// This test depends on the system having a home directory.
	// Skip if HOME is not set.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home directory available")
	}

	t.Run("creates directory under home", func(t *testing.T) {
		result := tryHomeNexusDataDir("test")

		expectedPath := filepath.Join(homeDir, spikeHiddenFolderName,
			spikeDataFolderName)
		if result != expectedPath {
			t.Errorf("expected %q, got %q", expectedPath, result)
		}

		// Verify the directory exists.
		info, statErr := os.Stat(result)
		if statErr != nil {
			t.Errorf("directory was not created: %v", statErr)
		}
		if !info.IsDir() {
			t.Error("path is not a directory")
		}
	})
}

func TestTryHomePilotRecoveryDir(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home directory available")
	}

	t.Run("creates directory under home", func(t *testing.T) {
		result := tryHomePilotRecoveryDir("test")

		expectedPath := filepath.Join(homeDir, spikeHiddenFolderName,
			spikeRecoveryFolderName)
		if result != expectedPath {
			t.Errorf("expected %q, got %q", expectedPath, result)
		}

		// Verify the directory exists.
		info, statErr := os.Stat(result)
		if statErr != nil {
			t.Errorf("directory was not created: %v", statErr)
		}
		if !info.IsDir() {
			t.Error("path is not a directory")
		}
	})
}

func TestCreateTempNexusDataDir(t *testing.T) {
	t.Run("creates directory with USER env var", func(t *testing.T) {
		t.Setenv("USER", "testuser")
		result := createTempNexusDataDir("test")

		expectedPath := "/tmp/.spike-testuser/data"
		if result != expectedPath {
			t.Errorf("expected %q, got %q", expectedPath, result)
		}

		// Verify the directory exists.
		info, err := os.Stat(result)
		if err != nil {
			t.Errorf("directory was not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("path is not a directory")
		}

		// Verify permissions are 0700.
		if info.Mode().Perm() != 0700 {
			t.Errorf("expected permissions 0700, got %o", info.Mode().Perm())
		}

		// Clean up.
		if removeErr := os.RemoveAll("/tmp/.spike-testuser"); removeErr != nil {
			t.Logf("failed to clean up test directory: %v", removeErr)
		}
	})

	t.Run("uses 'spike' when USER is empty", func(t *testing.T) {
		t.Setenv("USER", "")
		result := createTempNexusDataDir("test")

		expectedPath := "/tmp/.spike-spike/data"
		if result != expectedPath {
			t.Errorf("expected %q, got %q", expectedPath, result)
		}

		// Clean up.
		if removeErr := os.RemoveAll("/tmp/.spike-spike"); removeErr != nil {
			t.Logf("failed to clean up test directory: %v", removeErr)
		}
	})
}

func TestCreateTempPilotRecoveryDir(t *testing.T) {
	t.Run("creates directory with USER env var", func(t *testing.T) {
		t.Setenv("USER", "testuser2")
		result := createTempPilotRecoveryDir("test")

		expectedPath := "/tmp/.spike-testuser2/recover"
		if result != expectedPath {
			t.Errorf("expected %q, got %q", expectedPath, result)
		}

		// Verify the directory exists.
		info, err := os.Stat(result)
		if err != nil {
			t.Errorf("directory was not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("path is not a directory")
		}

		// Verify permissions are 0700.
		if info.Mode().Perm() != 0700 {
			t.Errorf("expected permissions 0700, got %o", info.Mode().Perm())
		}

		// Clean up.
		if removeErr := os.RemoveAll("/tmp/.spike-testuser2"); removeErr != nil {
			t.Logf("failed to clean up test directory: %v", removeErr)
		}
	})

	t.Run("uses 'spike' when USER is empty", func(t *testing.T) {
		t.Setenv("USER", "")
		result := createTempPilotRecoveryDir("test")

		expectedPath := "/tmp/.spike-spike/recover"
		if result != expectedPath {
			t.Errorf("expected %q, got %q", expectedPath, result)
		}

		// Clean up.
		if removeErr := os.RemoveAll("/tmp/.spike-spike"); removeErr != nil {
			t.Logf("failed to clean up test directory: %v", removeErr)
		}
	})
}

func TestInitNexusDataFolder_ResolutionOrder(t *testing.T) {
	// This test verifies the priority order:
	// 1. SPIKE_NEXUS_DATA_DIR (custom)
	// 2. ~/.spike/data (home)
	// 3. /tmp/.spike-$USER/data (temp fallback)

	t.Run("custom dir takes priority over home", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv(env.NexusDataDir, tempDir)

		result := initNexusDataFolder()

		// Should use a custom directory, not home.
		if !strings.HasPrefix(result, tempDir) {
			t.Errorf("expected path under %q, got %q", tempDir, result)
		}
	})

	t.Run("home dir used when custom not set", func(t *testing.T) {
		t.Setenv(env.NexusDataDir, "")

		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Skip("no home directory available")
		}

		result := initNexusDataFolder()

		expectedPrefix := filepath.Join(homeDir, spikeHiddenFolderName)
		if !strings.HasPrefix(result, expectedPrefix) {
			t.Errorf("expected path under %q, got %q", expectedPrefix, result)
		}
	})
}

func TestInitPilotRecoveryFolder_ResolutionOrder(t *testing.T) {
	t.Run("custom dir takes priority over home", func(t *testing.T) {
		tempDir := t.TempDir()
		t.Setenv(env.PilotRecoveryDir, tempDir)

		result := initPilotRecoveryFolder()

		if !strings.HasPrefix(result, tempDir) {
			t.Errorf("expected path under %q, got %q", tempDir, result)
		}
	})

	t.Run("home dir used when custom not set", func(t *testing.T) {
		t.Setenv(env.PilotRecoveryDir, "")

		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Skip("no home directory available")
		}

		result := initPilotRecoveryFolder()

		expectedPrefix := filepath.Join(homeDir, spikeHiddenFolderName)
		if !strings.HasPrefix(result, expectedPrefix) {
			t.Errorf("expected path under %q, got %q", expectedPrefix, result)
		}
	})
}

func TestDirectoryPermissions(t *testing.T) {
	// Verify that all created directories have 0700 permissions.
	tempDir := t.TempDir()

	t.Run("nexus data dir has 0700 permissions", func(t *testing.T) {
		customDir := filepath.Join(tempDir, "nexus-test")
		if err := os.MkdirAll(customDir, 0755); err != nil {
			t.Fatalf("failed to create test directory: %v", err)
		}
		t.Setenv(env.NexusDataDir, customDir)

		result := tryCustomNexusDataDir("test")
		if result == "" {
			t.Fatal("expected non-empty result")
			return
		}

		info, err := os.Stat(result)
		if err != nil {
			t.Fatalf("failed to stat directory: %v", err)
		}

		if info.Mode().Perm() != 0700 {
			t.Errorf("expected permissions 0700, got %o", info.Mode().Perm())
		}
	})

	t.Run("recovery dir has 0700 permissions", func(t *testing.T) {
		customDir := filepath.Join(tempDir, "recovery-test")
		if err := os.MkdirAll(customDir, 0755); err != nil {
			t.Fatalf("failed to create test directory: %v", err)
		}
		t.Setenv(env.PilotRecoveryDir, customDir)

		result := tryCustomPilotRecoveryDir("test")
		if result == "" {
			t.Fatal("expected non-empty result")
			return
		}

		info, err := os.Stat(result)
		if err != nil {
			t.Fatalf("failed to stat directory: %v", err)
		}

		if info.Mode().Perm() != 0700 {
			t.Errorf("expected permissions 0700, got %o", info.Mode().Perm())
		}
	})
}
