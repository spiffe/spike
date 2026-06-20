# Conventions

<!--
UPDATE WHEN:
- New pattern is established and should be followed consistently
- Existing pattern is deprecated or superseded
- Team adopts new tooling that changes workflows
- Code review reveals recurring issues that need a convention

DO NOT UPDATE FOR:
- One-off exceptions (document in code comments)
- Experimental patterns not yet proven
- Personal preferences without team consensus
-->

## Naming

- **Constants use semantic prefixes**: Group related constants with prefixes
  - `Dir*` for directories (`DirContext`, `DirArchive`)
  - `File*` for file paths (`FileSettings`, `FileClaudeMd`)
  - `Filename*` for file names only (`FilenameTask`, `FilenameDecision`)
  - `*Type*` for enum-like values (`UpdateTypeTask`, `UpdateTypeDecision`)
- **Package name = folder name**: Go canonical pattern
  - `package initialize` in `initialize/` folder
  - Never `package initcmd` in `init/` folder
- **Maps reference constants**: Use constants as keys, not literals
  - `map[string]X{ConstKey: value}` not `map[string]X{"literal": value}`

## Patterns

- **Centralize magic strings**: All repeated literals belong in a `config` or `constants` package
  - If a string appears in 3+ files, it needs a constant
  - If a string is used for comparison, it needs a constant
- **Path construction**: Always use stdlib path joining
  - Go: `filepath.Join(dir, file)`
  - Python: `os.path.join(dir, file)`
  - Node: `path.join(dir, file)`
  - Never: `dir + "/" + file`
- **Constants reference constants**: Self-referential definitions
  - `FileType[UpdateTypeTask] = FilenameTask` not `FileType["task"] = "TASKS.md"`
- **Colocate related code**: Group by feature, not by type
  - `session/run.go`, `session/types.go`, `session/parse.go`
  - Not: `runners/session.go`, `types/session.go`, `parsers/session.go`

## Testing

- **Colocate tests**: Test files live next to source files
  - `foo.go` → `foo_test.go` in same package
  - Not a separate `tests/` folder
- **Test the unit, not the file**: One test file can test multiple related functions
- **Integration tests are separate**: `cli_test.go` for end-to-end binary tests

## Documentation

- **Godoc format**: Use canonical sections
  ```go
  // FunctionName does X.
  //
  // Longer description if needed.
  //
  // Parameters:
  //   - param1: Description
  //   - param2: Description
  //
  // Returns:
  //   - Type: Description of return value
  func FunctionName(param1, param2 string) error
  ```
- **Package doc in doc.go**: Each package gets a `doc.go` with package-level documentation
- **Copyright headers**: All source files get the project copyright header
