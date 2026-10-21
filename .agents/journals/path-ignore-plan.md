<!-- ABOUTME: Records the scope and unresolved decisions for path-based ignore planning. -->
<!-- ABOUTME: Provides implementation context without changing executable behavior. -->

# Path ignore planning

- [decision] Fabian requested a plan, not implementation. The plan is in
  `.agents/plans/path-ignore.md`.
- [decision] The proposed interface accepts one literal path, uses `--here` for
  the shell working directory's `.gitignore`, and uses `--allow` for negation.
- [risk] An exception below an ignored parent can remain ineffective. Report the
  blocking rule without automatically rewriting parent rules.
- [risk] A global ignore pattern cannot represent an arbitrary absolute filesystem
  path. Global pattern selection requires an explicit design decision and consent.
- [question] Confirm missing-target behavior, symbolic-link policy, Git availability,
  and `--here` outside repositories before implementing the affected stages.
- [technique] Use real temporary repositories and isolated Git configuration to
  verify ignore semantics, worktrees, repository boundaries, and global effects.
- [lesson] Some existing prompt tests in `main_test.go` only log expected outcomes.
  Tests for path-mode prompts must execute behavior and assert results.

## Reflection

The actionable decisions and risks are captured in the plan. No separate memory
file is needed for this planning-only session.

## Operator decisions

- [decision] Allow missing targets. A trailing slash identifies a directory;
  otherwise use a file-style rule.
- [decision] Ignore a target symbolic link itself, not its referent.
- [decision] Refuse to write through a symbolic-link ignore file.
- [decision] Require Git for path mode; template commands remain independent of it.
- [decision] Allow `--here` outside repositories, with a notice about when the file
  takes effect.
- [decision] Outside repositories, suggest the escaped basename as a global
  pattern, show the destination, warn about cross-repository effects, and require
  explicit confirmation with a default of No.
- [decision] Reject paths that traverse linked directories; ask the user to supply
  the destination's actual path instead.
- [decision] The questions recorded above are resolved in the plan. No executable
  implementation is authorized by these planning decisions.
