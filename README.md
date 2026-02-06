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

./ai-review-analyzer owner/repo
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
    "ai_review_tool": "coderabbit",
    "bot_username": "coderabbitai[bot]",
    "total_prs": 5,
    "ai_suggestion_threads": 23,
    "accepted_suggestions": 8,
    "auto_resolved_by_user": 10,
    "rejected_comments": 5,
    "critical_bugs_fixed": 2,
    "thumbs_up": 8,
    "thumbs_down": 2,
    "avg_developer_feedback_score": 7,
    "prs": [
      {
        "url": "https://github.com/owner/repo/pull/101",
        "author": "dev_user",
        "metrics": {
          "ai_suggestion_threads": 12,
          "accepted_suggestions": 4,
          "auto_resolved_by_user": 6,
          "rejected_comments": 2,
          "critical_bugs_fixed": 1,
          "avg_developer_feedback_score": 7
        },
        "developer_feedback_analyses": [
          {
            "developer_feedback_score": 7,
            "summary": "Developer accepted the suggestion and applied the fix."
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
./ai-review-analyzer RedHatInsights/notifications-backend

# Save output to file
./ai-review-analyzer owner/repo > results.json

# Suppress progress logs
./ai-review-analyzer owner/repo 2>/dev/null
```

## Supported AI Review Tools

| Tool | Bot Username | Status |
|------|--------------|--------|
| CodeRabbit | `coderabbitai[bot]` | Supported |
| Sourcery.ai | `sourcery-ai[bot]` | Supported |

## Metrics

| Metric | Description |
|--------|-------------|
| `ai_suggestion_threads` | Number of review threads started by the AI bot |
| `accepted_suggestions` | Suggestions applied via "Apply Suggestion" button |
| `auto_resolved_by_user` | Issues fixed manually by the developer |
| `rejected_comments` | Unresolved threads on merged PRs |
| `critical_bugs_fixed` | Resolved threads with critical severity (marked with red emoji) |
| `thumbs_up` / `thumbs_down` | Reaction counts on bot comments |
| `developer_feedback_score` | LLM-analyzed score (0-10) based on user discussion |

## Scope

- Analyzes PRs from the past month
- Includes all PR states (open, closed, merged)
- LLM analysis only for threads with user replies
