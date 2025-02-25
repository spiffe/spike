+++
# //    \\ SPIKE: Secure your secrets with SPIFFE.
# //  \\\\\ Copyright 2024-present SPIKE contributors.
# // \\\\\\\ SPDX-License-Identifier: Apache-2.

title = "Changelog"
weight = 2
sort_by = "weight"
+++

# SPIKE Changelog

## Recent

TBD

## [0.3.0] - 2026-02-20

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
* Added coverage report to the repository. The coverage is not as high as
  we would like to be; yet we have to start somewhere :).
* Added several architectural decision records to share the projects vision
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

## [0.2.1] - 2026-01-23

### Added

* Enabled policy-based access control.
* The root key that SPIKE Nexus generates is now split into several Shamir
  shards and distribute to SPIKE Keepers.
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
* Max secret versions is now configurable.
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
* SPIKE is demoable, however we need to update certain login and initialization
  flows.
* In memory secrets storage only (*using database as a backing store is coming up
  next*)
* Created a `jira.txt` to track things (*to avoid polluting GitHub issues
  unnecessarily*)
* This is an amazing start; more will come. Turtle power 🐢⚡️.

----

{{ toc_tracking() }}

----

{{ toc_top() }}
