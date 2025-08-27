# SPIKE Project Context for Claude Code

## Key Conventions

### SPIFFE ID and Path Patterns

SPIKE Policies use `SPIFFEIDPattern` and `PathPattern` fields. Those fields
are regular expression Strings; NOT globs.

- **For Policy SPIFFEID and Path patterns, ALWAYS use regex patterns, NOT globs**
- ✅ Correct: `/path/to/.*`, `spiffe://example\.com/workload/.*`
- ❌ Wrong: `/path/to/*`, `spiffe://example.com/workload/*`

### Paths used in Secrets and Policies are NOT Unix-like paths; they are Namespaces

The path is just a key to define a namespace (as in `secrets/db/creds`)
Thus, they should **NEVER** start with a forward slash:

- ✅ Correct: `secrets/db/creds`
- ❌ Wrong: `/secrets/db/creds`

While the system allows trailing slashes in paths, that is
1. highly-discouraged.
2. the behavior may change and the system may give an error or warning in
   the future.

### Do not invent environment variables

The following table lists the environment variables that you can use to
configure the SPIKE components. **DO NOT** make you your own environment 
variables. Use them from the table below. -- If the environment variable
does not exist in this table, scan the codebase to see if there are any
missing environment variables that are not mentioned, and suggest updates here.

| Component    | Environment Variable                    | Description                                                                                                                                         | Default Value                                                            |
|--------------|-----------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------|
| SPIKE Keeper | `SPIKE_KEEPER_TLS_PORT`                 | The TLS port the current SPIKE Keeper instance listens on.                                                                                          | `":8443"`                                                                |
| SPIKE Nexus  | `SPIKE_NEXUS_KEEPER_PEERS`              | A mapping that contains a comma-delimited list of URLs for all SPIKE Keepers that SPIKE Nexus knows about.                                          | "" (check `./hack/bare-metal/startup/start-nexus.sh` for usage examples. |
| SPIKE Nexus  | `SPIKE_NEXUS_TLS_PORT`                  | The TLS port SPIKE Nexus listens on.                                                                                                                | `":8553"`                                                                |
| SPIKE Nexus  | `SPIKE_NEXUS_MAX_SECRET_VERSIONS`       | The maximum number of versions of a secret that SPIKE Nexus stores.                                                                                 | `10`                                                                     |
| SPIKE Nexus  | `SPIKE_NEXUS_BACKEND_STORE`             | The backend store SPIKE Nexus uses to store secrets (memory, s3, sqlite).                                                                           | `"sqlite"`                                                               |
| SPIKE Nexus  | `SPIKE_NEXUS_DB_OPERATION_TIMEOUT`      | The timeout for database operations.                                                                                                                | `"15s"`                                                                  |
| SPIKE Nexus  | `SPIKE_NEXUS_DB_JOURNAL_MODE`           | The journal mode for the SQLite database.                                                                                                           | `"WAL"`                                                                  |
| SPIKE Nexus  | `SPIKE_NEXUS_DB_BUSY_TIMEOUT_MS`        | The timeout for the database to wait for a lock.                                                                                                    | `1000`                                                                   |
| SPIKE Nexus  | `SPIKE_NEXUS_DB_MAX_OPEN_CONNS`         | The maximum number of open connections to the database.                                                                                             | `10`                                                                     |
| SPIKE Nexus  | `SPIKE_NEXUS_DB_MAX_IDLE_CONNS`         | The maximum number of idle connections to the database.                                                                                             | `5`                                                                      |
| SPIKE Nexus  | `SPIKE_NEXUS_DB_CONN_MAX_LIFETIME`      | The maximum lifetime of a database connection.                                                                                                      | `"1h"`                                                                   |
| SPIKE Nexus  | `SPIKE_NEXUS_DB_INITIALIZATION_TIMEOUT` | The maximum initialization time for SPIKE Nexus DB before bailing out                                                                               | `30s`                                                                    |
| SPIKE Nexus  | `SPIKE_NEXUS_DB_SKIP_SCHEMA_CREATION`   | If set to `true`, skip creating SPIKE Nexus backing store. When set to `true`, the operator will manually have to create the initial backing store. | `false`                                                                  |
| SPIKE Nexus  | `SPIKE_NEXUS_PBKDF2_ITERATION_COUNT`    | The number of iterations for the PBKDF2 key derivation function.                                                                                    | `600000`                                                                 |
| SPIKE Nexus  | `SPIKE_NEXUS_RECOVERY_TIMEOUT`          | The timeout for attempting recovery from SPIKE Keepers. 0 = unlimited                                                                               | `0`                                                                      |
| SPIKE Nexus  | `SPIKE_NEXUS_RECOVER_MAX_INTERVAL`      | Maximum interval between retries the recovery operation's backing off algorithm                                                                     | `60s`                                                                    |
| SPIKE Nexus  | `SPIKE_NEXUS_RECOVERY_POLL_INTERVAL`    | The duration between attempts to poll the list of SPIKE Keepers during initial bootstrapping.                                                       | `5s`                                                                     |
| SPIKE Nexus  | `SPIKE_NEXUS_SHAMIR_SHARES`             | The total number of shares used for secret sharding, this should be equal to the number of SPIKE Keepers too.                                       | `3`                                                                      |
| SPIKE Nexus  | `SPIKE_NEXUS_SHAMIR_THRESHOLD`          | The minimum number of shares to be able to reconstruct the root key.                                                                                | `2`                                                                      |
| SPIKE Nexus  | `SPIKE_NEXUS_KEEPER_UPDATE_INTERVAL`    | The duration between SPIKE Nexus updates SPIKE Keepers with the relevant shard information.                                                         | `5m`                                                                     |
| SPIKE Pilot  | `SPIKE_PILOT_SHOW_MEMORY_WARNING`       | Whether to show a warning when the system cannot lock memory for security.                                                                          | `false`                                                                  |
| All          | `SPIKE_SYSTEM_LOG_LEVEL`                | The log level for all SPIKE components (`"DEBUG"`, `"INFO"`, `"WARN"`, `"ERROR"`).                                                                  | `"WARN"`                                                                 |
| All          | `SPIKE_NEXUS_API_URL`                   | The URL where SPIKE Nexus can be reached                                                                                                            | `"https://localhost:8553"`                                               |
| All          | `SPIKE_TRUST_ROOT`                      | The SPIFFE trust root used within the SPIKE trust boundary. Can be a single entry, or a comma-delimited list of suitable trust roots.               | `"spike.ist"`                                                            |
| All          | `SPIKE_TRUST_ROOT_KEEPER`               | The SPIFFE trust root used for SPIKE Keeper instances. Can be a single entry, or a comma-delimited list of suitable trust roots.                    | `"spike.ist"`                                                            |
| All          | `SPIKE_TRUST_ROOT_PILOT`                | The SPIFFE trust root used for SPIKE Pilot instances. Can be a single entry, or a comma-delimited list of suitable trust roots.                     | `"spike.ist"`                                                            |
| All          | `SPIKE_TRUST_ROOT_NEXUS`                | The SPIFFE trust root used for SPIKE Nexus instances. Can be a single entry, or a comma-delimited list of suitable trust roots.                     | `"spike.ist"`                                                            |
| All          | `SPIKE_BANNER_ENABLED`                  | Whether to display the SPIKE banner on startup. Set to `true` to enable.                                                                            | `true`                                                                   |
| All          | `SPIFFE_ENDPOINT_SOCKET`                | The Unix domain socket path used for SPIFFE Workload API                                                                                            | `"unix:///tmp/spire-agent/public/api.sock"`                              |


### Error Handling Strategy
- `panic()` for "should never happen" errors (testable)
- `os.Exit(1)` should NEVER happen (panic instead; it is testable)
- `os.Exit(0)` for successful early termination (--help, --version)
- Libraries should return errors, not call os.Exit()

### Architecture
- SPIKE Nexus: Secret management service
- SPIKE Pilot: CLI tool for users
- SPIKE Bootstrap: Initial setup tool
- SPIKE Keeper: Secret injection agent

### Database
- SQLite backend uses `~/.spike/data/spike.db` (hardcoded, not configurable)
- Encryption keys are `crypto.AES256KeySize` byte (32 bytes)
- Schema in `app/nexus/internal/state/backend/sqlite/ddl/statements.go`

### Common Mistakes to Avoid
1. Don't invent environment variables---check existing code first
2. Use regex patterns, not globs, for SPIFFE ID / path pattern matching  
3. Don't assume libraries exist---check imports/dependencies
4. Follow existing naming conventions and file organization
5. Test files should mirror the structure they're testing

## Project Structure
```
app/
├── nexus/          # Secret management service
├── pilot/          # CLI tool
├── bootstrap/      # Setup tool
└── keeper/         # Agent
internal/config/    # Configuration helpers
```

## Coding Conventions

### Use Proper English

During generating documentation, you often forget articles and prepositions,
and sometimes make basic grammatical errors.

For example `// Test with empty map` should better have been
`// Test with an empty map`.

Pay attention to punctuation and grammar.

### Line Length

The code has 80-character line length (including tests, and markdown files).
Tabs are counted as two characters.

When it's not possible, it's okay to make exceptions, but try your best to keep
the code within 80 chars.

## When in Doubt
- Look at existing similar files for patterns
- Check imports to see what's actually available
- Use Grep/Glob tools to find existing implementations