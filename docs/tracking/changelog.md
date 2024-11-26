# SPIKE Changelog

## Recent

* Enabled policy-based access control.

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