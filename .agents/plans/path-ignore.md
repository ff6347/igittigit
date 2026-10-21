<!-- ABOUTME: Specifies path-based ignore and exception commands for igittigit. -->
<!-- ABOUTME: Defines implementation stages, safety boundaries, and verification criteria. -->

# Path-based ignore rules

## Goal

Allow a filesystem path as the command argument to append an ignore rule to the
nearest applicable `.gitignore`. Support choosing `./.gitignore` explicitly and
adding exceptions with `!` rules. Preserve the existing template commands.

This document is an implementation plan, not a statement of shipped behavior.

## Command contract

```bash
igittigit ./build/
igittigit --here ./build/
igittigit --allow ./config/example.env
igittigit --here --allow ./config/example.env
```

- Accept one literal filesystem path, not a gitignore pattern or a batch of paths.
- Resolve relative paths against the shell's current working directory.
- `--here` selects `.gitignore` in that working directory, creating it if absent.
  It does not mean the directory named by the argument. Allow this outside a
  working tree, with a notice that it takes effect once the directory belongs to
  a Git working tree; do not offer the global fallback when `--here` is supplied.
- Without `--here`, select the closest ancestor `.gitignore` within the target's
  Git working tree. Start at the target's parent, including for directory targets:
  a directory's own `.gitignore` cannot ignore that directory itself.
- If there is no applicable `.gitignore`, offer to create one at the working-tree
  root. Declining or reaching EOF must leave the filesystem unchanged.
- `--allow` appends a negated rule. It does not force-add files, change the index,
  or rewrite other ignore rules.
- Flags precede the path. Support `--` for paths beginning with `-`.
- Existing subcommands retain their meaning. Use `./list`, for example, to refer
  to a local path whose name collides with a subcommand.
- Reject unknown flags, missing paths, and multiple path arguments before writing.

## Destination selection

### Repository-local behavior

1. Resolve the target and find its containing working tree using Git, not merely
   the presence of a `.git` directory. Support nested repositories and worktrees.
2. Walk upward from the target's parent to that working-tree root, inclusive.
3. Select the first applicable `.gitignore`; never cross the repository boundary.
4. With `--here`, use the working directory instead, but reject targets outside
   its descendant tree or across a nested-repository boundary.
5. Reject a rule targeting the selected directory itself or Git administrative
   paths. A local ignore file cannot exclude its containing directory, and Git
   administrative paths are outside this feature's scope.

Do not stop discovery at the user's home directory. Repositories can exist above
it or outside it. Git discovery must distinguish "not a working tree" from an
actual invocation or permission failure.

### No containing working tree

The requested fallback is an explicit offer to add a global ignore pattern.
Never interpret a literal absolute path as an equivalent global rule.

- Explain that global patterns apply across repositories, not arbitrary locations
  on the filesystem.
- Show the destination and exact proposed pattern before confirmation.
- Resolve `core.excludesFile` through Git configuration, respecting path expansion
  and the default `$XDG_CONFIG_HOME/git/ignore` or `$HOME/.config/git/ignore`.
- Do not alter Git configuration or use `~/.gitignore` by assumption.
- Declining, EOF, or invalid confirmation leaves all files and directories untouched.
- Treat global negation as subject to normal Git precedence; it cannot override a
  repository-local exclusion.

Suggest the target's literal basename as the global pattern, escaped using the
same literal-name rules as local entries. Preserve the directory suffix and
`--allow` negation where applicable. For `/some/project/build/`, suggest `build/`,
not an absolute path or a root-anchored rule. Warn that this broader pattern
matches across all repositories. Show both the exact pattern and destination,
and require explicit confirmation with a default of No.

## Rule generation and writes

- Generate a slash-separated path relative to the selected ignore file directory.
- Anchor local rules with `/` to avoid matching unrelated names at other depths.
- Add a trailing `/` for directory targets; prefix with `!` for `--allow`.
- Escape literal pattern metacharacters, backslashes, and significant spaces.
  Verify behavior for `#` and `!` names as well as `*`, `?`, and brackets.
- Reject names containing line breaks rather than emitting multiple rules.
- Preserve existing file contents, comments, permissions, and line-ending style.
  Insert a separator if the existing final line has no newline.
- Check the effective rule before treating an operation as a no-op. A matching
  line earlier in the file is not sufficient if a later rule overrides it.
- Report the destination and resulting rule, or explain why no write is needed.
- Check and surface open, write, and close errors. Do not report success after a
  failed write. Do not overwrite template content.

Example: with `app/.gitignore` selected, ignoring `app/output/` writes:

```gitignore
/output/
```

Allowing `app/config/example.env` writes:

```gitignore
!/config/example.env
```

## Git-specific outcomes

- Detect tracked files and explain that ignore rules do not untrack them. Never
  run `git rm`, `git add`, or otherwise mutate the index.
- Verify ignore and exception effects with Git, including parent exclusions and
  more-specific rules. Do not implement an independent gitignore matcher.
- A parent rule such as `/config/` blocks `!/config/example.env`. Append the
  requested exception, then explain that it is ineffective and identify the
  blocking rule. Do not automatically broaden access to the parent directory.
- Distinguish "rule written" from "requested ignore state achieved" in output;
  an ineffective exception must not be presented as success. Cover its exit
  status in the command-level tests.
- Directory exceptions do not guarantee every descendant is allowed; descendant
  rules still apply.

## Confirmed path policies

- Allow nonexistent targets. A trailing `/` identifies a future directory;
  otherwise generate a file-style rule without a trailing slash. Git can also
  match a directory with a file-style rule; it cannot express a file-only rule.
- For a missing target or missing parent directories, start Git discovery at the
  closest existing ancestor directory. Do not create the target or its parents.
- When the target itself is a symbolic link, operate on the named link, not its
  referent. Do not follow it for repository selection or directory classification.
- Reject target paths that traverse a symbolic-link directory, explaining that
  the user must supply the destination's actual path.
- Refuse to write through a symbolic-link ignore file, including for the global
  destination. Explain the refusal without changing the link or its referent.
- Require the Git executable for path mode, including `--here`. Give an actionable
  error if it is absent. Template commands remain independent of Git.

## Implementation sequence

Use Go's standard library and focused files. Do not restructure unrelated template,
cache, or HTTP code. For each executable behavior below, first add a failing test,
run it to establish the failure, implement the minimum behavior, then rerun it.

### 1. Establish baseline and command dispatch

- Run the existing tests before changes; raise any failures rather than hiding them.
- Inspect `main.go`, existing tests, and completion generation.
- Add executable CLI tests for path dispatch, both flags, `--`, invalid input,
  subcommand-name collisions, and unchanged template command dispatch.
- Keep path-mode errors and confirmation I/O testable without broad refactoring.

### 2. Resolve paths and destinations

- Add `ignore.go` and `ignore_test.go` for the path operation; split by
  responsibility only if needed to keep files readable.
- Test real temporary repositories: root and nested ignore files, no ignore file,
  invocation from a different directory, nested repositories, and linked worktrees.
- Cover `--here` inside and outside repositories, missing targets and parent
  directories, cancellation, missing Git, and repository boundaries.
- Cover target links, rejected linked-directory traversal, and refused writes
  through local or global ignore-file links. Assert no referent is modified.
- Ensure selection and validation happen before filesystem mutation.

### 3. Append literal ignore rules

- Add table-driven escaping tests and verify actual matching with Git.
- Cover files, directories, absolute inputs, spaces, metacharacters, duplicate
  invocations, and a later rule overriding an earlier identical rule.
- Cover empty files, missing final newline, line endings, content preservation,
  and write failures using real filesystem conditions.
- Assert sibling files with similar names are not accidentally ignored.

### 4. Add exceptions and effective-state checks

- Test an effective exception, an already allowed path, an excluded parent,
  descendant overrides, tracked paths, and rules from other ignore sources.
- Use `git check-ignore` with appropriate untracked-path handling to verify
  pattern semantics; use a separate tracked-path check.
- Assert output, exit status, unchanged index, and that blocking rules are not
  rewritten. Capture expected diagnostics rather than printing unvalidated errors.

### 5. Add the confirmed global fallback

- Test escaped basename suggestions, directory suffixes, `--allow`, explicit
  consent, and the default-No prompt. Verify `--here` bypasses the fallback.
- Test configured destinations, Git defaults, XDG paths, paths containing spaces,
  file creation after consent, rejection, EOF, and ineffective global exceptions.
- Show using two real repositories that the accepted pattern applies to both.
- Isolate HOME, XDG locations, and Git configuration in tests; never read or mutate
  the developer's actual global ignore file or Git configuration.

### 6. Document and verify

- Update `printUsage`, `README.md`, and generated fish completions for path mode,
  both flags, filenames, and the limitations of exceptions and global rules.
- Validate documentation and completion output directly; do not write tests that
  merely assert declarative source contents.
- Run `gofmt` on changed Go files, `go test ./...`, `go test -race ./...`,
  `go vet ./...`, and a build to a temporary output path.
- Perform CLI smoke tests with the built binary in temporary repositories.
- Review the diff for unrelated changes, commit focused increments, and push the
  feature branch. Remove this plan in a commit once implementation is complete.

## Acceptance criteria

- The four example commands select the documented destination and generate literal,
  correctly scoped rules whose effects are verified against real Git.
- No local operation crosses repository boundaries or silently edits the index.
- An ineffective exception identifies its blocker without rewriting other rules.
- Global changes require informed confirmation of both destination and pattern.
- Cancellation and pre-write validation failures leave files unchanged.
- Tests use real repositories and filesystem operations, not mocked Git behavior.
- Existing template commands, caching, and GitHub request behavior remain intact.
