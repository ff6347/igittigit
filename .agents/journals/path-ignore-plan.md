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
