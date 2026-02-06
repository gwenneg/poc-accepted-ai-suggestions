# PoC Plan

## Overview

A CLI tool written in Go that analyzes GitHub repositories to measure the effectiveness of AI-assisted code review tools (CodeRabbit, Sourcery.ai). It detects bot comments and determines whether suggestions were accepted or rejected by developers.

## Input/Output

**Input:** `owner/repo` (e.g., `RedHatInsights/notifications-backend`)

**Output:** JSON array with analysis per AI tool:
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
          {"developer_feedback_score": 7, "summary": "Developer accepted the suggestion..."}
        ]
      }
    ]
  }
]
```

## Project Structure

```
├── main.go              # CLI entry point
├── github/              # REST API client (PRs, reviews, commits)
├── ghgql/               # GraphQL client (review threads)
├── coderabbit/          # CodeRabbit analyzer
├── sourcery/            # Sourcery.ai analyzer
├── llm/                 # Claude API client + system prompt
└── README.md
```

## Supported AI Tools

| Tool | Bot Username |
|------|--------------|
| CodeRabbit | `coderabbitai[bot]` |
| Sourcery.ai | `sourcery-ai[bot]` |

## GitHub API Strategy

**REST API** (`go-github/v80`): PRs, reviews, commits - with full pagination

**GraphQL API** (`githubv4`): Review threads per PR - for `isResolved`, `isOutdated`, `resolvedBy`, reactions

Why hybrid: REST handles pagination properly; GraphQL provides thread resolution data not available via REST. Per-PR queries avoid the 500k node limit.

## Metrics

| Metric | Detection Method |
|--------|------------------|
| `ai_suggestion_threads` | Count bot-authored review threads |
| `accepted_suggestions` | Commits with "Apply suggestion from [bot]" message |
| `auto_resolved_by_user` | `meaningful - accepted` |
| `rejected_comments` | Unresolved bot threads on merged PRs |
| `critical_bugs_fixed` | Resolved threads with 🔴 emoji |
| `thumbs_up/down` | Reaction counts (Sourcery: subtract 1 default each) |

**Thread resolution logic:**
- Meaningful: `isResolved && !isOutdated`
- False positive: `isResolved && isOutdated`

## LLM Analysis (Claude)

Threads with user replies are sent to Claude for scoring.

**Output:**
```json
{"developer_feedback_score": 7, "summary": "..."}
```

**Score (0-10):** 0-2 rejected, 3-5 partial, 6-8 applied, 9-10 clearly applied

**Cost optimization:** Only message bodies, skip threads without replies.

## Configuration

| Variable | Description |
|----------|-------------|
| `GITHUB_TOKEN` | GitHub PAT |
| `MODEL_API` | Claude API base URL |
| `MODEL_USER_KEY` | Claude API bearer token |

## Scope

- PRs from past month (hardcoded)
- All PR states (open, closed, merged)
- Log errors, skip PR, report partial results
