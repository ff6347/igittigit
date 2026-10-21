# igittigit

A minimal CLI for adding literal path ignore rules and fetching .gitignore
templates from GitHub's API.

## Prerequisites

- [Go](https://go.dev/) 1.19+ (for installation)
- [Git](https://git-scm.com/) (required for path rules; not for template commands)
- [gh](https://cli.github.com/) (optional, for authenticated GitHub API requests)
- [fzf](https://github.com/junegunn/fzf) (optional, for `igittigit search`)

## Installation

### Homebrew

```bash
brew install ff6347/tap/igittigit
```

### Go

```bash
go install github.com/ff6347/igittigit@latest
```

### Build from source

```bash
git clone https://github.com/ff6347/igittigit.git
cd igittigit
go build -o igittigit
```

## Usage

```bash
igittigit list                 # List all available templates
igittigit search               # Interactive fuzzy search (requires fzf)
igittigit get Go               # Output template to stdout
igittigit get Go > .gitignore  # Save to file
igittigit get Go >> .gitignore # Append to existing file
igittigit completions          # Generate fish shell completions
```

### Path rules

```bash
igittigit ./build/                          # Nearest ancestor .gitignore
igittigit --here ./build/                   # Shell working directory's .gitignore
igittigit --allow ./config/example.env      # Append an exception
igittigit --here --allow ./config/example.env
igittigit -- -cache                         # Path beginning with a dash
igittigit ./list                            # Path named like a subcommand
```

Supply exactly one literal path, not a pattern. Quote spaces and shell
metacharacters. Flags must precede the path. Relative paths resolve from the shell
working directory. Missing paths are allowed without creating them; a trailing
`/` marks a future directory. Existing directories receive directory rules.

By default, the closest ancestor `.gitignore` within the target's working tree is
selected, starting at the target's parent even for directories. Nested repositories
and linked worktrees have their own boundaries. If no applicable file exists,
creation at the working-tree root requires confirmation (default **No**).

`--here` selects or creates `.gitignore` in the shell working directory. The target
must be below that directory and cannot cross a nested-repository boundary. Outside
Git, the file takes effect once its directory belongs to a working tree.

Rules are escaped literally and anchored to the selected file's directory. Existing
contents, comments, permissions, and line-ending style are preserved. A target
symbolic link means the link itself; paths traversing linked directories are rejected
and require the destination's actual path. Symbolic-link ignore files, Git
administrative paths, and names containing line breaks are refused.

`--allow` appends a negated rule, never changes the index, and never rewrites parent
rules. An ignored parent can block an exception: the command writes the requested
rule, identifies Git's blocking rule, and exits with status **1** rather than claiming
the exception works. More-specific rules may also override an entry. Directory
exceptions do not guarantee that every descendant is allowed. Tracked files remain
tracked; the command warns rather than untracking them.

### Global patterns

Outside a working tree, without `--here`, the command offers an escaped **basename**
pattern, such as `build/`, not an absolute-path equivalent. This broader pattern
matches at any depth across **every repository**, not arbitrary filesystem locations.
The exact pattern and destination are shown before confirmation (default **No**).
Declining, invalid input, or EOF leaves files and directories unchanged.

The destination comes from Git's `core.excludesFile` (including Git's `~` expansion),
or defaults to `$XDG_CONFIG_HOME/git/ignore` or `$HOME/.config/git/ignore`. Relative
configured destinations are refused because they resolve differently in each
repository. Git configuration is never changed. Global exceptions cannot override
repository-local exclusions or ignored parents; writing a global pattern does not
claim an effective state in any particular repository.

### Fish completions

```fish
igittigit completions > ~/.config/fish/completions/igittigit.fish
```

Completions include template commands, path-mode flags, and filenames.

## Development

```bash
go test ./...      # Run tests
go run . list      # Run from source
```

## How it works

1. Fetches templates from GitHub's official gitignore API
2. Caches responses locally for 24 hours in `~/.config/igittigit/`
3. Uses `gh` CLI for authenticated requests when available (avoids rate limiting)
4. Falls back to direct HTTP requests gracefully

## License

MIT
