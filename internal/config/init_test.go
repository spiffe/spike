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
	// Save original env var and restore after test.
	originalVal := os.Getenv(env.NexusDataDir)
	defer func() {
		if originalVal == "" {
			os.Unsetenv(env.NexusDataDir)
		} else {
			os.Setenv(env.NexusDataDir, originalVal)
		}
	}()

	t.Run("empty env var returns empty string", func(t *testing.T) {
		os.Unsetenv(env.NexusDataDir)
		result := tryCustomNexusDataDir("test")
		if result != "" {
			t.Errorf("expected empty string, got %q", result)
		}
	})

	t.Run("invalid directory returns empty string", func(t *testing.T) {
		// Use a restricted path.
		os.Setenv(env.NexusDataDir, "/etc")
		result := tryCustomNexusDataDir("test")
		if result != "" {
			t.Errorf("expected empty string for restricted path, got %q", result)
		}
	})

	t.Run("non-existent parent returns empty string", func(t *testing.T) {
		os.Setenv(env.NexusDataDir, "/nonexistent/path/to/dir")
		result := tryCustomNexusDataDir("test")
		if result != "" {
			t.Errorf("expected empty string for non-existent path, got %q", result)
		}
	})

	t.Run("valid directory creates data subdirectory", func(t *testing.T) {
		tempDir := t.TempDir()
		os.Setenv(env.NexusDataDir, tempDir)

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
	// Save original env var and restore after test.
	originalVal := os.Getenv(env.PilotRecoveryDir)
	defer func() {
		if originalVal == "" {
			os.Unsetenv(env.PilotRecoveryDir)
		} else {
			os.Setenv(env.PilotRecoveryDir, originalVal)
		}
	}()

	t.Run("empty env var returns empty string", func(t *testing.T) {
		os.Unsetenv(env.PilotRecoveryDir)
		result := tryCustomPilotRecoveryDir("test")
		if result != "" {
			t.Errorf("expected empty string, got %q", result)
		}
	})

	t.Run("invalid directory returns empty string", func(t *testing.T) {
		os.Setenv(env.PilotRecoveryDir, "/etc")
		result := tryCustomPilotRecoveryDir("test")
		if result != "" {
			t.Errorf("expected empty string for restricted path, got %q", result)
		}
	})

	t.Run("valid directory creates recover subdirectory", func(t *testing.T) {
		tempDir := t.TempDir()
		os.Setenv(env.PilotRecoveryDir, tempDir)

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
		info, err := os.Stat(result)
		if err != nil {
			t.Errorf("directory was not created: %v", err)
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
		info, err := os.Stat(result)
		if err != nil {
			t.Errorf("directory was not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("path is not a directory")
		}
	})
}

func TestCreateTempNexusDataDir(t *testing.T) {
	// Save and restore USER env var.
	originalUser := os.Getenv("USER")
	defer os.Setenv("USER", originalUser)

	t.Run("creates directory with USER env var", func(t *testing.T) {
		os.Setenv("USER", "testuser")
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
		os.RemoveAll("/tmp/.spike-testuser")
	})

	t.Run("uses 'spike' when USER is empty", func(t *testing.T) {
		os.Setenv("USER", "")
		result := createTempNexusDataDir("test")

		expectedPath := "/tmp/.spike-spike/data"
		if result != expectedPath {
			t.Errorf("expected %q, got %q", expectedPath, result)
		}

		// Clean up.
		os.RemoveAll("/tmp/.spike-spike")
	})
}

func TestCreateTempPilotRecoveryDir(t *testing.T) {
	originalUser := os.Getenv("USER")
	defer os.Setenv("USER", originalUser)

	t.Run("creates directory with USER env var", func(t *testing.T) {
		os.Setenv("USER", "testuser2")
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
		os.RemoveAll("/tmp/.spike-testuser2")
	})

	t.Run("uses 'spike' when USER is empty", func(t *testing.T) {
		os.Setenv("USER", "")
		result := createTempPilotRecoveryDir("test")

		expectedPath := "/tmp/.spike-spike/recover"
		if result != expectedPath {
			t.Errorf("expected %q, got %q", expectedPath, result)
		}

		// Clean up.
		os.RemoveAll("/tmp/.spike-spike")
	})
}

func TestInitNexusDataFolder_ResolutionOrder(t *testing.T) {
	// This test verifies the priority order:
	// 1. SPIKE_NEXUS_DATA_DIR (custom)
	// 2. ~/.spike/data (home)
	// 3. /tmp/.spike-$USER/data (temp fallback)

	originalVal := os.Getenv(env.NexusDataDir)
	defer func() {
		if originalVal == "" {
			os.Unsetenv(env.NexusDataDir)
		} else {
			os.Setenv(env.NexusDataDir, originalVal)
		}
	}()

	t.Run("custom dir takes priority over home", func(t *testing.T) {
		tempDir := t.TempDir()
		os.Setenv(env.NexusDataDir, tempDir)

		result := initNexusDataFolder()

		// Should use custom directory, not home.
		if !strings.HasPrefix(result, tempDir) {
			t.Errorf("expected path under %q, got %q", tempDir, result)
		}
	})

	t.Run("home dir used when custom not set", func(t *testing.T) {
		os.Unsetenv(env.NexusDataDir)

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
	originalVal := os.Getenv(env.PilotRecoveryDir)
	defer func() {
		if originalVal == "" {
			os.Unsetenv(env.PilotRecoveryDir)
		} else {
			os.Setenv(env.PilotRecoveryDir, originalVal)
		}
	}()

	t.Run("custom dir takes priority over home", func(t *testing.T) {
		tempDir := t.TempDir()
		os.Setenv(env.PilotRecoveryDir, tempDir)

		result := initPilotRecoveryFolder()

		if !strings.HasPrefix(result, tempDir) {
			t.Errorf("expected path under %q, got %q", tempDir, result)
		}
	})

	t.Run("home dir used when custom not set", func(t *testing.T) {
		os.Unsetenv(env.PilotRecoveryDir)

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

	originalNexusDir := os.Getenv(env.NexusDataDir)
	originalRecoveryDir := os.Getenv(env.PilotRecoveryDir)
	defer func() {
		if originalNexusDir == "" {
			os.Unsetenv(env.NexusDataDir)
		} else {
			os.Setenv(env.NexusDataDir, originalNexusDir)
		}
		if originalRecoveryDir == "" {
			os.Unsetenv(env.PilotRecoveryDir)
		} else {
			os.Setenv(env.PilotRecoveryDir, originalRecoveryDir)
		}
	}()

	t.Run("nexus data dir has 0700 permissions", func(t *testing.T) {
		customDir := filepath.Join(tempDir, "nexus-test")
		if err := os.MkdirAll(customDir, 0755); err != nil {
			t.Fatalf("failed to create test directory: %v", err)
		}
		os.Setenv(env.NexusDataDir, customDir)

		result := tryCustomNexusDataDir("test")
		if result == "" {
			t.Fatal("expected non-empty result")
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
		os.Setenv(env.PilotRecoveryDir, customDir)

		result := tryCustomPilotRecoveryDir("test")
		if result == "" {
			t.Fatal("expected non-empty result")
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
