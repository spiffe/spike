![SPIKE](../../assets/spike-banner-lg.png)

## SPIKE Demo Workload

This is a sample application that demonstrates how to use the SPIKE SDK to
interact with SPIKE Nexus for secret management.

## What It Does

The demo performs the following operations:

1. Connects to SPIKE Nexus using the SPIFFE Workload API
2. Creates a secret at path `tenants/demo/db/creds` containing database
   credentials (username and password)
3. Reads the secret back from the same path
4. Displays the retrieved secret data

## Prerequisites

The demo requires a complete SPIKE environment with:

1. **SPIRE infrastructure** (*server and agent*) running
2. **SPIKE Keeper instances** running (*for root key sharding*)
3. **SPIKE Nexus** running (*the secret management service*)
4. **SPIKE Bootstrap** completed (*system initialization*)
5. **SPIRE registration** for the demo workload
6. **SPIKE policies** granting the demo app read/write permissions

## Environment Variables

The demo requires the following environment variable to connect to the SPIRE
agent:

* `SPIFFE_ENDPOINT_SOCKET`---The Unix domain socket path for the SPIFFE
  Workload API (default: `unix:///tmp/spire-agent/public/api.sock`)

## Running the Demo

### Using the Complete Bare-Metal Setup

The easiest way to run the demo is using the provided startup script that
configures the entire SPIKE environment:

```bash
./hack/bare-metal/startup/start.sh
```

This script will:
* Build all SPIKE binaries (including the demo)
* Start SPIRE server and agent
* Start SPIKE Keeper instances
* Start SPIKE Nexus
* Run SPIKE Bootstrap
* Register the demo workload with SPIRE
* Create policies for the demo workload
* Run the demo to verify the setup

### Manual Execution

If the environment is already configured, run the built binary:

```bash
demo
```

Or build and run from source:

```bash
go run ./app/demo/cmd/main.go
```

Note: Manual execution requires all prerequisites to be properly configured.

## API Usage

The demo showcases these SPIKE SDK operations:

* **`spike.New()`** - Creates a connection to SPIKE Nexus using the
  Workload API
* **`api.PutSecret(path, data)`** - Stores a secret at the specified path
* **`api.GetSecret(path)`** - Retrieves a secret from the specified path
* **`api.Close()`** - Closes the connection when done

For more details, see the
[SPIKE SDK documentation](https://pkg.go.dev/github.com/spiffe/spike-sdk-go/api).
