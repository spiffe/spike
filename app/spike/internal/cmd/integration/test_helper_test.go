//go:build integration

//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// runSpike execs the `spike` binary (resolved from PATH, honoring the
// project's PATH-based harness convention) with the given args, optional
// stdin, and any extra environment. The timeout bounds the call so a
// command that hangs surfaces as timedOut rather than blocking the suite.
//
// Parameters:
//   - t: The test, failed when the binary is not on PATH.
//   - timeout: The maximum time the invocation may take.
//   - stdin: Bytes fed to the command's standard input, or nil for none.
//   - extraEnv: Additional environment entries appended to the parent's.
//   - args: The command-line arguments passed to spike.
//
// Returns:
//   - spikeResult: The captured stdout, stderr, exit code, and whether the
//     timeout elapsed.
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

	// A non-zero exit is an expected outcome that the caller inspects
	// through exitCode. Anything else means the binary could not run at
	// all, which no assertion downstream can explain.
	if runErr := cmd.Run(); runErr != nil {
		var exitErr *exec.ExitError
		if !errors.As(runErr, &exitErr) {
			t.Fatalf("could not run spike: %v", runErr)
			return spikeResult{}
		}
	}

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
//
// Parameters:
//   - name: The exact process name to look for.
//
// Returns:
//   - bool: True when at least one matching process is running.
func processRunning(name string) bool {
	return exec.Command("pgrep", "-x", name).Run() == nil
}

// requireHealthyEnv fails fast with a clear message when the live
// environment is not up, rather than letting each command fail obscurely.
//
// Parameters:
//   - t: The test, failed when Nexus, a Keeper, or SPIRE Server is down.
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
//
// Parameters:
//   - timeout: How long to keep polling.
//   - cond: The condition to evaluate every half second.
//
// Returns:
//   - bool: The final value of cond, true when it was met in time.
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

// repoRoot walks up from this helper file to the module root (the directory
// holding go.mod). `go test` runs with the working directory set to the
// package directory, not the repository root, so repo-relative scripts such
// as the startup helpers must be addressed by absolute path.
//
// Parameters:
//   - t: The test, failed when the root cannot be determined.
//
// Returns:
//   - string: The absolute path of the repository root.
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

// killProcess sends SIGTERM to every process with the given name through
// pkill. A non-zero pkill exit means nothing matched, which is an accepted
// outcome; any other failure means pkill itself could not run.
//
// Parameters:
//   - t: The test, failed when pkill cannot be executed.
//   - name: The exact process name to match.
func killProcess(t *testing.T, name string) {
	t.Helper()

	runErr := exec.Command("pkill", "-x", name).Run()
	var exitErr *exec.ExitError
	if runErr != nil && !errors.As(runErr, &exitErr) {
		t.Fatalf("could not run pkill for %s: %v", name, runErr)
	}
}

// stopProcess kills a process this test started and reaps it. A process
// that already exited is fine; a kill that fails for any other reason is
// reported, because a lingering Nexus would hold its port after the test.
//
// Parameters:
//   - t: The test, marked failed when the process cannot be killed.
//   - cmd: The started command to stop.
func stopProcess(t *testing.T, cmd *exec.Cmd) {
	t.Helper()

	if killErr := cmd.Process.Kill(); killErr != nil &&
		!errors.Is(killErr, os.ErrProcessDone) {
		t.Errorf("could not kill the restarted Nexus: %v", killErr)
	}
	// Wait reports the kill signal as an ExitError, which is expected.
	var exitErr *exec.ExitError
	if waitErr := cmd.Wait(); waitErr != nil &&
		!errors.As(waitErr, &exitErr) {
		t.Errorf("could not reap the restarted Nexus: %v", waitErr)
	}
}
