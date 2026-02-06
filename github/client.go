package github

import (
	"context"
	"time"

	"github.com/google/go-github/v80/github"
)

// Client wraps the GitHub REST API client
type Client struct {
	client *github.Client
}

// NewClient creates a new REST API client with the given token
func NewClient(token string) *Client {
	return &Client{
		client: github.NewClient(nil).WithAuthToken(token),
	}
}

// PR represents a pull request with basic data from REST API
type PR struct {
	Number   int
	URL      string
	Author   string
	IsMerged bool
}

// FetchPRs fetches all PRs updated since the given time with pagination
func (c *Client) FetchPRs(ctx context.Context, owner, repo string, since time.Time) ([]PR, error) {
	var allPRs []PR

	opts := &github.PullRequestListOptions{
		State:     "all",
		Sort:      "updated",
		Direction: "desc",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		prs, resp, err := c.client.PullRequests.List(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}

		for _, pr := range prs {
			// Stop if PR is older than our time window
			if pr.UpdatedAt != nil && pr.UpdatedAt.Before(since) {
				return allPRs, nil
			}

			allPRs = append(allPRs, PR{
				Number:   pr.GetNumber(),
				URL:      pr.GetHTMLURL(),
				Author:   pr.GetUser().GetLogin(),
				IsMerged: !pr.GetMergedAt().IsZero(),
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allPRs, nil
}

// FetchReviewers fetches all reviewer usernames for a PR with pagination
func (c *Client) FetchReviewers(ctx context.Context, owner, repo string, prNumber int) ([]string, error) {
	var reviewers []string
	seen := make(map[string]bool)

	opts := &github.ListOptions{PerPage: 100}

	for {
		reviews, resp, err := c.client.PullRequests.ListReviews(ctx, owner, repo, prNumber, opts)
		if err != nil {
			return nil, err
		}

		for _, review := range reviews {
			login := review.GetUser().GetLogin()
			if !seen[login] {
				seen[login] = true
				reviewers = append(reviewers, login)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return reviewers, nil
}

// FetchCommitMessages fetches all commit messages for a PR with pagination
func (c *Client) FetchCommitMessages(ctx context.Context, owner, repo string, prNumber int) ([]string, error) {
	var messages []string

	opts := &github.ListOptions{PerPage: 100}

	for {
		commits, resp, err := c.client.PullRequests.ListCommits(ctx, owner, repo, prNumber, opts)
		if err != nil {
			return nil, err
		}

		for _, commit := range commits {
			if commit.Commit != nil {
				messages = append(messages, commit.Commit.GetMessage())
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return messages, nil
}
