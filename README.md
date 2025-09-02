# igittigit

A minimal Go CLI tool for fetching .gitignore templates from GitHub's API.

## Installation

```bash
go install
```

## Usage

```bash
# List all available templates
igittigit list

# Interactive fuzzy search (requires fzf)
igittigit search

# Get a specific template
igittigit get Go

# Save directly to .gitignore
igittigit get Go > .gitignore
```

## Features

- 🚀 **Fast** - 24-hour file-based caching
- 🔐 **Smart** - Uses `gh` CLI for authenticated requests (no rate limiting)
- 🔍 **Interactive** - `fzf` integration for fuzzy search
- 🛡️ **Safe** - Prompts before overwriting existing .gitignore files
- 🧪 **Tested** - Comprehensive test suite
- 📦 **Minimal** - Only Go standard library, no external dependencies

## Requirements

- Go 1.19+ for installation
- `gh` CLI (optional, for authenticated GitHub API requests)
- `fzf` (optional, for interactive search)

## Examples

```bash
# Browse and select a template interactively
igittigit search

# Get multiple templates
igittigit get Go >> .gitignore
igittigit get Node >> .gitignore

# List templates and filter
igittigit list | grep -i python
```

## How it works

1. Fetches templates from GitHub's official gitignore API
2. Caches responses locally for 24 hours in `~/.config/igittigit/`
3. Uses `gh` CLI for authenticated requests when available
4. Falls back to direct HTTP requests gracefully