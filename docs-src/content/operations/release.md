+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.

title = "SPIKE Relase Management"
weight = 5
sort_by = "weight"
+++

# SPIKE Release Management

This document provides an overview of how the cut a **SPIKE** release, including
testing guidelines and instructions.

> **We Still Use Human Push-buttons**
> 
> Although some steps, audits, and integration tests of the release
> process are automated, we still follow several manual steps
> outlined in this document.


Below, you will find detailed instructions and examples to guide contributors
through the release and testing process.

This document is targeted for **core contributors** who are responsible for
managing the release cuts of **SPIKE**. It provides detailed instructions to
ensure a smooth and reliable release process.

## Coverage Report

The coverage report is available at [https://spike.ist/coverage.html][coverage].

[coverage]: https://spike.ist/coverage.html "SPIKE Coverage Report"

We update the coverage report at every release cut.

If you want to increase test coverage, you are more than welcome to contribute
to the project.

## Before Every Release

Before every release:

1.  Run the unit tests: `make test`.
2.  Run `make start` and verify you see the message "Everything is set up."
    to confirm the smoke tests pass, then press `Ctrl+C` to stop.
3.  Switch to "in-memory" mode, run `make start` and verify you see the message
    "Everything is set up." again to confirm the smoke tests pass in that mode
    too, then press `Ctrl+C` to stop.
4.  Run `make audit` to ensure the project is free of security vulnerabilities.
5.  If everything passes, update `./app/VERSION.txt` to the release version.
6.  Update any necessary documentation.
7.  Update the changelog
    (`docs-src/content/tracking/changelog.md`). 
8.  Run `make docs` to generate and publish the documentation, including the
    coverage report.

Release process:

* Merge all the changes to the `main` branch.
* Tag a version by running `make tag` (*this creates a GPG-signed tag using the 
  version from `app/VERSION.txt` and pushes it to origin*).
* Convert the tag to a **release** on **GitHub**.
* Copy the current version's changelog over to the release notes on **GitHub**.
* On a Mac machine [follow cross-platform build instructions][cross-platform]
  to generate binaries.
* Add binaries to the release as assets.
* Announce the release in relevant channels.
* You are all set.

[cross-platform]: @/operations/build.md "SPIKE Cross-Platform Build"

<p>&nbsp;</p>

----

{{ toc_operations() }}

----

{{ toc_top() }}
