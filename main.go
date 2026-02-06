package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"ai-review-analyzer/coderabbit"
	"ai-review-analyzer/ghgql"
	"ai-review-analyzer/github"
	"ai-review-analyzer/llm"
	"ai-review-analyzer/sourcery"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <owner/repo>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Analyzes AI-assisted code review tools usage in a GitHub repository.\n\n")
		fmt.Fprintf(os.Stderr, "Environment variables:\n")
		fmt.Fprintf(os.Stderr, "  GITHUB_TOKEN     GitHub Personal Access Token (required)\n")
		fmt.Fprintf(os.Stderr, "  MODEL_API        Base URL for Claude API (required)\n")
		fmt.Fprintf(os.Stderr, "  MODEL_USER_KEY   Bearer token for Claude API (required)\n\n")
		fmt.Fprintf(os.Stderr, "Example:\n")
		fmt.Fprintf(os.Stderr, "  %s owner/repo\n", os.Args[0])
	}

	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	repoArg := flag.Arg(0)
	owner, repo, err := parseRepoArg(repoArg)
	if err != nil {
		log.Fatalf("Invalid repository: %v", err)
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN environment variable is required")
	}

	modelAPI := os.Getenv("MODEL_API")
	if modelAPI == "" {
		log.Fatal("MODEL_API environment variable is required")
	}

	modelUserKey := os.Getenv("MODEL_USER_KEY")
	if modelUserKey == "" {
		log.Fatal("MODEL_USER_KEY environment variable is required")
	}

	log.Printf("Analyzing repository: %s/%s", owner, repo)

	// Create clients
	restClient := github.NewClient(token)
	gqlClient := ghgql.NewClient(token)
	llmClient := llm.NewClient(modelAPI, modelUserKey)
	ctx := context.Background()

	// Calculate date range (past month)
	since := time.Now().AddDate(0, -1, 0)
	log.Printf("Fetching PRs since: %s", since.Format("2006-01-02"))

	// Fetch all PRs via REST API (with pagination)
	prs, err := restClient.FetchPRs(ctx, owner, repo, since)
	if err != nil {
		log.Fatalf("Failed to fetch PRs: %v", err)
	}
	log.Printf("Fetched %d PRs", len(prs))

	// Initialize analyzers
	crAnalyzer := coderabbit.NewAnalyzer(llmClient)
	srcAnalyzer := sourcery.NewAnalyzer(llmClient)

	// Analyze each PR
	for _, pr := range prs {
		// Fetch reviewers via REST API (with pagination)
		reviewers, err := restClient.FetchReviewers(ctx, owner, repo, pr.Number)
		if err != nil {
			log.Printf("Warning: failed to fetch reviewers for PR #%d: %v", pr.Number, err)
			continue
		}

		hasCodeRabbit := containsBot(reviewers, coderabbit.BotUsername)
		hasSourcery := containsBot(reviewers, sourcery.BotUsername)

		if !hasCodeRabbit && !hasSourcery {
			continue
		}

		log.Printf("PR #%d has AI reviews (CodeRabbit: %v, Sourcery: %v)", pr.Number, hasCodeRabbit, hasSourcery)

		// Fetch commits via REST API (with pagination)
		commitMsgs, err := restClient.FetchCommitMessages(ctx, owner, repo, pr.Number)
		if err != nil {
			log.Printf("Warning: failed to fetch commits for PR #%d: %v", pr.Number, err)
			commitMsgs = nil
		}

		// Fetch review threads via GraphQL (with pagination)
		threads, err := gqlClient.FetchReviewThreads(ctx, owner, repo, pr.Number)
		if err != nil {
			log.Printf("Warning: failed to fetch threads for PR #%d: %v", pr.Number, err)
			threads = nil
		}

		// Build combined PR data for analyzers
		prData := ghgql.PRData{
			Number:     pr.Number,
			URL:        pr.URL,
			Author:     pr.Author,
			IsMerged:   pr.IsMerged,
			CommitMsgs: commitMsgs,
			Threads:    threads,
		}

		if hasCodeRabbit {
			crAnalyzer.AnalyzePR(ctx, prData)
		}

		if hasSourcery {
			srcAnalyzer.AnalyzePR(ctx, prData)
		}
	}

	// Build results array
	var results []any
	if crAnalyzer.HasReviews() {
		results = append(results, crAnalyzer.GetResult())
	}
	if srcAnalyzer.HasReviews() {
		results = append(results, srcAnalyzer.GetResult())
	}

	if results == nil {
		results = []any{}
	}

	log.Printf("Analysis complete")

	// Output JSON to stdout
	output, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal results: %v", err)
	}
	fmt.Println(string(output))
}


// containsBot checks if the bot username is in the reviewers list
func containsBot(reviewers []string, botUsername string) bool {
	// REST API returns full username with [bot] suffix
	for _, r := range reviewers {
		if r == botUsername {
			return true
		}
	}
	return false
}

// parseRepoArg extracts owner and repo from an "owner/repo" argument
func parseRepoArg(arg string) (owner, repo string, err error) {
	parts := strings.Split(arg, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected format: owner/repo")
	}

	owner = strings.TrimSpace(parts[0])
	repo = strings.TrimSpace(parts[1])

	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("owner or repo is empty")
	}

	return owner, repo, nil
}
