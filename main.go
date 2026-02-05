package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"ai-review-analyzer/coderabbit"
	"ai-review-analyzer/sourcery"

	"github.com/google/go-github/v80/github"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <github-repo-url>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Analyzes AI-assisted code review tools usage in a GitHub repository.\n\n")
		fmt.Fprintf(os.Stderr, "Environment variables:\n")
		fmt.Fprintf(os.Stderr, "  GITHUB_TOKEN  GitHub Personal Access Token (required)\n\n")
		fmt.Fprintf(os.Stderr, "Example:\n")
		fmt.Fprintf(os.Stderr, "  %s https://github.com/owner/repo\n", os.Args[0])
	}

	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	repoURL := flag.Arg(0)
	owner, repo, err := parseGitHubURL(repoURL)
	if err != nil {
		log.Fatalf("Invalid GitHub URL: %v", err)
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN environment variable is required")
	}

	log.Printf("Analyzing repository: %s/%s", owner, repo)

	client := github.NewClient(nil).WithAuthToken(token)
	ctx := context.Background()

	// Calculate date range (past month)
	since := time.Now().AddDate(0, -1, 0)
	log.Printf("Fetching PRs since: %s", since.Format("2006-01-02"))

	// Fetch and analyze PRs
	results, err := analyzePRs(ctx, client, owner, repo, since)
	if err != nil {
		log.Fatalf("Failed to analyze PRs: %v", err)
	}

	// Output JSON to stdout
	output, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal results: %v", err)
	}
	fmt.Println(string(output))
}

// parseGitHubURL extracts owner and repo from a GitHub URL
func parseGitHubURL(rawURL string) (owner, repo string, err error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse URL: %w", err)
	}

	if parsed.Host != "github.com" && parsed.Host != "www.github.com" {
		return "", "", fmt.Errorf("not a GitHub URL: %s", parsed.Host)
	}

	// Path should be /owner/repo or /owner/repo/...
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid repository path: %s", parsed.Path)
	}

	owner = parts[0]
	repo = strings.TrimSuffix(parts[1], ".git")

	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("owner or repo is empty")
	}

	return owner, repo, nil
}

// analyzePRs fetches PRs and analyzes AI review tool comments
func analyzePRs(ctx context.Context, client *github.Client, owner, repo string, since time.Time) ([]any, error) {
	// Initialize analyzers
	crAnalyzer := coderabbit.NewAnalyzer(client, owner, repo)
	srcAnalyzer := sourcery.NewAnalyzer(client, owner, repo)

	// Fetch all PRs (all states) with pagination
	opts := &github.PullRequestListOptions{
		State:     "all",
		Sort:      "updated",
		Direction: "desc",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	prCount := 0
	for {
		prs, resp, err := client.PullRequests.List(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list PRs: %w", err)
		}

		for _, pr := range prs {
			// Skip PRs older than our time window
			if pr.UpdatedAt.Before(since) {
				log.Printf("Reached PRs older than %s, stopping", since.Format("2006-01-02"))
				goto done
			}

			prCount++
			prNum := pr.GetNumber()

			// Check if any AI bot reviewed this PR
			reviews, _, err := client.PullRequests.ListReviews(ctx, owner, repo, prNum, nil)
			if err != nil {
				log.Printf("Warning: failed to get reviews for PR #%d: %v", prNum, err)
				continue
			}

			hasCodeRabbit := crAnalyzer.CheckReview(reviews)
			hasSourcery := srcAnalyzer.CheckReview(reviews)

			// If no AI bot reviewed this PR, skip detailed analysis
			if !hasCodeRabbit && !hasSourcery {
				continue
			}

			log.Printf("PR #%d has AI reviews (CodeRabbit: %v, Sourcery: %v)", prNum, hasCodeRabbit, hasSourcery)

			// Analyze CodeRabbit comments
			if hasCodeRabbit {
				crAnalyzer.MarkPRReviewed(prNum)
				if err := crAnalyzer.AnalyzeComments(ctx, prNum); err != nil {
					log.Printf("Warning: failed to analyze CodeRabbit comments for PR #%d: %v", prNum, err)
				}
				if err := crAnalyzer.AnalyzeIssueComments(ctx, prNum); err != nil {
					log.Printf("Warning: failed to analyze CodeRabbit issue comments for PR #%d: %v", prNum, err)
				}
			}

			// Analyze Sourcery comments
			if hasSourcery {
				srcAnalyzer.MarkPRReviewed(prNum)
				if err := srcAnalyzer.AnalyzeComments(ctx, prNum); err != nil {
					log.Printf("Warning: failed to analyze Sourcery comments for PR #%d: %v", prNum, err)
				}
				if err := srcAnalyzer.AnalyzeIssueComments(ctx, prNum); err != nil {
					log.Printf("Warning: failed to analyze Sourcery issue comments for PR #%d: %v", prNum, err)
				}
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

done:
	log.Printf("Analyzed %d PRs total", prCount)

	// Build results array (only include tools that reviewed at least one PR)
	var results []any
	if crAnalyzer.HasReviews() {
		results = append(results, crAnalyzer.GetResult())
	}
	if srcAnalyzer.HasReviews() {
		results = append(results, srcAnalyzer.GetResult())
	}

	// If no AI tools found, return empty array
	if results == nil {
		results = []any{}
	}

	return results, nil
}
