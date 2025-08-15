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

* SPIKE Nexus, SPIKE Pilot, and SPIKE Keeper can now be built as Windows
  binaries too.
* SPIKE and SPIKE Go SDK code has been refactored to better align with common
  Go idioms and conventions.
* Added stricter linting.
* Added vulnerability checks to SPIKE and SPIKE Go SDK.
* enabled `GOFIPS140=v1.0.0`, the modern way of enabling FIPS. We 
  are not using `boringcrypto` anymore.
* Separated bootstrap logic into its own app to enable a more deterministic
  initialization flow. This change will also unlock the ability to run SPIKE
  Nexus in HA mode.
* **BREAKING**: SPIKE now requires a separate initializer to begin its lifecycle.
* FIPS 140.3 is enabled and enforced everywhere.

## [0.4.2] - 2025-07-19

### Added

* Ability to configure to not how SPIKE banner on startup.
* Ability to configure to show a warning if memory locking is not
  available on the system.
* SPIKE can now be deployed from SPIFFE helm charts. Tested and verified!
* Documentation updates.
* SPIKE can be now be installed from [SPIFFE Helm 
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
