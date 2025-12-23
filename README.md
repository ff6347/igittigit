# igittigit

A minimal CLI for fetching .gitignore templates from GitHub's API.

## Prerequisites

- [Go](https://go.dev/) 1.19+ (for installation)
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
