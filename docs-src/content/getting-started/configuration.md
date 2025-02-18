+++
# //    \\ SPIKE: Secure your secrets with SPIFFE.
# //  \\\\\ Copyright 2024-present SPIKE contributors.
# // \\\\\\\ SPDX-License-Identifier: Apache-2.

title = "Configuring SPIKE"
weight = 3
sort_by = "weight"
+++



# Configuring SPIKE

You can use environment variables to configure the **SPIKE** components.

The following table lists the environment variables that you can use to
configure the SPIKE components:

| Component    | Environment Variable                 | Description                                                                                               | Default Value                                   |
|--------------|--------------------------------------|-----------------------------------------------------------------------------------------------------------|-------------------------------------------------|
| SPIKE Keeper | `SPIKE_KEEPER_TLS_PORT`              | The TLS port the current SPIKE Keeper instance listens on.                                                | `":8443"`                                       |
| SPIKE Keeper | `SPIKE_KEEPER_API_URL`               | The API URL of the current SPIKE Keeper instance.                                                         | `"https://localhost:8443"                       |
| SPIKE Nexus  | `SPIKE_KEEPER_PEERS`                 | A mapping that contains `[{id:keeperurl}]` collection for all SPIKE Keepers that SPIKE Nexus knows about. | "" (check `./hack/start.sh` for usage examples. |
| SPIKE Nexus  | `SPIKE_NEXUS_POLL_INTERVAL`          | The interval SPIKE Nexus syncs up its state with SPIKE Keeper.                                            | `"5m"`                                          |
| SPIKE Nexus  | `SPIKE_NEXUS_TLS_PORT`               | The TLS port SPIKE Nexus listens on.                                                                      | `":8553"`                                       |
| SPIKE Nexus  | `SPIKE_NEXUS_MAX_SECRET_VERSIONS`    | The maximum number of versions of a secret that SPIKE Nexus stores.                                       | `10`                                            |
| SPIKE Nexus  | `SPIKE_NEXUS_BACKEND_STORE`          | The backend store SPIKE Nexus uses to store secrets (memory, s3, sqlite).                                 | `"sqlite"`                                      |
| SPIKE Nexus  | `SPIKE_NEXUS_DB_OPERATION_TIMEOUT`   | The timeout for database operations.                                                                      | `"5s"`                                          |
| SPIKE Nexus  | `SPIKE_NEXUS_DB_DATA_DIR`            | The directory where SPIKE Nexus stores the database.                                                      | `"./data"`                                      |
| SPIKE Nexus  | `SPIKE_NEXUS_DB_JOURNAL_MODE`        | The journal mode for the SQLite database.                                                                 | `"WAL"`                                         |
| SPIKE Nexus  | `SPIKE_NEXUS_DB_BUSY_TIMEOUT_MS`     | The timeout for the database to wait for a lock.                                                          | `1000`                                          |
| SPIKE Nexus  | `SPIKE_NEXUS_DB_MAX_OPEN_CONNS`      | The maximum number of open connections to the database.                                                   | `10`                                            |
| SPIKE Nexus  | `SPIKE_NEXUS_DB_MAX_IDLE_CONNS`      | The maximum number of idle connections to the database.                                                   | `5`                                             |
| SPIKE Nexus  | `SPIKE_NEXUS_DB_CONN_MAX_LIFETIME`   | The maximum lifetime of a database connection.                                                            | `"1h"`                                          |
| SPIKE Nexus  | `SPIKE_NEXUS_PBKDF2_ITERATION_COUNT` | The number of iterations for the PBKDF2 key derivation function.                                          | `600000`                                        |
| All          | `SPIKE_SYSTEM_LOG_LEVEL`             | The log level for all SPIKE components (DEBUG, INFO, WARN, ERROR).                                        | `"DEBUG"`                                       |
| All          | `SPIKE_TRUST_ROOT`                   | The trust root that SPIKE components use in their SVIDs and SPIFFE IDs.                                   | `"spike.ist"`                                   |

We'll add more configuration options in the future. Stay tuned.

----

{{ toc_getting_started() }}

----

{{ toc_top() }}
