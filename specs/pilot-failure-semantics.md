# Spec: SPIKE Pilot failure semantics

## Problem Statement

Every SPIKE Pilot command handler used `Run:` and reported failures by
printing to stderr and returning, so the process exited 0 on every error.
`spike secret get x || echo failed` never fired; a malformed `key=value`
pair in `secret put` was skipped and the command still printed OK. For a
secrets CLI driven by scripts, a wrong exit code is a correctness defect.

The entry point ended with `os.Exit(1)`, which CLAUDE.md forbids in favor
of the SDK fatal helper. The SDK logger, however, binds to `os.Stdout` on
first use and offers no way to choose a writer, so a fatal there would put
a JSON log line on the data stream that #299 had just cleaned up.

## Proposed Solution

1. Every command handler is `RunE` and returns an error; success returns
   nil. Error strings are Go-shaped (lowercase, no prefix); the entry point
   prints `Error: <message>` to stderr exactly once. The root command sets
   `SilenceErrors` and `SilenceUsage` so cobra does not print twice.
2. `stdout.HandleAPIError(c, err) bool` becomes `stdout.APIError(c, err)
   error`, returning the user-facing error; `flags.String` and `flags.Int`
   return `(value, *sdkErrors.SDKError)` instead of terminating.
3. Partial-success paths become failures: a malformed `key=value` pair, a
   missing key on `secret get`, a missing secret on `metadata get`, and a
   failed close on a cipher input or output file.
4. Exit: `Execute` terminates through `log.FatalLn`. Before anything can
   log, `main` calls `logfile.Route()`, which opens the well-known
   diagnostics log (`$HOME/.spike/pilot.log`, or
   `/tmp/.spike-$USER/pilot.log` without a home directory; 0600 inside
   0700) and initializes the SDK logger while `os.Stdout` temporarily
   points at that file. Command output on stdout stays clean; structured
   failure records go to the file. The user chose this over a documented
   `os.Exit` exception and over accepting JSON on stdout.
5. The dead CI harness (`ci/test/main.go`, `hack/test/test.sh`), which
   drove commands that no longer exist and hard-coded a home directory,
   is removed; `go mod tidy` drops `github.com/google/goexpect`.

## File Surface

- `app/spike/internal/cmd/**` (15 handlers to `RunE`), `cmd.go`, `root.go`
- `app/spike/internal/stdout/` (`APIError`), `app/spike/internal/cmd/flags/`
- `app/spike/internal/logfile/` (new: `doc.go`, `logfile.go`, test)
- `app/spike/cmd/main.go` (routes the logger first)
- `docs-src/content/usage/cli.md` ("Diagnostics Log" section)
- removed: `ci/test/`, `hack/test/test.sh`; `go.mod`, `go.sum` (goexpect)

## Error / Edge Cases

- **Logger already initialized.** `Route` cannot rebind a logger that was
  used earlier; it is the first statement in `main`, and the package doc
  says so. If the log file cannot be prepared, `main` writes the reason to
  stderr and terminates through `log.FatalErr`; in that single corner the
  record lands on stdout because there is nowhere else.
- **Relies on the SDK binding lazily.** Documented in the package; the
  durable fix is a `log.SetOutput` in spike-sdk-go, tracked in TASKS.md.
- **Message case.** The first letter of user-facing errors is now
  lowercase after `Error: `; wording is otherwise unchanged.

## Non-Goals

- No change to success-path stdout bytes (tests compare them).
- No new environment variable for the log location.

## Verification

- Unit tests drive `spike decrypt --version 999` and the encrypt
  counterpart through cobra with a silenced root and assert a returned
  error with nothing on stdout or stderr; `logfile` tests assert the path
  resolution, the 0600 mode, and that a logged record lands in the file
  while `os.Stdout` is restored.
- `grep -rn 'os\.Exit(1)' --include='*.go' .` prints nothing.
- `make test` (race, uncached) and `make audit` exit 0 (numbers recorded
  in `specs/lint-hardening.md`).
