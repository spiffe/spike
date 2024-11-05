![SPIKE](assets/spike-banner.png)

## Quickstart

This quickstart guide assumes you are using an Ubuntu Linux operating system.
The steps may slightly differ if you are using a different operating system.

Make sure you have [SPIRE](https://spiffe.io/spire) installed on your system.

The `./hack/build-spire.sh` can be used as a starting point to build SPIRE
from source and install it to your system.

Set up environment variables

```bash
# This is the SPIKE repo folder that you cloned.
export SPIKE_ROOT=/path/to/the/project/folder/of/spike/
```

Start SPIRE Server (and create necessary tokens and registration entries):

```bash
./hack/spire-server-starter.sh
```

Start SPIRE Agent (and consume the tokens):

```bash
./hack/spire-agent-starter.sh
```

Start the workloads:

```bash
./nexus
./keeper
./spike
```

Define an alias to **SPIKE** Pilot:

```bash
# ~/.bashrc

# path to the SPIKE Pilot binary (`spike`)
alias spike=$USER/WORKSPACE/spike/spike
```

Run **SPIKE** Pilot and explore the CLI:

```bash
spike

# Sample Output:
# SPIKE v0.1.0
# >> Secure your secrets with SPIFFE: https://spike.ist/ #
# Usage: spike <command> [args...]
# Commands:
#   init
#   put <path> <key=value>...
#   get <path> [-version=<n>]
#   delete <path> [-versions=<n1,n2,...>]
#   undelete <path> [-versions=<n1,n2,...>]
#   list
```

That's about it.

Enjoy.

## Setting Up Postgres

Here are steps to set up Postgres for Ubuntu Linux:

Install Postgres:

```bash
sudo apt install postgres
```

Configure Postgres to listen everywhere:

```bash 
sudo vim /etc/postgresql/$version/main/postgresql.conf
# change listen_address as follows:
# listen_address = '*'
```

Create database `spike`:

```bash
sudo -u postgres psql -c 'create database spike;';
```

Set a password for the postgres user:

```bash 
ALTER USER postgres with encrypted password 'your-password-here';
```

Enable SSL:

```bash
sudo vim /etc/postgresql/16/main/pg_hba.conf

# Update the file and set your IP range accordingly.
# hostssl spike postgres 10.211.55.1/24 scram-sha-256
```

That's it. Your database is configured for local development.