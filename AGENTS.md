# igittigit

A minimal Go CLI tool for fetching .gitignore templates from GitHub's API.

## Installation

```bash
go install
```

This installs `igittigit` to `$GOPATH/bin` and makes it globally available.

## Usage

- `igittigit list` - List all available templates
- `igittigit search` - Interactive fuzzy search (requires fzf)
- `igittigit get <template>` - Fetch specific template
- `igittigit get Go > .gitignore` - Save to file

## Features

✅ **Implemented**
- 24-hour file-based caching in `~/.config/igittigit/`
- `gh` CLI integration for authenticated requests (avoids rate limiting)
- `fzf` integration for interactive template search
- Smart .gitignore handling (prompts for overwrite/append when exists)
- Graceful fallbacks when external tools unavailable
- Comprehensive test suite (28.9% coverage)

## Architecture

### Core Components
- `main.go`: CLI entry point with commands and interactive features
- `cache.go`: File-based caching system
- `github.go`: HTTP client with `gh` CLI integration and fallback
- `*_test.go`: Test suites for all components

### Technical Details
- Uses only Go standard library
- Cache structure: `{"templates": [...], "last_updated": "timestamp"}`
- Detects `gh` and `fzf` via `exec.LookPath()`
- Interactive prompts with input validation

### GitHub API Endpoints
```
GET https://api.github.com/gitignore/templates
GET https://api.github.com/gitignore/templates/NAME
```

Headers: 
- Accept: application/vnd.github+json
- X-GitHub-Api-Version: 2022-11-28