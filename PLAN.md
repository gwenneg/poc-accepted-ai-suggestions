# PoC Plan

## Overview

A CLI tool written in GoLang that analyzes GitHub repositories to measure the effectiveness of AI-assisted code review tools. It detects comments from tools like Sourcery.ai and CodeRabbit, then determines whether their suggestions were accepted or rejected by developers.

## Requirements

### Input
- GitHub repository URL (e.g., `https://github.com/owner/repo`)

### Output
- JSON structure containing:
  - Which AI review tools commented on PRs
  - Number of comments per tool
  - Number of accepted suggestions per tool
  - Acceptance rate

### Functional Requirements
1. Parse GitHub repo URL to extract owner/repo
2. Authenticate with GitHub API
3. Fetch PR data including comments
4. Identify comments from known AI review tools
5. Determine acceptance/rejection status of each suggestion
6. Aggregate statistics and output JSON

## Known AI Review Tools to Detect

| Tool | GitHub Username/Bot | Notes |
|------|---------------------|-------|
| Sourcery.ai | `sourcery-ai[bot]` | TBD - need to verify |
| CodeRabbit | `coderabbitai[bot]` | TBD - need to verify |
| Others? | TBD | |

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐
│  CLI Input  │────▶│  GitHub API  │────▶│  Comment Parser │
│  (repo URL) │     │    Client    │     │  (detect AI)    │
└─────────────┘     └──────────────┘     └────────┬────────┘
                                                  │
                                                  ▼
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐
│ JSON Output │◀────│  Aggregator  │◀────│   Acceptance    │
│             │     │              │     │    Analyzer     │
└─────────────┘     └──────────────┘     └─────────────────┘
```

## Implementation Steps

### Phase 1: Foundation
- [ ] Initialize Go module (go 1.24.0)
- [ ] Set up CLI with standard `flag` package
- [ ] Implement GitHub URL parser
- [ ] Set up GitHub API client with go-github/v80

### Phase 2: Data Fetching
- [ ] Fetch list of PRs from repository
- [ ] Fetch comments for each PR (review comments + issue comments)
- [ ] Handle pagination for large repos

### Phase 3: AI Comment Detection
- [ ] Build registry of known AI tool bot usernames
- [ ] Filter comments to identify AI-generated ones
- [ ] Parse/categorize AI comments (suggestions vs general feedback)

### Phase 4: Acceptance Detection (Core Challenge)
- [ ] Implement acceptance detection logic (see Open Questions)
- [ ] Handle edge cases

### Phase 5: Output
- [ ] Aggregate statistics
- [ ] Generate JSON output
- [ ] Add CLI flags for verbosity/filtering

## Open Questions

### 1. How do we determine if an AI suggestion was accepted?

This is the core challenge. Possible signals to consider:

**Strong signals:**
- Suggestion was applied via GitHub's "Apply suggestion" button (creates a commit with specific message format)
- Comment thread was marked as "resolved"
- Subsequent commit changes the exact lines mentioned in the suggestion

**Weaker signals:**
- PR author replied with positive sentiment ("thanks", "fixed", "good catch")
- The suggested code appears in a later commit (fuzzy matching)
- Comment received a thumbs-up reaction

**Potential approaches:**
1. **Commit message analysis**: Look for "Co-authored-by" or "Apply suggestions from code review" patterns
2. **Diff analysis**: Compare suggested code changes with actual commits made after the comment
3. **Thread resolution status**: Check if the comment thread was resolved
4. **Hybrid approach**: Combine multiple signals with confidence scoring

### 2. What GitHub API endpoints do we need?

- `GET /repos/{owner}/{repo}/pulls` - List PRs
- `GET /repos/{owner}/{repo}/pulls/{pull_number}/comments` - Review comments
- `GET /repos/{owner}/{repo}/issues/{issue_number}/comments` - Issue comments
- `GET /repos/{owner}/{repo}/pulls/{pull_number}/commits` - Commits on PR
- Others?

### ~~3. Authentication approach?~~ DECIDED

See Decisions Made section.

### 4. Rate limiting strategy?

- GitHub API has rate limits (5000 requests/hour for authenticated)
- Need to handle pagination efficiently
- Consider caching?

### 5. Scope for PoC?

- All PRs or just recent/merged ones?
- All comments or just those with actionable suggestions?
- Time range filter?

## Decisions Made

1. **Authentication**: Use `GITHUB_TOKEN` environment variable for GitHub API authentication
2. **Go version**: 1.24.0
3. **GitHub API library**: `github.com/google/go-github/v80`

## Notes

- Focus on PoC quality over production readiness
- Start with a simple acceptance heuristic and iterate
