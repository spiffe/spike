//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package trust

import (
	"os"
	"testing"
)

// The trust module uses environment variables to determine the trust root.
// The default is "spike.ist" if SPIKE_TRUST_ROOT_PILOT is not set.
// The expected SPIFFE ID patterns are (based on SDK spiffeid module):
//   - Pilot: spiffe://<trust_root>/spike/pilot/role/superuser
//   - Recover: spiffe://<trust_root>/spike/pilot/role/recover
//   - Restore: spiffe://<trust_root>/spike/pilot/role/restore

func TestAuthenticateForPilot_ValidPilotID(t *testing.T) {
	// Enable panic on FatalErr for testing
	t.Setenv("SPIKE_STACK_TRACES_ON_LOG_FATAL", "true")
	// Set trust root for pilot
	t.Setenv("SPIKE_TRUST_ROOT_PILOT", "spike.ist")

	// Use the pilot SPIFFE ID pattern (includes role/superuser)
	pilotID := "spiffe://spike.ist/spike/pilot/role/superuser"

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("AuthenticateForPilot() panicked for valid pilot ID: %v", r)
		}
	}()

	AuthenticateForPilot(pilotID)
}

func TestAuthenticateForPilot_InvalidID(t *testing.T) {
	// Enable panic on FatalErr for testing
	t.Setenv("SPIKE_STACK_TRACES_ON_LOG_FATAL", "true")
	// Set trust root for pilot
	t.Setenv("SPIKE_TRUST_ROOT_PILOT", "spike.ist")

	// Use an invalid SPIFFE ID
	invalidID := "spiffe://spike.ist/some/other/workload"

	defer func() {
		if r := recover(); r == nil {
			t.Error("AuthenticateForPilot() should have panicked for invalid ID")
		}
	}()

	AuthenticateForPilot(invalidID)
	t.Error("Should have panicked before reaching here")
}

func TestAuthenticateForPilotRecover_ValidRecoverID(t *testing.T) {
	// Enable panic on FatalErr for testing
	t.Setenv("SPIKE_STACK_TRACES_ON_LOG_FATAL", "true")
	// Set trust root for pilot recover
	t.Setenv("SPIKE_TRUST_ROOT_PILOT_RECOVER", "spike.ist")

	// Use the recover SPIFFE ID pattern
	recoverID := "spiffe://spike.ist/spike/pilot/role/recover"

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("AuthenticateForPilotRecover() panicked for valid ID: %v", r)
		}
	}()

	AuthenticateForPilotRecover(recoverID)
}

func TestAuthenticateForPilotRecover_InvalidID(t *testing.T) {
	// Enable panic on FatalErr for testing
	t.Setenv("SPIKE_STACK_TRACES_ON_LOG_FATAL", "true")
	// Set trust root for pilot recover
	t.Setenv("SPIKE_TRUST_ROOT_PILOT_RECOVER", "spike.ist")

	// Use an invalid SPIFFE ID (pilot, not pilot/role/recover)
	invalidID := "spiffe://spike.ist/spike/pilot"

	defer func() {
		if r := recover(); r == nil {
			t.Error("AuthenticateForPilotRecover() should have panicked " +
				"for invalid ID")
		}
	}()

	AuthenticateForPilotRecover(invalidID)
	t.Error("Should have panicked before reaching here")
}

func TestAuthenticateForPilotRestore_ValidRestoreID(t *testing.T) {
	// Enable panic on FatalErr for testing
	t.Setenv("SPIKE_STACK_TRACES_ON_LOG_FATAL", "true")
	// Set trust root for pilot restore
	t.Setenv("SPIKE_TRUST_ROOT_PILOT_RESTORE", "spike.ist")

	// Use the restore SPIFFE ID pattern
	restoreID := "spiffe://spike.ist/spike/pilot/role/restore"

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("AuthenticateForPilotRestore() panicked for valid ID: %v", r)
		}
	}()

	AuthenticateForPilotRestore(restoreID)
}

func TestAuthenticateForPilotRestore_InvalidID(t *testing.T) {
	// Enable panic on FatalErr for testing
	t.Setenv("SPIKE_STACK_TRACES_ON_LOG_FATAL", "true")
	// Set trust root for pilot restore
	t.Setenv("SPIKE_TRUST_ROOT_PILOT_RESTORE", "spike.ist")

	// Use an invalid SPIFFE ID (pilot, not pilot/role/restore)
	invalidID := "spiffe://spike.ist/spike/pilot"

	defer func() {
		if r := recover(); r == nil {
			t.Error("AuthenticateForPilotRestore() should have panicked " +
				"for invalid ID")
		}
	}()

	AuthenticateForPilotRestore(invalidID)
	t.Error("Should have panicked before reaching here")
}

func TestMain(m *testing.M) {
	// Ensure SPIKE_STACK_TRACES_ON_LOG_FATAL is set for all tests
	os.Setenv("SPIKE_STACK_TRACES_ON_LOG_FATAL", "true")
	os.Exit(m.Run())
}
