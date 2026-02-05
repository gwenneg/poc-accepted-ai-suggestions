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
├── coderabbit/             # CodeRabbit package
│   └── coderabbit.go
├── sourcery/               # Sourcery.ai package
│   └── sourcery.go
├── llm/                    # Claude LLM client
│   └── claude.go
├── .env.example
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

**REST API (current implementation):**
1. `GET /repos/{owner}/{repo}/pulls` - List PRs from past month
2. `GET /repos/{owner}/{repo}/pulls/{pull_number}/reviews` - Check if bot is in reviewers
3. `GET /repos/{owner}/{repo}/pulls/{pull_number}/comments` - Get comments (only for PRs with AI reviews)

**GraphQL API (future - for bot-resolved detection):**
- REST API does not expose `resolved` or `resolvedBy` fields for review comments
- GraphQL API provides `reviewThreads` with `isResolved` and `resolvedBy` fields
- Will be added later to detect threads resolved by the bot itself

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

**Note:** Sourcery adds 1 👍 and 1 👎 by default to each review comment (for user feedback). These defaults are subtracted when counting reactions.

**Metrics API:** None available

## LLM Analysis Feature (Claude)

For each discussion thread initiated by an AI review tool, submit all comments to Claude for analysis.

**Input:** All comments from a review thread (bot suggestion + user replies)

**Condition:** Only submit threads that have at least one user comment (skip threads with only the bot's initial suggestion)

**Output:** JSON structure with:
```json
{
  "score": 75,
  "summary": "The developer acknowledged the suggestion was valid but noted it would require refactoring other parts of the codebase. They planned to address it in a follow-up PR."
}
```

**Fields:**
- `score` (0-100): How useful the AI suggestion was based on user discussion
  - 0-25: Rejected/unhelpful suggestion
  - 26-50: Partially useful but not applied
  - 51-75: Useful and likely applied
  - 76-100: Very useful and clearly applied
- `summary`: Brief summary of user discussions, reflecting their opinion and intent

**Implementation approach:**
- Simple HTTP client to Claude API (inspired by [release-confidence-score](https://github.com/RedHatInsights/release-confidence-score/tree/main/internal/llm))
- Single `llm/` package with Claude client only
- Environment variables: `MODEL_API` (base URL), `MODEL_USER_KEY` (Bearer token)
- Model ID: hardcoded (Claude Sonnet)

**Cost optimization:**
- Only include message body from each comment (no usernames, timestamps, or metadata)
- Skip threads without user comments
- Keep prompts minimal and focused

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
- [ ] Implement reactions detection (👍/👎)
- [ ] (Future) Implement bot-resolved threads detection via GraphQL API

### Phase 4: LLM Analysis
- [ ] Implement Claude API client in `llm/claude.go`
- [ ] Fetch full thread comments for each bot-initiated discussion
- [ ] Submit thread comments to Claude for analysis
- [ ] Parse Claude response (score + summary)
- [ ] Include LLM analysis in output

### Phase 5: Output
- [ ] Generate JSON array output
- [ ] Create README.md (usage-focused)

## Decisions

| Decision | Value |
|----------|-------|
| CLI tool name | `ai-review-analyzer` |
| GitHub authentication | `GITHUB_TOKEN` env var |
| LLM configuration | `MODEL_API` (base URL) + `MODEL_USER_KEY` (Bearer token) |
| Go version | 1.24.0 |
| GitHub API library | `github.com/google/go-github/v80` |
| LLM provider | Claude only (via Anthropic API) |
| LLM model | Claude Sonnet |
| Initial scope | CodeRabbit + Sourcery.ai only |
| Project structure | Separate Go package per tool |
| Time window | Past month (hardcoded, configurable later) |
| PR state filter | All PRs (open, closed, merged) |
| Output format | JSON array, tool-specific structures |
| Error handling | Log and skip PR, report partial results |
| Logging | Simple progress logs to stderr |

## Notes

- Focus on PoC quality over production readiness
- Handle GitHub API rate limits gracefully (5000 requests/hour authenticated)
