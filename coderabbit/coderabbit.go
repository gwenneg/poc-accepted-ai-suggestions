package coderabbit

import (
	"context"
	"log"

	"ai-review-analyzer/llm"

	"github.com/google/go-github/v80/github"
)

const BotUsername = "coderabbitai[bot]"

// ThreadAnalysis represents an analyzed discussion thread
type ThreadAnalysis struct {
	Score   int    `json:"score"`
	Summary string `json:"summary"`
}

// PRDetail represents per-PR metrics and analysis
type PRDetail struct {
	PRNumber       int              `json:"pr_number"`
	Comments       int              `json:"comments"`
	ThumbsUp       int              `json:"thumbs_up"`
	ThumbsDown     int              `json:"thumbs_down"`
	ThreadAnalyses []ThreadAnalysis `json:"thread_analyses,omitempty"`
}

// Result represents the analysis result for CodeRabbit
type Result struct {
	Tool                string     `json:"tool"`
	BotUsername         string     `json:"bot_username"`
	PRsReviewed         int        `json:"prs_reviewed"`
	TotalComments       int        `json:"total_comments"`
	BotResolvedComments int        `json:"bot_resolved_comments"`
	ThumbsUp            int        `json:"thumbs_up"`
	ThumbsDown          int        `json:"thumbs_down"`
	AvgLLMScore         *float64   `json:"avg_llm_score,omitempty"`
	Details             []PRDetail `json:"details,omitempty"`
}

// Analyzer handles CodeRabbit-specific analysis
type Analyzer struct {
	client    *github.Client
	llmClient *llm.Client
	owner     string
	repo      string
	result    Result
	prDetails map[int]*PRDetail
}

// NewAnalyzer creates a new CodeRabbit analyzer
func NewAnalyzer(client *github.Client, llmClient *llm.Client, owner, repo string) *Analyzer {
	return &Analyzer{
		client:    client,
		llmClient: llmClient,
		owner:     owner,
		repo:      repo,
		result: Result{
			Tool:        "coderabbit",
			BotUsername: BotUsername,
		},
		prDetails: make(map[int]*PRDetail),
	}
}

// CheckReview checks if CodeRabbit reviewed a PR based on the reviews list
func (a *Analyzer) CheckReview(reviews []*github.PullRequestReview) bool {
	for _, review := range reviews {
		if review.User != nil && review.User.GetLogin() == BotUsername {
			return true
		}
	}
	return false
}

// MarkPRReviewed marks a PR as reviewed by CodeRabbit
func (a *Analyzer) MarkPRReviewed(prNum int) {
	if _, exists := a.prDetails[prNum]; !exists {
		a.prDetails[prNum] = &PRDetail{PRNumber: prNum}
	}
}

// AnalyzeComments analyzes review comments for a PR
func (a *Analyzer) AnalyzeComments(ctx context.Context, prNum int) error {
	opts := &github.PullRequestListCommentsOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	detail := a.prDetails[prNum]

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
			detail.Comments++

			if comment.Reactions != nil {
				up := comment.Reactions.GetPlusOne()
				down := comment.Reactions.GetMinusOne()
				a.result.ThumbsUp += up
				a.result.ThumbsDown += down
				detail.ThumbsUp += up
				detail.ThumbsDown += down
			}
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

	detail := a.prDetails[prNum]

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
			detail.Comments++

			if comment.Reactions != nil {
				up := comment.Reactions.GetPlusOne()
				down := comment.Reactions.GetMinusOne()
				a.result.ThumbsUp += up
				a.result.ThumbsDown += down
				detail.ThumbsUp += up
				detail.ThumbsDown += down
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
	a.result.PRsReviewed = len(a.prDetails)

	var totalScore int
	var scoreCount int
	for _, detail := range a.prDetails {
		a.result.Details = append(a.result.Details, *detail)
		for _, ta := range detail.ThreadAnalyses {
			totalScore += ta.Score
			scoreCount++
		}
	}

	if scoreCount > 0 {
		avg := float64(totalScore) / float64(scoreCount)
		a.result.AvgLLMScore = &avg
	}

	return a.result
}

// HasReviews returns true if CodeRabbit reviewed any PRs
func (a *Analyzer) HasReviews() bool {
	return len(a.prDetails) > 0
}

// LogAnalysis logs the start of PR analysis
func (a *Analyzer) LogAnalysis(prNum int) {
	log.Printf("  CodeRabbit: analyzing PR #%d", prNum)
}

// AnalyzeThreadsWithLLM fetches review threads and analyzes them with the LLM
func (a *Analyzer) AnalyzeThreadsWithLLM(ctx context.Context, prNum int) error {
	if a.llmClient == nil {
		return nil // LLM not configured
	}

	// Fetch all review comments
	opts := &github.PullRequestListCommentsOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var allComments []*github.PullRequestComment
	for {
		comments, resp, err := a.client.PullRequests.ListComments(ctx, a.owner, a.repo, prNum, opts)
		if err != nil {
			return err
		}
		allComments = append(allComments, comments...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	// Build threads from comments
	threads := buildThreads(allComments, BotUsername)

	// Analyze each thread with user comments
	for _, thread := range threads {
		if len(thread) < 2 {
			continue // Skip threads without user replies
		}

		// Extract just the message bodies
		var messages []string
		for _, comment := range thread {
			messages = append(messages, comment.GetBody())
		}

		analysis, err := a.llmClient.AnalyzeThread(ctx, messages)
		if err != nil {
			log.Printf("Warning: LLM analysis failed for thread in PR #%d: %v", prNum, err)
			continue
		}

		detail := a.prDetails[prNum]
		detail.ThreadAnalyses = append(detail.ThreadAnalyses, ThreadAnalysis{
			Score:   analysis.Score,
			Summary: analysis.Summary,
		})
	}

	return nil
}

// buildThreads groups comments into threads starting from bot comments
func buildThreads(comments []*github.PullRequestComment, botUsername string) [][]*github.PullRequestComment {
	// Map comment ID to comment
	byID := make(map[int64]*github.PullRequestComment)
	for _, c := range comments {
		byID[c.GetID()] = c
	}

	// Map parent ID to replies
	replies := make(map[int64][]*github.PullRequestComment)
	for _, c := range comments {
		if c.InReplyTo != nil && c.GetInReplyTo() != 0 {
			replies[c.GetInReplyTo()] = append(replies[c.GetInReplyTo()], c)
		}
	}

	// Find bot comments that start threads
	var threads [][]*github.PullRequestComment
	for _, c := range comments {
		if c.User == nil || c.User.GetLogin() != botUsername {
			continue
		}
		// Only consider top-level bot comments (not replies)
		if c.InReplyTo != nil && c.GetInReplyTo() != 0 {
			continue
		}

		// Build thread: bot comment + all replies (recursively)
		thread := []*github.PullRequestComment{c}
		thread = collectReplies(c.GetID(), replies, thread)

		// Only include threads with at least one non-bot reply
		hasUserReply := false
		for i, tc := range thread {
			if i > 0 && tc.User != nil && tc.User.GetLogin() != botUsername {
				hasUserReply = true
				break
			}
		}
		if hasUserReply {
			threads = append(threads, thread)
		}
	}

	return threads
}

// collectReplies recursively collects all replies to a comment
func collectReplies(parentID int64, replies map[int64][]*github.PullRequestComment, thread []*github.PullRequestComment) []*github.PullRequestComment {
	for _, reply := range replies[parentID] {
		thread = append(thread, reply)
		thread = collectReplies(reply.GetID(), replies, thread)
	}
	return thread
}
