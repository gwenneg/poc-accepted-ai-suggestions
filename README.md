# ai-review-analyzer

A CLI tool that analyzes GitHub repositories to measure the effectiveness of AI-assisted code review tools (CodeRabbit, Sourcery.ai).

## Requirements

- Go 1.24.0+
- GitHub Personal Access Token

## Installation

```bash
go build -o ai-review-analyzer .
```

## Usage

```bash
export GITHUB_TOKEN=your_github_token_here
./ai-review-analyzer https://github.com/owner/repo
```

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `GITHUB_TOKEN` | Yes | GitHub Personal Access Token with `repo` scope (private repos) or `public_repo` scope (public repos only) |

### Output

The tool outputs a JSON array to stdout with analysis results for each AI review tool found:

```json
[
  {
    "tool": "sourcery",
    "bot_username": "sourcery-ai[bot]",
    "prs_reviewed": 34,
    "total_comments": 50,
    "bot_resolved_comments": 0,
    "thumbs_up": 54,
    "thumbs_down": 53
  }
]
```

Progress logs are written to stderr.

### Example

```bash
# Analyze a repository
./ai-review-analyzer https://github.com/RedHatInsights/notifications-backend

# Save output to file
./ai-review-analyzer https://github.com/owner/repo > results.json

# Suppress progress logs
./ai-review-analyzer https://github.com/owner/repo 2>/dev/null
```

## Supported AI Review Tools

| Tool | Bot Username | Status |
|------|--------------|--------|
| CodeRabbit | `coderabbitai[bot]` | Supported |
| Sourcery.ai | `sourcery-ai[bot]` | Supported |

## Scope

- Analyzes PRs from the past month
- Includes all PR states (open, closed, merged)
