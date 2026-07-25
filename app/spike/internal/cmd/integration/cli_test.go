//go:build integration

//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestMain enforces the run-time gate. The `integration` build tag keeps
// these tests out of ordinary builds; SPIKE_INTEGRATION_TEST=1 is the
// operator's explicit confirmation that a live `make start` environment is
// up and may be probed. Without it the package runs no tests.
func TestMain(m *testing.M) {
	if os.Getenv("SPIKE_INTEGRATION_TEST") != "1" {
		fmt.Fprintln(os.Stderr,
			"integration: set SPIKE_INTEGRATION_TEST=1 to run the live "+
				"Pilot integration suite against a running `make start` "+
				"environment; skipping.")
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// spikeResult captures one `spike` invocation. stdout and stderr are kept
// separate on purpose: after the stdout fix, data output lands on stdout
// and diagnostics on stderr, and the smoke pass asserts on that split.
type spikeResult struct {
	stdout   string
	stderr   string
	exitCode int
	timedOut bool
}

// runSpike execs the `spike` binary (resolved from PATH, honoring the
// project's PATH-based harness convention) with the given args, optional
// stdin, and any extra environment. The timeout bounds the call so a
// command that hangs surfaces as timedOut rather than blocking the suite.
func runSpike(
	t *testing.T, timeout time.Duration, stdin []byte,
	extraEnv []string, args ...string,
) spikeResult {
	t.Helper()

	bin, lookErr := exec.LookPath("spike")
	if lookErr != nil {
		t.Fatalf("the 'spike' binary is not on PATH: %v "+
			"(run make build and put ./bin on PATH)", lookErr)
		return spikeResult{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Env = append(os.Environ(), extraEnv...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	_ = cmd.Run()

	res := spikeResult{
		stdout:   outBuf.String(),
		stderr:   errBuf.String(),
		timedOut: ctx.Err() == context.DeadlineExceeded,
	}
	if cmd.ProcessState != nil {
		res.exitCode = cmd.ProcessState.ExitCode()
	}
	return res
}

// processRunning reports whether a process with the exact given name is
// alive, via `pgrep -x` (exit 0 on a match).
func processRunning(name string) bool {
	return exec.Command("pgrep", "-x", name).Run() == nil
}

// requireHealthyEnv fails fast with a clear message when the live
// environment is not up, rather than letting each command fail obscurely.
func requireHealthyEnv(t *testing.T) {
	t.Helper()
	for _, proc := range []string{"nexus", "keeper", "spire-server"} {
		if !processRunning(proc) {
			t.Fatalf("%s is not running; start the environment with "+
				"make start before running the live integration suite",
				proc)
			return
		}
	}
}

// waitFor polls cond until it returns true or the timeout elapses.
func waitFor(timeout time.Duration, cond func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	return cond()
}

// repoRoot walks up from this test file to the module root (the directory
// holding go.mod). `go test` runs with the working directory set to the
// package directory, not the repository root, so repo-relative scripts such
// as the startup helpers must be addressed by absolute path.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine the test file path")
		return ""
	}
	dir := filepath.Dir(file)
	for {
		if _, statErr := os.Stat(
			filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find the repository root (go.mod)")
			return ""
		}
		dir = parent
	}
}

// TestPilotSmokePass drives the everyday CLI surface against a healthy
// environment: secret put/get/delete, policy create/get, and a cipher
// round trip. Data assertions target stdout, proving the stdout fix holds
// end to end. It cleans up the artifacts it creates.
func TestPilotSmokePass(t *testing.T) {
	requireHealthyEnv(t)

	suffix := fmt.Sprintf("%d-%d", os.Getpid(), time.Now().UnixNano())
	secretPath := "integration/smoke/" + suffix
	policyName := "integration-smoke-" + suffix
	password := "smoke-" + suffix

	// --- secret put, then get (value must land on stdout) ---
	put := runSpike(t, 30*time.Second, nil, nil,
		"secret", "put", secretPath,
		"username=spike", "password="+password)
	if put.timedOut || put.exitCode != 0 {
		t.Fatalf("secret put failed (exit %d, timedOut %v); stderr=%q",
			put.exitCode, put.timedOut, put.stderr)
		return
	}
	t.Cleanup(func() {
		_ = runSpike(t, 15*time.Second, nil, nil,
			"secret", "delete", secretPath)
	})

	get := runSpike(t, 30*time.Second, nil, nil,
		"secret", "get", secretPath)
	if get.timedOut {
		t.Fatalf("secret get timed out; stderr=%q", get.stderr)
		return
	}
	if !strings.Contains(get.stdout, password) {
		t.Fatalf("secret value not on stdout; stdout=%q stderr=%q",
			get.stdout, get.stderr)
		return
	}

	// --- secret delete, then confirm the value is gone from stdout ---
	del := runSpike(t, 30*time.Second, nil, nil,
		"secret", "delete", secretPath)
	if del.exitCode != 0 {
		t.Fatalf("secret delete failed (exit %d); stderr=%q",
			del.exitCode, del.stderr)
		return
	}
	gone := runSpike(t, 30*time.Second, nil, nil,
		"secret", "get", secretPath)
	if strings.Contains(gone.stdout, password) {
		t.Errorf("deleted secret value still on stdout; stdout=%q",
			gone.stdout)
	}

	// --- policy create, then get by name (name must land on stdout) ---
	create := runSpike(t, 30*time.Second, nil, nil,
		"policy", "create",
		"--name="+policyName,
		"--path-pattern=^integration/smoke/.*$",
		`--spiffeid-pattern=^spiffe://spike\.ist/workload/.*$`,
		"--permissions=read")
	if create.exitCode != 0 {
		t.Fatalf("policy create failed (exit %d); stderr=%q",
			create.exitCode, create.stderr)
		return
	}
	t.Cleanup(func() {
		_ = runSpike(t, 15*time.Second, nil, nil,
			"policy", "delete", policyName)
	})

	policyGet := runSpike(t, 30*time.Second, nil, nil,
		"policy", "get", policyName)
	if !strings.Contains(policyGet.stdout, policyName) {
		t.Fatalf("policy name not on stdout; stdout=%q stderr=%q",
			policyGet.stdout, policyGet.stderr)
		return
	}

	// --- cipher round trip: encrypt stdin, decrypt back to plaintext ---
	plaintext := []byte("integration cipher round trip " + suffix)
	enc := runSpike(t, 30*time.Second, plaintext, nil,
		"cipher", "encrypt")
	if enc.timedOut || enc.exitCode != 0 {
		t.Fatalf("cipher encrypt failed (exit %d, timedOut %v); stderr=%q",
			enc.exitCode, enc.timedOut, enc.stderr)
		return
	}
	dec := runSpike(t, 30*time.Second, []byte(enc.stdout), nil,
		"cipher", "decrypt")
	if dec.timedOut || dec.exitCode != 0 {
		t.Fatalf("cipher decrypt failed (exit %d, timedOut %v); stderr=%q",
			dec.exitCode, dec.timedOut, dec.stderr)
		return
	}
	if dec.stdout != string(plaintext) {
		t.Fatalf("cipher round trip mismatch: got %q want %q",
			dec.stdout, string(plaintext))
	}
}

// TestPilotWarnsWhenNexusUnreachable points a single invocation at a closed
// port, leaving the real Nexus untouched, and asserts the Pilot warns on
// stderr without hanging or panicking. The no-hang assertion also guards
// the still-open SVID-acquisition timeout task: a regression there would
// surface here as a timeout.
func TestPilotWarnsWhenNexusUnreachable(t *testing.T) {
	requireHealthyEnv(t)

	// Port 1 is reserved and refuses connections, so the dial fails fast
	// while the running Nexus keeps serving every other test.
	const timeout = 20 * time.Second
	res := runSpike(t, timeout, nil,
		[]string{"SPIKE_NEXUS_API_URL=https://127.0.0.1:1"},
		"secret", "get", "integration/unreachable-probe")

	if res.timedOut {
		t.Fatalf("the Pilot appears to hang when Nexus is unreachable "+
			"(no exit within %s); stderr=%q", timeout, res.stderr)
		return
	}

	combined := res.stdout + res.stderr
	if strings.Contains(combined, "panic:") ||
		strings.Contains(combined, "goroutine ") {
		t.Fatalf("the Pilot panicked when Nexus was unreachable; "+
			"stdout=%q stderr=%q", res.stdout, res.stderr)
		return
	}

	if !strings.Contains(res.stderr, "Failed to connect to SPIKE Nexus") {
		t.Fatalf("expected a connection-failure warning on stderr; "+
			"stdout=%q stderr=%q", res.stdout, res.stderr)
	}
}

// TestPilotDeniesWhenNexusUninitialized proves the Pilot denies operations
// against a reachable-but-uninitialized Nexus. That state is only reachable
// destructively: killing Nexus and every Keeper (so the in-memory shards
// are lost and auto-recovery is impossible), then restarting Nexus alone.
// The test is gated behind SPIKE_INTEGRATION_DESTRUCTIVE=1 and leaves the
// environment broken; reset it with `make kill && make start`.
func TestPilotDeniesWhenNexusUninitialized(t *testing.T) {
	if os.Getenv("SPIKE_INTEGRATION_DESTRUCTIVE") != "1" {
		t.Skip("destructive: set SPIKE_INTEGRATION_DESTRUCTIVE=1 to run. " +
			"It kills Nexus and every Keeper and restarts Nexus alone, " +
			"leaving it reachable but uninitialized. Reset afterward by " +
			"Ctrl+C on the make start terminal (or make kill), then " +
			"make start.")
		return
	}
	requireHealthyEnv(t)
	if os.Getenv("SPIKE_NEXUS_BACKEND_STORE") == "memory" {
		t.Skip("needs a persistent backend; unset SPIKE_NEXUS_BACKEND_STORE")
		return
	}

	// 1. Crash: kill Nexus and every Keeper. pkill exits non-zero when
	//    nothing matches, which is fine; only the wait below must succeed.
	_ = exec.Command("pkill", "-x", "nexus").Run()
	_ = exec.Command("pkill", "-x", "keeper").Run()
	crashed := waitFor(20*time.Second, func() bool {
		return !processRunning("nexus") && !processRunning("keeper")
	})
	if !crashed {
		t.Fatal("Nexus/Keeper processes did not exit after pkill")
		return
	}

	// 2. Restart Nexus alone. With no Keepers and no operator restore it
	//    comes up reachable over mTLS but without a root key: not ready.
	nexusLog, logErr := os.CreateTemp("", "spike-integration-nexus-*.log")
	if logErr != nil {
		t.Fatalf("could not create a Nexus log file: %v", logErr)
		return
	}
	root := repoRoot(t)
	startNexus := filepath.Join(
		root, "hack", "bare-metal", "startup", "start-nexus.sh")
	t.Logf("restarting Nexus alone (log: %s)", nexusLog.Name())
	restart := exec.Command(startNexus)
	// Run from the repository root, matching how the recovery drill invokes
	// the same script, so the restarted Nexus inherits the expected cwd.
	restart.Dir = root
	restart.Stdout = nexusLog
	restart.Stderr = nexusLog
	if startErr := restart.Start(); startErr != nil {
		t.Fatalf("could not restart Nexus alone: %v", startErr)
		return
	}
	// This Nexus was not launched through `make start`, so bg.sh's trap
	// (which reaps only the PIDs it recorded) does not know about it: left
	// alive, it would outlive a Ctrl+C on the make start terminal and hold
	// the Nexus port. Kill it when the test ends so the only survivors are
	// the tracked SPIRE processes, which Ctrl+C (or make kill) then cleans
	// up. `exec nexus` in start-nexus.sh means this PID is the Nexus itself.
	t.Cleanup(func() {
		if restart.Process != nil {
			_ = restart.Process.Kill()
			_ = restart.Wait()
		}
		t.Log("environment left uninitialized (Nexus and every Keeper " +
			"down; SPIRE still up). Reset with: Ctrl+C the make start " +
			"terminal (or make kill), then make start.")
	})

	// 3. Poll until Nexus is reachable and reports not-ready (the early
	//    probes may see a connection error while it is still starting),
	//    then assert the Pilot denies the read: no data on stdout.
	const deadline = 90 * time.Second
	var last spikeResult
	ready := waitFor(deadline, func() bool {
		last = runSpike(t, 10*time.Second, nil, nil,
			"secret", "get", "integration/uninitialized-probe")
		return strings.Contains(last.stderr, "not ready") ||
			strings.Contains(last.stderr, "not initialized")
	})
	if !ready {
		t.Fatalf("Nexus never reported not-ready within %s; last "+
			"stdout=%q stderr=%q (Nexus log: %s)",
			deadline, last.stdout, last.stderr, nexusLog.Name())
		return
	}
	if strings.TrimSpace(last.stdout) != "" {
		t.Errorf("expected no secret data on stdout when uninitialized; "+
			"stdout=%q", last.stdout)
	}
}
