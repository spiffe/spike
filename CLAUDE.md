# SPIKE Project Context for Claude Code

## Key Conventions

### SPIFFE ID and Path Patterns

SPIKE Policies use `SPIFFEIDPattern` and `PathPattern` fields. Those fields
are regular expression Strings; NOT globs.

- **For Policy SPIFFEID and Path patterns, ALWAYS use regex patterns, NOT globs**
- ✅ Correct: `/path/to/.*`, `spiffe://example\.org/workload/.*`
- ❌ Wrong: `/path/to/*`, `spiffe://example.org/workload/*`

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

The table in `docs-src/content/usage/configuration.md` contains a list of
environment variables that you can use to configure the SPIKE components.
**DO NOT** make you your own environment variables. Use them from the table
in that file---If the environment variable does not exist in the table, scan
the codebase to see if there are any missing environment variables that are not
mentioned and suggest updates in that table.

### Error Handling Strategy
- `panic()` for "should never happen", use `log.FatalLn()` instead---you can
  find usage examples in the codebase.
- `os.Exit(1)` should NEVER happen (use `log.FatalLn()` instead)
- `os.Exit(0)` for successful early termination (`--help`, `--version`)
- Libraries should return errors, **not** call `os.Exit()`.

### Architecture
- SPIKE Nexus: Secret management service
- SPIKE Pilot: CLI tool for users
- SPIKE Bootstrap: Initial setup tool
- SPIKE Keeper: Secret injection agent

### Database
- SQLite backend uses `~/.spike/data/spike.db`
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

During generating documentation, you often forget articles and prepositions
and sometimes make basic grammatical errors.

For example `// Test with empty map` should better have been
`// Test with an empty map`.

`Super permission acts as a joker — grants all permissions` should have been
`The "Super" permission acts as a joker—grants all permissions.` (no space
before and after em-dash).

While at it, you can tone down your em-dash usage. Yes, it is good grammar,
 but you tend to overuse it and liberally sprinkle it everywhere.

The same goes with emoji usage: This is a security-focused codebase, 
NOT a preteen's playground.

In short, pay extra attention to punctuation and grammar.

### Line Length

The code has 80-character line length (including tests and markdown files).
Tabs are counted as two characters.

When it's not possible, it's okay to make exceptions, but try your best to keep
the code within 80 chars.

## When in Doubt
- Look at existing similar files for patterns
- Check imports to see what's actually available
- Use Grep/Glob tools to find existing implementations
