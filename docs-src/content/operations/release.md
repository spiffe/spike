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

> **We Use Humans as Push-buttons**
> 
> At the moment, we don't have a CI pipeline in place for the release
> process. Most of the operations mentioned here are manual. However, we are
> actively working on improving and automating the release process. This 
> document will be updated as we introduce more automation into the pipeline.

Below, you will find detailed instructions and examples to guide contributors
through the release and testing process.

This document is targeted for **core contributors** who are responsible for
managing the release cuts of **SPIKE**. It provides detailed instructions to
ensure a smooth and reliable release process.

## Before Every Release

Before every release:

1. Run the unit tests
2. Run the following smoke tests documented in the next section.
3. If everything passes, update `NexusVersion`, `PilotVersion` 
   and `KeeperVersion` in `$WORKSPACE/spike/internal/config/config.go`
4. Update any necessary documentation.
5. Update the changelog.
6. Update the documentation snapshots page.
7. Run `./hack/cover.sh` to update and send the coverage report to the public 
   docs.
8. Make sure you update `./app/VERSION.txt` with the new version.

Release process:

* Publish documentation by running `zola build` in `./docs-src` and then
  copying the generated HTML in `./docs-src/public` into `/.docs`.
* Merge all the changes to the `main` branch.
* Tag a version and convert that version to a **release* on **GitHub**.
  * Make sure you GPG sign your tag.
* Copy the current version's changelog over to the release notes on **GitHub**.
* On a Mac machine [follow cross-platform build instructions][cross-platform]
  to generate binaries.
* Add binaries to the release as assets.
* Announce the release in relevant channels.
* You are all set.

[cross-platform]: @/operations/build.md "SPIKE Cross-Platform Build"

## SPIKE Smoke Tests

> **How About Automation**?
>
> There is a partially-working `./ci/test/main.go` binary that you can play
> with. But unless that is fixed, we'll have to run **SPIKE** smoke test
> manually.
>
> `./ci/test/main.go` is designed for integration tests, and it assumes a
> working **SPIKE** and **SPIKE** environment to execute properly.

Note that these are instructions for a manual smoke test, and it's not a
replacement for a full integration test. We may add more steps, but we'll
keep it lightweight---Passing the smoke test means that the core components
and the features of the system are reliably functional.

Here is a list of manual tests that can be done before every release:

1. Reset the test bed.
2. Switch to "in-memory" mode.
3. Make sure you can create a secret and read it back.
4. Make sure you can create a policy and read it back.
5. Make sure you can list secrets.
6. Make sure you can list policies.
7. Make sure the demo workload reads and writes based on the
   policies you created.
8. Reset the entire test bed.
9. Switch to sqlite mode.
10. Make sure SPIKE Nexus and SPIKE Keepers are up and running.
11. Repeat steps 2--7.
12. Test recover and restore scenarios:
    (a: Nexus auto-recover; b: doomsday recovery)

If everything looks good so far and the unit tests pass, you can cut a release.

Ideally, this setup should be automated, but since our releasee cadence is 
not that frequent, it's okay to do these checks manually.

### Start the Test Environment

```bash
make start

# enter user password if prompted

# wait for the following prompt:
# <<
# >
# > Everything is set up.
# > You can now experiment with SPIKE.
# >
# <<
# > >> To begin, run './spike' on a separate terminal window.
# <<
# >
# > When you are done with your experiments, you can press 'Ctrl+C'
# > on this terminal to exit and cleanup all background processes.
# >
# <<
```

### Create a Secret

```bash
spike secret put /acme/db user=spike pass=SPIKERocks
# Output:
# OK
```

Verify secret:

```bash 
spike secret get /acme/db
# Output:
# pass: SPIKERocks
# user: spike
```

### Create a Policy

```bash
spike policy create --name=workload-can-read \
  --path="/tenants/demo/db/*" \
  --spiffeid="^spiffe://spike.ist/workload/*" \
  --permissions="read"
```

### Verify the Policy Creation

```bash
spike policy list --format=json

# Sample output:
# [
#  {
#    "id": "872478b1-cef6-45c1-8417-1de82995aaa4",
#    "name": "workload-can-read",
#    "spiffeIdPattern": "^spiffe://spike.ist/workload/*",
#    "pathPattern": "/tenants/demo/db/*",
#    "permissions": [
#      "read"
#    ],
#    "createdAt": "2025-02-15T07:43:47.306286591-08:00",
#    "createdBy": ""
#  }
# ]
```

### Test the Demo Workload

```bash
./examples/consume-secrets/demo-create-policy.sh
./examples/consume-secrets/demo-register-entry.sh

./demo

# Output
# Secret found:
# password: SPIKE_Rocks
# username: SPIKE
```

If everything went well so far, we can assume that the current **SPIKE** release
is in a good enough shape.

<p>&nbsp;</p>

----

{{ toc_operations() }}

----

{{ toc_top() }}
