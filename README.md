# ai-review-analyzer

A CLI tool that analyzes GitHub repositories to measure the effectiveness of AI-assisted code review tools (CodeRabbit, Sourcery.ai).

## Requirements

- Go 1.24.0+
- GitHub Personal Access Token
- Claude API access (via Vertex AI or compatible endpoint)

## Installation

```bash
go build -o ai-review-analyzer .
```

## Usage

```bash
export GITHUB_TOKEN=your_github_token
export MODEL_API=https://your-claude-endpoint.example.com
export MODEL_USER_KEY=your_api_key

./ai-review-analyzer https://github.com/owner/repo
```

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `GITHUB_TOKEN` | Yes | GitHub Personal Access Token with `repo` scope (private repos) or `public_repo` scope (public repos only) |
| `MODEL_API` | Yes | Base URL for Claude API endpoint |
| `MODEL_USER_KEY` | Yes | Bearer token for Claude API authentication |

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
    "thumbs_up": 4,
    "thumbs_down": 3,
    "avg_llm_score": 68.5,
    "details": [
      {
        "pr_number": 123,
        "comments": 5,
        "thumbs_up": 2,
        "thumbs_down": 1,
        "thread_analyses": [
          {
            "score": 75,
            "summary": "User agreed with the suggestion and applied the fix."
          }
        ]
      }
    ]
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
- LLM analysis only for threads with user replies
