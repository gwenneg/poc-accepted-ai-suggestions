# PoC Plan

## Overview

A CLI tool written in GoLang that analyzes GitHub repositories to measure the effectiveness of AI-assisted code review tools. It detects comments from CodeRabbit and Sourcery.ai, then determines whether their suggestions were accepted or rejected by developers.

## Requirements

### Input
- GitHub repository URL (e.g., `https://github.com/owner/repo`)

### Output
- JSON array where each item represents one AI review tool's analysis
- Each tool has its own data structure (not unified across tools)
- Example:
```json
[
  {
    "tool": "coderabbit",
    "bot_username": "coderabbitai[bot]",
    "prs_reviewed": 5,
    "total_comments": 23,
    "bot_resolved_comments": 12,
    "thumbs_up": 8,
    "thumbs_down": 2
  },
  {
    "tool": "sourcery",
    "bot_username": "sourcery-ai[bot]",
    "prs_reviewed": 3,
    "total_comments": 12,
    "bot_resolved_comments": 6,
    "thumbs_up": 5,
    "thumbs_down": 1
  }
]
```

## Project Structure

```
/
├── main.go                 # CLI entry point
├── go.mod
├── coderabbit/             # CodeRabbit module
│   ├── go.mod
│   └── coderabbit.go
├── sourcery/               # Sourcery.ai module
│   ├── go.mod
│   └── sourcery.go
└── README.md
```

## AI Review Tools

| Tool | Bot Username | Status |
|------|--------------|--------|
| CodeRabbit | `coderabbitai[bot]` | **Phase 1** |
| Sourcery.ai | `sourcery-ai[bot]` | **Phase 1** |
| GitHub Copilot | `copilot[bot]` | Future |
| Codium/Qodo | `qodo-merge-pro[bot]` | Future |
| Ellipsis | `ellipsis-dev[bot]` | Future |
| Bito AI | `bito-ai[bot]` | Future |
| Sweep AI | `sweep-ai[bot]` | Future |
| What The Diff | `whatthediff[bot]` | Future |

*Bot usernames need verification against actual PRs*

## GitHub API Strategy

1. `GET /repos/{owner}/{repo}/pulls` - List PRs from past month
2. `GET /repos/{owner}/{repo}/pulls/{pull_number}/reviews` - Check if bot is in reviewers
3. `GET /repos/{owner}/{repo}/pulls/{pull_number}/comments` - Get comments (only for PRs with AI reviews)

## Tool-Specific Details

### CodeRabbit

**Acceptance detection:**
- Thread resolved **by the bot itself** = AI detected issue was fixed
- 👍/👎 reactions on comments

**Future enhancement:** CodeRabbit Metrics API (Enterprise plan only) at `https://api.coderabbit.ai/v1/metrics/reviews`

### Sourcery.ai

**Acceptance detection:**
- Thread resolved **by the bot itself** = AI detected issue was fixed
- 👍/👎 reactions on comments

**Metrics API:** None available

## Implementation Steps

### Phase 1: Foundation
- [ ] Initialize Go module (go 1.24.0)
- [ ] Set up CLI with standard `flag` package
- [ ] Implement GitHub URL parser
- [ ] Set up GitHub API client with go-github/v80

### Phase 2: Data Fetching
- [ ] Fetch PRs from past month
- [ ] Filter PRs by checking reviews for bot usernames
- [ ] Fetch comments for matching PRs
- [ ] Handle pagination

### Phase 3: Analysis
- [ ] Implement acceptance detection (bot-resolved threads + reactions)

### Phase 4: Output
- [ ] Generate JSON array output
- [ ] Create README.md (usage-focused)

## Decisions

| Decision | Value |
|----------|-------|
| CLI tool name | `ai-review-analyzer` |
| Authentication | `GITHUB_TOKEN` env var |
| Go version | 1.24.0 |
| GitHub API library | `github.com/google/go-github/v80` |
| Initial scope | CodeRabbit + Sourcery.ai only |
| Project structure | Separate Go module per tool |
| Time window | Past month (hardcoded, configurable later) |
| PR state filter | All PRs (open, closed, merged) |
| Output format | JSON array, tool-specific structures |
| Error handling | Log and skip PR, report partial results |
| Logging | Simple progress logs to stderr |

## Notes

- Focus on PoC quality over production readiness
- Handle GitHub API rate limits gracefully (5000 requests/hour authenticated)
