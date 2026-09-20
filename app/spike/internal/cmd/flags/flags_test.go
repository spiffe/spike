//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package flags

import (
	"strings"
	"testing"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

func TestString_ReturnsValue(t *testing.T) {
	cmd := newCommand(t)
	if setErr := cmd.Flags().Set("name", "value"); setErr != nil {
		t.Fatalf("failed to set the flag: %v", setErr)
	}

	got, readErr := String(cmd, "name")
	if readErr != nil {
		t.Fatalf("String() error = %v, want nil", readErr)
	}
	if got != "value" {
		t.Errorf("String() = %q, want %q", got, "value")
	}
}

func TestString_ReturnsDefault(t *testing.T) {
	got, readErr := String(newCommand(t), "name")
	if readErr != nil {
		t.Fatalf("String() error = %v, want nil", readErr)
	}
	if got != "default" {
		t.Errorf("String() = %q, want %q", got, "default")
	}
}

func TestInt_ReturnsValue(t *testing.T) {
	cmd := newCommand(t)
	if setErr := cmd.Flags().Set("count", "42"); setErr != nil {
		t.Fatalf("failed to set the flag: %v", setErr)
	}

	got, readErr := Int(cmd, "count")
	if readErr != nil {
		t.Fatalf("Int() error = %v, want nil", readErr)
	}
	if got != 42 {
		t.Errorf("Int() = %d, want %d", got, 42)
	}
}

func TestString_MissingFlagIsAnError(t *testing.T) {
	got, readErr := String(newCommand(t), "missing")
	if readErr == nil {
		t.Fatal("String() error = nil, want an error for a missing flag")
	}
	if got != "" {
		t.Errorf("String() = %q, want an empty string", got)
	}
	if !readErr.Is(sdkErrors.ErrDataInvalidInput) {
		t.Errorf("String() error code = %v, want ErrDataInvalidInput",
			readErr.Code)
	}
	if !strings.Contains(readErr.Error(), "--missing") {
		t.Errorf("String() error = %q, want it to name the flag",
			readErr.Error())
	}
}

func TestString_WrongTypeIsAnError(t *testing.T) {
	// "count" is registered as an integer flag.
	if _, readErr := String(newCommand(t), "count"); readErr == nil {
		t.Error("String() error = nil, want an error for a non-string flag")
	}
}

func TestInt_MissingFlagIsAnError(t *testing.T) {
	got, readErr := Int(newCommand(t), "missing")
	if readErr == nil {
		t.Fatal("Int() error = nil, want an error for a missing flag")
	}
	if got != 0 {
		t.Errorf("Int() = %d, want 0", got)
	}
	if !readErr.Is(sdkErrors.ErrDataInvalidInput) {
		t.Errorf("Int() error code = %v, want ErrDataInvalidInput",
			readErr.Code)
	}
}
