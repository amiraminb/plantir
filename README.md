# Plantir

A CLI tool to manage GitHub pull requests where you're requested as a reviewer.

> Named after the Palantír - the seeing stones from Lord of the Rings.

## Requirements

- Go 1.21+
- [GitHub CLI](https://cli.github.com/) (`gh`) installed and authenticated

### Setting up GitHub CLI

```bash
# Install (macOS)
brew install gh

# Authenticate
gh auth login
```

Follow the prompts to authenticate via browser or token.

## Installation

```bash
go install github.com/amiraminb/plantir@latest
```

Or build from source:

```bash
git clone https://github.com/amiraminb/plantir.git
cd plantir
go build -o plantir .
```

## Usage

```bash
# List PRs awaiting your review (max 20, newest first)
plantir list

# Filter by repository
plantir list --repo=auth

# Filter by type (feature, dependabot, bot)
plantir list --type=feature
plantir list --type=dependabot

# Show only stale PRs (older than N days)
plantir list --stale=7

# Show more results
plantir list --limit=50
plantir list --limit=0  # unlimited

# JSON output (for scripting)
plantir list --json

# Combine filters
plantir list --repo=auth --type=feature --stale=3

# Open a PR in your browser
plantir open 1234

# Show summary statistics
plantir stats
```

## Commands

| Command | Description |
|---------|-------------|
| `list` | List PRs where you're requested as reviewer |
| `open <PR#>` | Open a PR in your browser |
| `stats` | Show summary statistics by repo and type |

## Flags for `list`

| Flag | Short | Description |
|------|-------|-------------|
| `--repo` | `-r` | Filter by repository name |
| `--type` | `-t` | Filter by PR type (feature, dependabot, bot) |
| `--stale` | `-s` | Only show PRs older than N days |
| `--limit` | `-n` | Max PRs to show (default: 20, 0 for unlimited) |
| `--json` | | Output as JSON |
