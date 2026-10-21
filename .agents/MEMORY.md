<!-- ABOUTME: Preserves durable implementation constraints for igittigit. -->
<!-- ABOUTME: Summarizes Git semantics and the isolation required for path-mode tests. -->

# Memory

## Path rules

- [decision] Path mode accepts one literal path. Target symlinks mean the link,
  linked-directory traversal and symlinked ignore files are refused, missing
  targets are allowed, and a trailing slash marks a future directory.
- [technique] Use real temporary repositories and isolated HOME, XDG, and Git
  configuration for path tests. Never write the operator's actual global file.
- [lesson] `git check-ignore --no-index -z -v --stdin` reports the effective
  source/line/pattern, including negations. Inspect the pattern's `!` prefix;
  exit status alone does not distinguish an effective exception. Check tracking
  separately without mutating the index.
- [decision] Blocked exceptions are written but return exit 1 and identify the
  blocker. Never rewrite parent exclusions or claim descendants are all allowed.
- [lesson] Not-a-repository discovery has separate normal and filesystem-boundary
  diagnostics. Other Git failures must not trigger a global fallback.
- [decision] Global basename offers require informed default-No consent and an
  absolute configured/default destination. Relative excludes-file/XDG locations
  cannot promise cross-repository effects; Git configuration stays unchanged.
- [lesson] Validate the resolved path for line breaks, including inherited current
  directory components. Escape bytes without lossy Unicode conversion.
