+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.

title = "Changelog"
weight = 2
sort_by = "weight"
+++

# SPIKE Changelog

## Recent

* Added configurable retry backoff intervals for SPIKE Bootstrap keeper
  communication. New environment variables `SPIKE_BOOTSTRAP_KEEPER_RETRY_INITIAL_INTERVAL`
  (default 2s) and `SPIKE_BOOTSTRAP_KEEPER_RETRY_MAX_INTERVAL` (default 30s)
  allow operators to tune retry behavior during bootstrap.
* SDK: `retry.WithMaxAttempts` now accepts optional `RetrierOption` parameters,
  enabling callers to customize backoff settings while maintaining backward
  compatibility.
* Update documentation to reflect the new SPIKE architecture.
* Fix occasional dangling process issues when `make start` does not exit cleanly.
* SDK API methods now return cloned versions of sentinel *SDKErrors instead of
  returning the original reference. This prevents accidental mutation of the
  error values.
* mem.Lock() does not print JSON error logs on CLI startup anymore.
* moved some of the internal reusable feature from in-tree to SPIKE Go SDK.
* binaries are now create at the ./bin folder instead of the root of the project.
* log files are now created at the ./logs folder instead of the root of the project.
* factored out some common validation and error handling logic from in-tree to the SDK.

## [0.8.0] - 2025-11-28

### Added 

* Additional and comprehensive logging to all SPIKE Nexus and SPIKE Keeper API
  methods.
* Better error handling across the entire codebase.
* Pilot: Reduced CLI verbosity by removing structured JSON log output from 
  all commands (policy, secret, cipher, operator). The CLI now outputs clean,
  concise error messages to stderr without internal debug logs cluttering the
  terminal.
* "Encryption as a service" support for SPIKE Pilot. There is an outstanding
  issue for JSON mode; however, streaming mode works as expected.
* `make start` includes additional smoke tests to ensure all SPIKE components are
  in good shape and ready to roll.
* Added extensive package documentation to ALL packages of SPIKE and 
  SPIKE Go SGK.
* SDK: Improved documentation clarity for single return value functions, CSPRNG
  fatal behavior, and function distinctions (ValidatePath vs. 
  ValidatePathPattern).
* SDK: Significantly increased test coverage across all SDK packages with
  comprehensive unit and integration tests.
* SDK: Enhanced documentation for the version numbering system---version numbers
  start at 1, and `CurrentVersion == 0` indicates all versions have been deleted.
* SDK: Updated `Delete()` documentation to clarify soft-delete behavior and that
  paths remain in storage even when all versions are deleted.
* SDK: Added `HasValidVersions()` and `Empty()` helper methods to `kv.Value` for
  checking if secrets have any non-deleted versions, useful for identifying
  purgeable secrets.
* SDK: Added `Destroy()` method to `kv.KV` for hard-delete operations that
  permanently remove secret paths from storage and reclaim memory. Unlike
  soft-delete (`Delete()`), this cannot be undone.
* Nexus: Comprehensive documentation updates across ALL files ensuring
  consistency between function signatures, parameter types, return values, and
  actual code behavior. Updated error type references from generic `error` to
  specific `*sdkErrors.SDKError` types.
* Nexus: Added defensive nil source checks across concurrent/distributed systems
  where workload API can asynchronously invalidate X509Source. Updated
  `InitializeBackingStoreFromKeepers`, `SendShardsPeriodically`, CLI commands,
  and server startup with proper nil handling and documentation explaining
  retry behavior for transient failures.
* Nexus, Keeper: Added AST-based tests to enforce guard function usage in all
  route handlers. The tests scan route handler files and verify each `Route*`
  function calls either `net.ReadParseAndGuard` or a guard function directly.
  This prevents contributors from accidentally adding routes without
  authorization checks. See ADR-0031.

### Changed

* **BREAKING**: SDK now returns typed sentinel errors instead of generic `error`
  values.
* **BREAKING**: SDK: Enhanced error handling---Get methods now return 
  `ErrAPINotFound` instead of `(nil, nil)` when resources are not found, 
  following idiomatic Go patterns (similar to `os.Open`, `database/sql`).
* SDK: Improved API consistency by standardizing policy function 
  parameters from `name` to `id` across all operations, matching internal 
  implementation.
* Nexus: Enhanced backend interface documentation with proper parameter and
  return type information, and documented `CurrentVersion == 0` behavior in
  `LoadSecret` and `LoadAllSecrets` methods.
* Nexus: Comprehensive documentation updates for all secret management functions
  with accurate parameter names, return types, and behavioral details including
  soft-delete semantics and metadata update logic.
* Nexus: Made `DeleteSecret` more defensive when finding the new current version 
  by removing unnecessary condition, improving code clarity and robustness.
* **BREAKING**: Nexus: Fixed inconsistent error returns in memory backend - 
  `LoadSecret` now returns `ErrEntityNotFound` instead of `(nil, nil)` for 
  missing secrets.
* Nexus: Optimized retry loop in `InitializeBackingStoreFromKeepers` with early
  nil check to avoid unnecessary function call overhead when X509 source is nil.
* Nexus: Refactored `ShardGetResponse` to return `([]byte, *sdkErrors.SDKError)`
  instead of logging errors internally and returning empty slices, following
  canonical Go error handling patterns.
* Nexus: Improved resilience in data loading functions (`LoadAllPolicies`,
  `LoadAllSecrets`) by changing from aggressive exit behavior to graceful
  degradation - now logs warnings and continues processing valid entries instead
  of abandoning entire dataset on single entry corruption.
* Pilot: Comprehensive refactoring of CLI output handling across all commands
  (14 files) to use Cobra's `cmd.Print*()` methods instead of `fmt.Print*()`.
  Error messages now properly route to stderr via `cmd.PrintErrln()`/
  `cmd.PrintErrf()`, while success and normal output routes to stdout via
  `cmd.Println()`/`cmd.Printf()`. This improves testability, respects Cobra's
  output configuration, and provides proper stderr/stdout separation. Updated
  helper functions `printSecretResponse()` and `handleAPIError()` to accept
  cmd parameter for consistent output handling.
* SDK: Added `UpdatedAt` field to `Policy` struct to track when policies are
  modified. Removed unused `CreatedBy` field.
* Nexus: Standardized error handling across recovery modules to use
  `log.WarnErr`/`log.FatalErr` with SDK error types instead of generic
  `log.Warn`/`log.FatalLn` calls. This provides searchable error codes and
  consistent error patterns.
* **BREAKING**: Nexus: Changed policy operations from create-only to upsert 
  semantics for consistency with secret operations. `state.CreatePolicy` is now
  `state.UpsertPolicy`. If a policy with the same name exists, it is updated
  (preserving ID and CreatedAt); otherwise, a new policy is created.
* Code Quality: Eliminated error variable shadowing across the codebase. Error
  variables now use descriptive names (`atoiErr`, `nonceErr`, `openErr`,
  `restoreErr`, etc.) instead of reusing `err`. This prevents subtle bugs where
  a later error could inadvertently shadow an earlier one and improves code
  readability by making error sources explicit.

### Fixed 

* Finally, fixed the flaky tests around the retry logic in SPIKE Go SDK for 
  good.
* Various other bugfixes, refactorings, and security improvements.
* SDK: Added nil validation to `CreateMTLSServer` functions with fail-fast 
  behavior for configuration errors.
* SDK: Fixed resource management bug in `StreamPostWithContentType` where defer
  was closing response body on the success path, causing callers to receive closed 
  body.
* SDK: Fixed critical bug in `Undelete` function that was ignoring the `versions`
  parameter due to missing else clause.
* Nexus: Added `OldestVersion` tracking to `UndeleteSecret` for consistency
  with `DeleteSecret`, ensuring metadata accurately reflects the oldest 
  non-deleted version.
* Nexus: Fixed bug in `UndeleteSecret` where undeleting a version higher than
  the current `CurrentVersion` did not update `CurrentVersion` to reflect the
  new highest active version, causing metadata inconsistency.
* Nexus: Fixed critical bug in `UpsertSecret` where adding a new version when all
  existing versions were deleted (CurrentVersion == 0) would create version 1,
  potentially colliding with an existing deleted version 1. Now correctly finds
  the highest existing version number and increments from there.
* Nexus: Fixed resource leak in `internal/net/post.go` where response body
  close was deferred after body read instead of immediately after response
  obtained, causing leaks when read operations failed.
* Nexus: Fixed a critical bug in secret route handlers where error paths were not
  sending HTTP responses to clients. Added missing `net.Fail()` calls in
  `put_intercept.go` (3 locations) and `undelete.go` to ensure proper error
  responses.
* Nexus: Fixed bug in `RouteDeletePolicy` that returned HTTP 500 for all errors
  including "not found." Now correctly returns HTTP 404 when the policy does not
  exist

### Security

* PoP validation after the bootstrap sequence to ensure SPIKE Nexus has 
  initialized properly.
* Update SPIKE Components' Go version to `1.25.3`.
* `log.FatalLn` exits cleanly by default to avoid leaking sensitive information
  via stack traces in production. Stack traces can be enabled for
  development/testing by setting `SPIKE_STACK_TRACES_ON_LOG_FATAL=true`.
* SDK upgrade to Go 1.25.3 to fix `GO-2025-4007`.
* Fixed error handling inconsistency in `NewPilotRecoveryShards` to
  ensure fail-fast behavior on shard generation failures. The function now
  consistently uses `log.FatalLn` for all critical errors during shard
  marshaling to prevent silent generation of corrupted recovery material.
* Added SPIFFE ID validation to SPIKE Keeper shard endpoints.
  The `RouteShard` endpoint now validates that only SPIKE Nexus can retrieve
  shards during recovery operations. The `RouteContribute` endpoint validates
  that only SPIKE Bootstrap (during initial setup) or SPIKE Nexus (during
  periodic updates) can contribute shards. This prevents unauthorized access
  to sensitive shard data.
* Crypto: Consolidated GCM nonce size constant (`crypto.GCMNonceSize`) to
  `internal/crypto/gcm.go`. This removes duplication across cipher and bootstrap
  packages and documents the decision to use the NIST-recommended 12-byte
  standard. See ADR-0032.
* Fixed [`CWE-117`: go-viper's mapstructure May Leak Sensitive Information in 
Logs When Processing Malformed 
Data](https://github.com/spiffe/spike/security/dependabot/7)
* Fixed [`CVE-2025-58181`: golang.org/x/crypto/ssh allows an attacker to cause 
unbounded memory 
consumption](https://github.com/spiffe/spike/security/dependabot/9)
* Fixed [`CVE-2025-47914`: golang.org/x/crypto/ssh/agent vulnerable to panic if
message is malformed due to out of bounds 
read](https://github.com/spiffe/spike/security/dependabot/9)

## [0.6.1] - 2025-10-02

This is a patch release to align with the changes in the upstream helm charts.

## [0.6.0] - 2025-10-01

This was a security release where the main focus was hardening SPIKE SDK mTLS
implementation. In addition, we created a configurable SPIKE backing store 
directory to enable future HA development.

### Added

* Added `SPIKE_TRUST_ROOT_BOOTSTRAP` to enable SPIKE Bootstrap to be used
  in different trust boundaries.
* Added `SPIKE_NEXUS_DATA_DIR` to enable setting up custom data directories for
  SPIKE Nexus backing store.
* Added convenience methods to the SPIKE Go SDK.

### Changed

* Improvements to the SPIKE Go SDK.
* Stricter SPIFFE ID validation. SPIKE SDK now ensures that the API client
  only talks to SPIKE Nexus as the server.

### Fixed

* Minor bug fixes.
* Fixed flaky unit tests.

### Security 

* SPIKE Go SDK clients are hardened to only talk to SPIKE Nexus as the
  server during mTLS.

## [0.5.1] - 2025-09-14

## Changed

* Updated SPIKE Bootstrap to be more robust by adding exponential backoff while
  waiting for SPIKE Keepers to be ready.
* Enhancements in startup scripts to better enable local development with
  SPIFFE Helm Charts that have not been published yet.

## [0.5.0] - 2025-09-11

This is still a **prerelease** version; however, it includes major changes
and improvements. We will cut a stable release once we have **SPIKE Bootstrap**
included in the SPIFFE Helm Charts.

### Added

* Updates to documentation and usage examples.
* Updates to the SPIKE Go SDK around the logging API.
* Moved certain reusable features from in-tree to SPIKE Go SDK.
* A new `make audit` target that helps contributors run style checks and
  linters before submitting a PR.
* Enhancements to bare-metal installation scripts.

### Changed

* Updated Go version to `1.25.1`
* Updated **SPIKE Bootstrap** to be more robust and enabled it to work on
  Kubernetes too.
* Clarified documentation around path pattern and SPIFFE ID pattern matching
  in SPIKE policies.
* Slight improvements in the SPIKE logo and a brand-new landing page that
  highlights the project's vision and goals.
* Moved environment variable names to the SPIKE Go SDK as constants to prevent
  typos and to make it easier to use the SDK.

### Fixed

* Bug fixes and stability improvements.
* Fixed failing unit tests on CI (that's a temporary fix that runs tests
  sequentially instead of in parallel; we will fix that soon)

### Security

* Along with secrets, SPIKE Nexus now encrypts policies at rest too.

### Upcoming

* A lot of ongoing design work around key rotation, encryption, and a secure
  web interface that leverages Web Cryptography API to provide a secure
  experience of managing secrets without having to interact with the command
  line.
* Ongoing work on the **Cipher** API to provide "encryption as a service" to
  systems and workloads that do not require to store secrets in a backing store.

## [0.4.3] - 2025-08-16 (*prerelease*)

This is a "*prerelease*" version to enable upstream SPIFFE Helm Charts
integration initiatives. The most significant change is the introduction of a 
[**SPIKE Bootstrap** app][bootstrap] that is responsible for initializing 
**SPIKE Nexus**. This new approach separates the bootstrapping workflow that 
had been inside **SPIKE Nexus**' initialization workflow before. And that
enables us an opportunity to run **SPIKE Nexus** in HA mode without designing
elaborate, and potentially error-prone, consensus algorithms.

[bootstrap]: https://github.com/spiffe/spike/blob/main/app/bootstrap/README.md "SPIKE Bootstrap"

### Added

* **FIPS 140.3 Compliance**: FIPS is now enabled at **build time**, and it's 
  enforced everywhere. We are using `GOFIPS140=v1.0.0`, the modern way of 
  enabling FIPS, retiring our older `boringcrypto` implementation.
* `spike policy list` command can now filter by SPIFFE ID pattern and path
  pattern.
* `spike policy` command cano now accept a YAML file as input, instead of
  requiring command-line parameters.
* SPIKE Go SDK now has a generator that creates pattern-based, secure, 
  randomized secrets.
* Implemented a (currently experimental) "SPIKE Lite" mode where SPIKE Nexus
  would not need a backing store, or policies, and can leverage the storage
  and policy mechanism of S3-compatible object stores (such as Minio). Once
  we fully implement and polish SPIKE Lite, we will also update documentation
  and use cases to allow users to understand the benefits and liabilities of 
  SPIKE Lite and why they might want to use one over the other.

### Changed

* Better alignment with idiomatic Go practices. SPIKE and SPIKE Go SDK code 
  has been refactored to better align with common Go idioms and conventions.
  We also created a `make audit` target to run style checks and linters that
  enforce a consistent code style and some of these guidelines. `make audit`
  is also a part of the CI pipeline to ensure that the code is always compliant
  at every commit. In addition `make audit` also does vulnerability checks.
* **BREAKING**: SPIKE Nexus now requires a separate initializer (SPIKE Bootstrap)
  to begin its lifecycle. The user guides and relevant documentation have been
  updated to reflect this change.
* Updated Go to the latest version (`1.24.6`).

### Fixed

* Fixed a bug related to Windows builds. SPIKE Nexus, SPIKE Pilot, and SPIKE 
  Keeper can now be built as Windows binaries too.
* Various refactorings, improvements, code cleanup, and bug fixes.

## [0.4.2] - 2025-07-19

### Added

* Ability to configure to not how SPIKE banner on startup.
* Ability to configure to show a warning if memory locking is not
  available on the system.
* SPIKE can now be deployed from SPIFFE helm charts. Tested and verified!
* Documentation updates.
* SPIKE can now be installed from [SPIFFE Helm 
  Charts](https://github.com/spiffe/helm-charts-hardened) and can 
  [federate secrets across clusters](https://vimeo.com/v0lkan/spike-federation)

### Changed

* Moved logging to SPIKE SDK. VSecM v2 will share the same logging setup.
* `spike policy` command now accepts file input; you can design your policies
  as `yaml` files and then `spike policy apply -f` them.

### Security

* Fixed [`GHSA-fv92-fjc5-jj9h`: `mapstructure` May Leak Sensitive Information 
  in Logs When Processing Malformed 
  Data](https://github.com/spiffe/spike/security/dependabot/6)

## [0.4.1] - 2025-06-01 (*prerelease*)

### Added

* Initial support for Kubernetes deployments.
* Better shard sanitization during recovery procedures.
* Added memory locking to SPIKE Pilot too.
* Finer control of the startup script via flags.
* Added the ability to optionally skip database schema creation during SPIKE
  initialization.

### Changed

* **BREAKING**: SDK validation methods now take trust root as an argument.
* **BREAKING**: `SPIKE_NEXUS_KEEPER_URL` is now a comma-delimited list of URLs
  (instead of JSON).
* SPIKE components can now be configured to accept multiple trust roots as
  legitimate peers---this will be useful in complex mesh and federation
  deployment scenarios.
* SPIKE now uses GitHub Container Registry to store its container image
  (instead of Docker Hub).

### Fixed

* Fixed a bug where the doomsday recovery procedure was not immediately 
  restoring the data.

## [0.4.0] - 2025-04-16

### Added

* Added more configuration options to SPIKE Nexus.
* Updated documentation around security and production hardening.
* Updated release instructions, added a series of tests to follow and cutting
  a release only after all tests pass. These tests are manual for now but
  can be automated later down the line.

### Fixed

* Fixed a bug related to policies not recovering after a SPIKE Nexus crash.
  Now, both secrets and policies recover without an issue.
* Ensured that "in memory" mode works as expected, and we can create policies
  and secrets.
* Fixed inconsistencies in the audit log format.
* Fixed NilPointer exception during certain shard creation paths.
* Fixed regressions due to premature memory cleanup. Now the memory is cleaned
  up when no longer needed (but not before).
* Various bug fixes and improvements.

### Changed

* Moved some common reusable code to `spike-sdk-go`.
* Various changes and improvements in SPIKE Go SDK.
* The startup script does not initiate SPIKE Keepers if SPIKE is running in
  "in memory" mode.
* Renamed `AuditCreated` enum as `AuditEntryCreated` to specify its intention
  better (i.e., it's not creation of an entity or a DAO, but rather it's
  the start of an audit trail).
* Improved `spike policy` commands with better UX and error handling.

### Security

* Added cache invalidation headers to all API responses.
* For added security, we strip symbols during the build process now.
* Implemented better memory protection with cleaning up memory when no longer needed.
* SPIKE Nexus and SPIKE Keepers use `mlock` to avoid memory swapping when possible.
* [Fixed `CVE-2025-22872`: golang.org/x/net vulnerable to Cross-site Scripting](https://github.com/spiffe/spike/security/dependabot/5)
* [Fixed `CVE-2025-22870`: HTTP Proxy bypass using IPv6 Zone IDs in golang.org/x/net](https://github.com/spiffe/spike/security/dependabot/4)


## [0.3.1] - 2025-03-04

### Added

* SPIKE Nexus now accepts a dynamic number of SPIKE Keepers and Shamir share
  threshold (defaults to 3 keepers, and minimum 2 shares (out of 3) to
  recreate the root key).
* Started containerization work (created a Dockerfile); yet it's far from
  complete: We will work on that.
* Various documentation updates.
* Minor bug fixes in initialization scripts.

### Changed

* Secrets now rehydrate from the backing store immediately after SPIKE
  Nexus crashes. Former implementation was using an optimistic algorithm
  (i.e., do not load the secret unless you need it), yet that was causing
  calls to `spike secret list` return an empty collection. This implementation
  fixes that issue and also ensures that SPIKE Nexus' memory continues to
  be the primary source of truth (by design).

### Security

* SPIKE Nexus now securely erases the old root key and shards from memory after
 it is no longer necessary. Before, it was left to the garbage collector to 
 handle that. The current approach is NIST recommendation and provides better
 memory protection.
* [Fixed `CVE-2025-271447`: DoS in go-jose Parsing](https://github.com/spiffe/spike/security/dependabot/3)

## [0.3.0] - 2025-02-20

This release was focused around bugfixes, stability, documentation, and 
disaster recovery.

### Added

* Documentation: SPIKE Production Hardening Guide is complete and ready for
  consumption (*it was in draft mode before*).
* Implemented `spike operator recover` and `spike operator restore` commands 
  that provide disaster recovery capabilities if there is a total system crash 
  and the remaining SPIKE Keepers are less than the threshold to recover the 
  root key.
* Several bugfixes and performance improvements.
* Added a coverage report to the repository. The coverage is not as high as
  we would like to be; yet we have to start somewhere :).
* Added several architectural decision records to share the projects' vision
  and design decisions transparently.
* Started working on containerization (*though it's still a work in progress*).

### Changed

* SPIKE Website has undergone a major overhaul.
* Documentation updates, especially around security and disaster recovery.
* Documentation is now consistent with the code: Removed outdated sections,
  introduced new modules, explained current workflows and state transitions.
* Moved documentation from Docsify to Zola, that gave, speed, flexibility,
  templateability, and consistency to the overall documentation.
* Significant updates in [SPIKE go SDK](https://github.com/spiffe/spike-sdk-go).

## [0.2.1] - 2025-01-23

### Added

* Enabled policy-based access control.
* The root key that SPIKE Nexus generates is now split into several Shamir
  shards and distributed to SPIKE Keepers.
* New additions and improvements to SPIKE Go SDK.
* Various minor bugfixes.
* Code cleanup.
* Implemented several recovery scenarios.
* SPIKE now has static analysis, CI integration, linting, and automated tests.
* Documentation updates. Documentation is still lagging behind, but we are
  updating and improving it along the way.
* Created a makefile to group related scripts into make targets.
* Made the start script more robust.
* Ensured that the policies and the demo app work as expected.
* Implemented a Secret Metadata API.
* Implemented exponential retries across several API-consuming methods.


### Changed

* **BREAKING**: changed the CLI usage. Instead of `spike get`, for example, we
  now use `spike secret get`. The reason for this change is that we introduced
  a `policy` command (i.e. `spike policy get`).

### Security

* [Fixed `CVE-2024-45337`: Misuse of ServerConfig.PublicKeyCallback may cause
  authorization bypass in golang.org/x/crypto](https://github.com/spiffe/spike/security/dependabot/1)
* [Fixed `CVE-2024-45338`: Non-linear parsing of case-insensitive content in
  `golang.org/x/net/htm`](https://github.com/spiffe/spike/security/dependabot/2)

## [0.2.0] - 2024-11-22

### Added

* Added configuration options for SPIKE Nexus and SPIKE Keeper.
* Documentation updates.
* Max secret version is now configurable.
* Introduced standard and configurable logging.
* Added sqlite3 as a backing store.
* Enabled cross-compilation and SHA checksums.
* Enhanced audit trails and error logging.
* Created initial smoke/integration tests.
* Stability improvements.

### Changed

* Removed password authentication for admin users. Admin users' SVIDs
  are good enough to authenticate them.
* Implemented passwordless admin login flow
  (*the neat thing about passwords is: you don't need them*).

## [0.1.0] - 2024-11-06

### Added

* Implemented `put`, `read`, `delete`, `undelete`, and `list` functionalities.
* Created initial documentation, README, and related files.
* Compiled binaries targeting various platforms (x86, arm64, darwin, linux).
* SPIKE is demoable; however, we need to update certain login and initialization
  flows.
* In-memory secrets storage only (*using database as a backing store is coming up
  next*)
* Created a `jira.txt` to track things (*to avoid polluting GitHub issues
  unnecessarily*)
* This is an amazing start; more will come. Turtle power 🐢⚡️.

<p>&nbsp;</p>

----

{{ toc_tracking() }}

----

{{ toc_top() }}
