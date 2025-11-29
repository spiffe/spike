//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenInput_Stdin(t *testing.T) {
	in, cleanup, err := openInput("")

	if err != nil {
		t.Errorf("openInput(\"\") error = %v, want nil", err)
	}

	if in != os.Stdin {
		t.Error("openInput(\"\") should return os.Stdin")
	}

	if cleanup == nil {
		t.Error("openInput(\"\") should return a cleanup function")
	}

	// Cleanup should be safe to call (should not close stdin)
	cleanup()
}

func TestOpenInput_ValidFile(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test-input.txt")
	content := []byte("test content for reading")

	if err := os.WriteFile(tempFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	in, cleanup, err := openInput(tempFile)

	if err != nil {
		t.Errorf("openInput(%q) error = %v, want nil", tempFile, err)
	}

	if in == nil {
		t.Fatal("openInput() should return a reader")
	}

	if cleanup == nil {
		t.Fatal("openInput() should return a cleanup function")
	}

	// Read content to verify it works
	data, readErr := io.ReadAll(in)
	if readErr != nil {
		t.Errorf("Failed to read from input: %v", readErr)
	}

	if string(data) != string(content) {
		t.Errorf("Read content = %q, want %q", string(data), string(content))
	}

	// Cleanup should close the file
	cleanup()
}

func TestOpenInput_NonexistentFile(t *testing.T) {
	in, cleanup, err := openInput("/nonexistent/path/to/file.txt")

	if err == nil {
		t.Error("openInput() should return error for nonexistent file")
	}

	if in != nil {
		t.Error("openInput() should return nil reader on error")
	}

	if cleanup != nil {
		t.Error("openInput() should return nil cleanup on error")
	}

	// Verify error type
	if err != nil && err.Code != "fs_file_open_failed" {
		t.Errorf("Error code = %q, want %q", err.Code, "fs_file_open_failed")
	}
}

func TestOpenOutput_Stdout(t *testing.T) {
	out, cleanup, err := openOutput("")

	if err != nil {
		t.Errorf("openOutput(\"\") error = %v, want nil", err)
	}

	if out != os.Stdout {
		t.Error("openOutput(\"\") should return os.Stdout")
	}

	if cleanup == nil {
		t.Error("openOutput(\"\") should return a cleanup function")
	}

	// Cleanup should be safe to call (should not close stdout)
	cleanup()
}

func TestOpenOutput_ValidFile(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test-output.txt")

	out, cleanup, err := openOutput(tempFile)

	if err != nil {
		t.Errorf("openOutput(%q) error = %v, want nil", tempFile, err)
	}

	if out == nil {
		t.Fatal("openOutput() should return a writer")
	}

	if cleanup == nil {
		t.Fatal("openOutput() should return a cleanup function")
	}

	// Write content to verify it works
	testContent := "test output content"
	n, writeErr := out.Write([]byte(testContent))
	if writeErr != nil {
		t.Errorf("Failed to write to output: %v", writeErr)
	}
	if n != len(testContent) {
		t.Errorf("Wrote %d bytes, want %d", n, len(testContent))
	}

	// Cleanup should close the file
	cleanup()

	// Verify content was written
	data, readErr := os.ReadFile(tempFile)
	if readErr != nil {
		t.Errorf("Failed to read output file: %v", readErr)
	}

	if string(data) != testContent {
		t.Errorf("File content = %q, want %q", string(data), testContent)
	}
}

func TestOpenOutput_InvalidPath(t *testing.T) {
	// Try to create a file in a nonexistent directory
	out, cleanup, err := openOutput("/nonexistent/directory/file.txt")

	if err == nil {
		t.Error("openOutput() should return error for invalid path")
	}

	if out != nil {
		t.Error("openOutput() should return nil writer on error")
	}

	if cleanup != nil {
		t.Error("openOutput() should return nil cleanup on error")
	}

	// Verify error type
	if err != nil && err.Code != "fs_file_open_failed" {
		t.Errorf("Error code = %q, want %q", err.Code, "fs_file_open_failed")
	}
}

func TestOpenInput_CleanupIdempotent(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test-cleanup.txt")

	if err := os.WriteFile(tempFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, cleanup, err := openInput(tempFile)
	if err != nil {
		t.Fatalf("openInput() error = %v", err)
	}

	// Calling cleanup multiple times should be safe
	cleanup()
	cleanup() // Should not panic
}

func TestOpenOutput_CleanupIdempotent(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test-cleanup-out.txt")

	_, cleanup, err := openOutput(tempFile)
	if err != nil {
		t.Fatalf("openOutput() error = %v", err)
	}

	// Calling cleanup multiple times should be safe
	cleanup()
	cleanup() // Should not panic
}

func TestOpenOutput_OverwritesExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "existing-file.txt")

	// Create a file with initial content
	initialContent := "initial content that should be overwritten"
	if err := os.WriteFile(tempFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	out, cleanup, err := openOutput(tempFile)
	if err != nil {
		t.Fatalf("openOutput() error = %v", err)
	}

	// Write new content
	newContent := "new"
	_, _ = out.Write([]byte(newContent))
	cleanup()

	// Verify the file was overwritten (not appended)
	data, _ := os.ReadFile(tempFile)
	if string(data) != newContent {
		t.Errorf("File content = %q, want %q (file should be overwritten)",
			string(data), newContent)
	}
}
