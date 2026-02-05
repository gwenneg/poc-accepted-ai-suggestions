package sourcery

import (
	"context"
	"log"

	"github.com/google/go-github/v80/github"
)

const BotUsername = "sourcery-ai[bot]"

// Result represents the analysis result for Sourcery
type Result struct {
	Tool                string `json:"tool"`
	BotUsername         string `json:"bot_username"`
	PRsReviewed         int    `json:"prs_reviewed"`
	TotalComments       int    `json:"total_comments"`
	BotResolvedComments int    `json:"bot_resolved_comments"`
	ThumbsUp            int    `json:"thumbs_up"`
	ThumbsDown          int    `json:"thumbs_down"`
}

// Analyzer handles Sourcery-specific analysis
type Analyzer struct {
	client *github.Client
	owner  string
	repo   string
	result Result
	prs    map[int]bool
}

// NewAnalyzer creates a new Sourcery analyzer
func NewAnalyzer(client *github.Client, owner, repo string) *Analyzer {
	return &Analyzer{
		client: client,
		owner:  owner,
		repo:   repo,
		result: Result{
			Tool:        "sourcery",
			BotUsername: BotUsername,
		},
		prs: make(map[int]bool),
	}
}

// CheckReview checks if Sourcery reviewed a PR based on the reviews list
func (a *Analyzer) CheckReview(reviews []*github.PullRequestReview) bool {
	for _, review := range reviews {
		if review.User != nil && review.User.GetLogin() == BotUsername {
			return true
		}
	}
	return false
}

// MarkPRReviewed marks a PR as reviewed by Sourcery
func (a *Analyzer) MarkPRReviewed(prNum int) {
	a.prs[prNum] = true
}

// AnalyzeComments analyzes review comments for a PR
func (a *Analyzer) AnalyzeComments(ctx context.Context, prNum int) error {
	opts := &github.PullRequestListCommentsOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		comments, resp, err := a.client.PullRequests.ListComments(ctx, a.owner, a.repo, prNum, opts)
		if err != nil {
			return err
		}

		for _, comment := range comments {
			if comment.User == nil || comment.User.GetLogin() != BotUsername {
				continue
			}

			a.result.TotalComments++

			// Sourcery adds 1 thumbs up and 1 thumbs down by default to each review comment.
			// We subtract these defaults to get actual user reactions.
			if comment.Reactions != nil {
				up := comment.Reactions.GetPlusOne() - 1
				down := comment.Reactions.GetMinusOne() - 1
				if up > 0 {
					a.result.ThumbsUp += up
				}
				if down > 0 {
					a.result.ThumbsDown += down
				}
			}

			// TODO: Bot-resolved detection requires GraphQL API
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return nil
}

// AnalyzeIssueComments analyzes issue comments (general PR comments) for a PR
func (a *Analyzer) AnalyzeIssueComments(ctx context.Context, prNum int) error {
	opts := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		comments, resp, err := a.client.Issues.ListComments(ctx, a.owner, a.repo, prNum, opts)
		if err != nil {
			return err
		}

		for _, comment := range comments {
			if comment.User == nil || comment.User.GetLogin() != BotUsername {
				continue
			}

			a.result.TotalComments++

			if comment.Reactions != nil {
				a.result.ThumbsUp += comment.Reactions.GetPlusOne()
				a.result.ThumbsDown += comment.Reactions.GetMinusOne()
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return nil
}

// GetResult returns the final analysis result
func (a *Analyzer) GetResult() Result {
	a.result.PRsReviewed = len(a.prs)
	return a.result
}

// HasReviews returns true if Sourcery reviewed any PRs
func (a *Analyzer) HasReviews() bool {
	return len(a.prs) > 0
}

// LogAnalysis logs the start of PR analysis
func (a *Analyzer) LogAnalysis(prNum int) {
	log.Printf("  Sourcery: analyzing PR #%d", prNum)
}
