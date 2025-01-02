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
keep it lightweight. --- Passing the smoke test means that the core components
and the features of the system are reliably functional.

## Start the Test Environment

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
# > >> To begin, run './spike init' on a separate terminal window.
# <<
# >
# > When you are done with your experiments, you can press 'Ctrl+C'
# > on this terminal to exit and cleanup all background processes.
# >
# <<
```

## Create a Secret

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

## Create a Policy

```bash
spike policy create --name=workload-can-read \
  --path="/tenants/demo/db/*" \
  --spiffeid="^spiffe://spike.ist/workload/*" \
  --permissions="read"
```