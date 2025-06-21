+++
# //    \\ SPIKE: Secure your secrets with SPIFFE.
# //  \\\\\ Copyright 2024-present SPIKE contributors.
# // \\\\\\\ SPDX-License-Identifier: Apache-2.

title = "SPIKE on Linux"
weight = 4
sort_by = "weight"
+++

# SPIKE on Linux

In this guide, you will learn how to build, deploy, and test [**SPIKE**][spike]
from the source. This guide assumes basic familiarity with terminal commands and
the ability to install and execute the required software. It is recommended to
have administrative privileges on your system, as some steps might require them.

The tools and resources mentioned in this guide are essential for building and
working with **SPIKE** effectively. Make sure to follow each step carefully to
ensure a smooth experience. In case you encounter issues, please discuss
them on the [SPIFFE community Slack][slack].

[slack]: https://slack.spiffe.io/ "SPIFFE Slack"
[spike]: @/getting-started/about.md "About SPIKE"

## Prerequisites

This quickstart guide assumes you are using an [Ubuntu Linux][ubuntu] operating
system. The steps may slightly differ if you are using a different operating
system.

**SPIKE** can run anywhere [SPIFFE][spiffe] can be deployed. For consistency,
the tutorials and guides in **SPIKE** documentation use [**Ubuntu**][ubuntu] as
the base operating system. Though, if you encounter issues with your OS, feel
free to discuss them on the [SPIFFE community Slack][slack].

[ubuntu]: https://ubuntu.com/
[spiffe]: https://spiffe.io/

Here are the OS details that we are testing this guide on:

```txt
DISTRIB_ID=Ubuntu
DISTRIB_RELEASE=24.04
DISTRIB_CODENAME=noble
DISTRIB_DESCRIPTION="Ubuntu 24.04 LTS"
```

In addition, you will need the usual suspects:

* [`git`](https://git-scm.com/)
* [`make`](https://www.gnu.org/software/make/)
* [`go`](https://go.dev/) (*the most recent version would do*)
* [`node`](https://nodejs.org) (*at least [GitHub Copilot][copilot] requires it on Linux*)
* [`build-essential`](https://packages.ubuntu.com/hirsute/build-essential)
  (*i.e., `sudo apt install build-essential`*)

[copilot]: https://copilot.github.com/ "GitHub Copilot"

## Go Environment Setup

Here's part of the `go env` setting we use for this guide. Yours might slightly
vary depending on your development configuration.

The environment setup shown below is mostly what Go uses by default, yet, we
provide them just-in-case to eliminate any environment-related setup issues you
might face.

```bash
go env
# GO111MODULE='on'
# GOCACHE='/home/spike/.cache/go-build'
# GOENV='/home/spike/.config/go/env'
# GOMODCACHE='/home/spike/packages/go/pkg/mod'
# GONOPROXY=''
# GONOSUMDB=''
# GOOS='linux'
# GOPATH='/home/spike/packages/go'
# GOPRIVATE=''
# GOPROXY='https://proxy.golang.org,direct'
# GOROOT='/usr/local/go'
# GOSUMDB='sum.golang.org'
# GOTOOLCHAIN='auto'
# GOMOD='/home/spike/Desktop/WORKSPACE/spike/go.mod'
# GOWORK=''
```

If you need, you can also use Go's built-in tooling to view and modify your Go
environment settings. Use the `go env` command to inspect or set specific
environment variables.

For example:

```bash
# View the current list of environment variables
go env

# Set a specific environment variable like GOPATH
go env -w GOPATH=$HOME/my-gopath

# Set multiple variables, e.g., GOROOT and GO111MODULE
go env -w GOROOT=/usr/local/go GO111MODULE=on

# Verify the changes were made
go env GOPATH
go env GOROOT
go env GO111MODULE
```

These changes made using the `go env -w` command are persistent and stored in
Go configuration files. You can view these changes in the file located at
`$(go env GOENV)`. To reset a variable to its default value, use:

```bash
go env -u GOPATH
```

## Building SPIRE

To get started, let's create a development version of [**SPIRE**][spire].
Note that this is not a production-ready setup. For production, you should
follow the [official SPIRE documentation][spire-prod].

[spire]: https://spiffe.io/docs/latest/spire-about/ "SPIRE"
[spire-prod]: https://spiffe.io/docs/latest/deploying/configuring/ "SPIRE Production Configuration"

Let's first build **SPIRE** from the source:

```bash
echo 'export WORKSPACE="$HOME/-change_to_dev_dir-"' >> ~/.profile
source ~/.profile
echo $WORKSPACE
cd $WORKSPACE
git clone https://github.com/spiffe/spire && cd spire
make build
```

## Adding SPIRE Binaries to `$PATH`

Add the **SPIRE** binaries to your `$PATH`:

```bash
# ~/.profile
export PATH=$PATH:$WORKSPACE/spire/bin
echo 'PATH=$PATH:$WORKSPACE/spire/bin' >> ~/.profile
```

## Adding SPIKE Binaries to `$PATH`

Additionally, you can source the following file to define additional
**SPIKE**-related environment variables for your convenience. This is not
required because if you don't define them, **SPIKE** will assume sensible
defaults.

Sourcing `./hack/lib/env.sh` allows you to override the default **SPIKE**
environment settings. This can be particularly useful for development
purposes to test custom setups or alternative paths.

Having all overrides in a single place is also handy as it doubles
as documentation to help understand the development environment.

```bash
# ~/.profile

# ...

# SPIKE Environment configuration                                                
source $WORKSPACE/spike/hack/lib/env.sh 
```

## Verifying SPIRE Installation

Verify SPIRE installation as follows:

```bash 
source ~/.profile
spire-server -h
```

Output:

```txt
Usage: spire-server [--version] [--help] <command> [<args>]

Available commands are:
    agent                
    bundle               
    entry                
    federation           
    healthcheck          Health status 
    jwt                  
    localauthority       
    logger               
    run                  Runs the server
    token                
    upstreamauthority    
    validate             Validates config 
    x509  
```

## Building SPIKE

Next, build **SPIKE** binaries:

```bash
cd $WORKSPACE/spike
make build

# Created files:
#   keeper*
#   nexus*
#   spike*
```

## Configure Local DNS

[The default agent configuration file][agent-config] uses
`spire.spike.ist` as the SPIRE Server DNS name. To resolve this name to the
loopback address, add the following entry to your `/etc/hosts` file:

[agent-config]: https://github.com/spiffe/spike/blob/main/config/spire/agent/agent.conf#L4

```bash
# /etc/hosts

# If SPIRE Server is running on a different IP, replace
# this with the correct IP address.
127.0.0.1 spire.spike.ist
```

## Starting SPIKE

There is a starter script that combines and automates some of the steps in the
following sections. It configures and runs SPIRE Server, SPIRE Agent,
SPIKE Nexus, and SPIKE Keeper.

You can run this to start all the required components:

```bash
# Start everything.
make start
```

And then, on a separate terminal, you can run `spike`:

```bash
# Make sure you have the `spike` binary in your PATH.
spike

# Sample Output: 
# SPIKE v$version
# >> Secure your secrets with SPIFFE: https://spike.ist/ #
#
# Usage:
#  spike [command]
#
# Available Commands:
#   completion  Generate the autocompletion script
#   help        Help about any command
#   operator    Manage admin operations
#   policy      Manage policies
#   secret      Manage secrets
#
# Flags:
#  -h, --help   help for spike
# 
# Use "spike [command] --help" for help.
```

Although the `make start` script is convenient, it might be useful
to run the components individually to understand the process better and
debug any issues that might arise.

The following sections will guide you through the individual steps.

> **CLI Reference**
>
> Since the **SPIKE CLI** is a work in progress and highly in flux, the best
> way to get the most up-to-date information is to run `spike --help` or
> `spike [command] --help` to learn about the available commands and flags.
>
> In addition, you can [check out the demo recordings][demo] to see the CLI in
> action.

[demo]: @/community/presentations.md

## Start SPIRE Server

Start the SPIRE Server:

```bash
cd $WORKSPACE/spike
./hack/bare-metal/startup/spire-server-start.sh
```

## Creating Registration Entries

The following script will create registration entries for the SPIKE components:

```bash
cd $WORKSPACE/spike
./hack/bare-metal/entry/spire-server-entry-spike-register.sh
```

## Start SPIRE Agent

Start the SPIRE Agent:

```bash
cd $WORKSPACE/spike
./hack/bare-metal/startup/spire-agent-start.sh
```

## Start SPIKE Components

Then start **SPIKE** components:

Make sure you started the following binaries, each runs on a specific terminal
window.

Start the workloads:

```bash
# Optional: Increase the log level to debug:
export SPIKE_SYSTEM_LOG_LEVEL=debug

cd $WORKSPACE/spike

# Start SPIKE Nexus in one terminal.
./hack/bare-metal/startup/start-nexus.sh

# Start SPIKE Keepers in separate terminals.
./hack/bare-metal/startup/start-keeper-1.sh
./hack/bare-metal/startup/start-keeper-2.sh
./hack/bare-metal/startup/start-keeper-3.sh
```

Here is how one of these **SPIKE Keeper** startup scripts looks like:

```bash
# ./hack/bare-metal/startup/start-keeper-1.sh
SPIKE_KEEPER_TLS_PORT=':8443' \
./keeper
````

And here is how **SPIKE Nexus** startup script looks like:

```bash
# ./hack/bare-metal/startup/start-nexus.sh
SPIKE_NEXUS_KEEPER_PEERS='https://localhost:8443,\
https://localhost:8543,https://localhost:8643'
./nexus
```

> **Sequential SPIKE Keeper IDs**
>
> The mapping in `SPIKE_NEXUS_KEEPER_PEERS` should start from `"1"`
> and increase monotonically without any gaps in the sequence as shown
> in the sample code above. This is because of the way SPIKE Nexus internally
> computes and distributes the Shamir Shards. Not following this sequence
> will lead to errors---We may improve this behavior in the future and make
> it more flexible.

## Using SPIKE Pilot

Define an alias to **SPIKE** Pilot:

```bash
# ~/.bashrc

# path to the SPIKE Pilot binary (`spike`)
alias spike=$WORKSPACE/spike/spike
```

Run **SPIKE** Pilot and explore the CLI:

```bash
spike
```

## Testing Out SPIKE

Let's test **SPIKE** by creating a secret:

```bash
spike secret put /tenants/acme/credentials/db \
  username=root pass=SPIKERocks

# Output:
# OK
```

Now, let's read the secret back:

```bash
spike secret get /tenants/acme/credentials/db

# Output:
# pass: SPIKERocks
# username: root
```

Let's delete the secret now:

```bash
spike secret delete /tenants/acme/credentials/db

# Output:
# OK
```

If you try to read the secret again, you won't be able to get it.

Feel free to experiment with other **SPIKE** commands in your sandbox
environment to explore its capabilities and better understand how it works. This
is a great way to familiarize yourself with its features and test various
scenarios safely.

## Uninstalling SPIKE

Retaining the **SPIKE** binaries on your system poses no issues. These binaries
are compact, consuming minimal disk space and no resources when inactive. As
simple executable files, they have no impact on your system's performance when
not in use. Therefore, keeping them installed is completely harmless.

However, if you want to wipe everything out, you can just remove the binaries
and **SPIKE**'s data folder, and that would be it:

```bash
rm -rf ~/.spike
rm spike
rm keeper
rm nexus
```

If you have `spire-server` and `spire-agent` on your system, and you are not
using them for anything else; you can remove them too:

```bash
rm spire-server
rm spire-agent
```

## Have Fun

That's about it.

Enjoy.

----

{{ toc_getting_started() }}

----

{{ toc_top() }}
