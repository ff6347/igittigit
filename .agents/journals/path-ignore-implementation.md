<!-- ABOUTME: Records implementation and verification of literal path ignore commands. -->
<!-- ABOUTME: Captures Git semantics, safety decisions, and remaining follow-up issues. -->

# Path ignore implementation

- [decision] Implemented the approved path command contract on `feat/path-ignore`:
  nearest applicable ancestor ignore file, `--here`, `--allow`, literal paths,
  nonexistent targets, target links, and explicit global basename consent.
- [technique] CLI integration tests compile the real executable and use temporary
  repositories, linked worktrees, and isolated HOME/XDG/Git configuration. No test
  modifies the operator's global ignore file or index.
- [lesson] Git's normal no-repository diagnostic differs at filesystem boundaries.
  Discovery recognizes both messages while propagating configuration and other
  invocation failures. A real `/dev` boundary regression verifies the distinction.
- [lesson] Validate resolved paths as well as supplied arguments: the current
  directory can contribute line breaks. Preserve literal filename bytes during
  escaping; rune iteration replaces invalid UTF-8 bytes.
- [decision] Refuse relative `core.excludesFile` and XDG destinations for global
  offers: their meaning changes with the repository directory. Explain the need
  for an absolute location without changing Git configuration.
- [technique] Ignore-file writes use exclusive creation for absent files and
  validate opened-descriptor identity against the expected regular file before
  writing. Replacements by symbolic links cannot redirect writes to a referent.
- [decision] An exception blocked by a parent exclusion is appended unchanged,
  reports the actual Git source/line/pattern, and exits 1. No parent rule or index
  entry is rewritten. Tracked paths are warned about even on repeated no-op calls.

## Verification

- Baseline tests passed before implementation. CLI dispatch/destination tests
  failed against the template-only executable, then passed with path mode.
- Global consent tests failed before the fallback was implemented. Review
  regressions demonstrated failures for inherited line breaks, empty arguments,
  filename byte escaping, relative global destinations, invalid directory
  traversal, filesystem-root basename suggestions, and filesystem-boundary
  discovery before their fixes.
- Final checks passed: gofmt, `go test -count=1 ./...`,
  `go test -race -count=1 ./...`, `go vet ./...`, and a temporary native build.
- Built CLI smoke tests passed for nested ignore files and CRLF preservation,
  worktrees, missing paths, target links, linked-directory refusal, effective and
  blocked exceptions, isolated global confirmation/cancellation and two-repository
  matching, outside-repository `--here`, and cached template listing without Git.
- Generated fish completions passed `fish --no-config -n` and actual completion
  checks for filenames, both flags, `--`, and cached template names.
- Linux/amd64 and Windows/amd64 cross-builds passed. Runtime verification was on
  macOS, not Linux or Windows.
- [lesson] Existing prompt tests only log expectations; tracked separately as
  git-bug `272bc75`. README's Go prerequisite differs from go.mod; tracked as
  git-bug `972534b`. Neither unrelated issue was mixed into this feature.
- [decision] The completed implementation plan is removed; the README and CLI
  help are the human-facing command reference.
