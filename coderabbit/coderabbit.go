package coderabbit

import (
	"context"
	"strings"

	"ai-review-analyzer/ghgql"
	"ai-review-analyzer/llm"
)

const BotUsername = "coderabbitai[bot]"

// botLogin returns the username without [bot] suffix (as returned by GraphQL)
func botLogin() string {
	return strings.TrimSuffix(BotUsername, "[bot]")
}

// ThreadAnalysis represents an analyzed discussion thread
type ThreadAnalysis struct {
	DeveloperFeedbackScore int    `json:"developer_feedback_score"`
	Summary                string `json:"summary"`
}

// PRMetrics represents per-PR metrics matching CodeRabbit API structure
type PRMetrics struct {
	AISuggestionThreads       int      `json:"ai_suggestion_threads"`
	AcceptedSuggestions       int      `json:"accepted_suggestions"`
	AutoResolvedByUser        int      `json:"auto_resolved_by_user"`
	RejectedComments          int      `json:"rejected_comments"`
	CriticalBugsFixed         int      `json:"critical_bugs_fixed"`
	AvgDeveloperFeedbackScore *int `json:"avg_developer_feedback_score,omitempty"`
}

// PRDetail represents per-PR data matching CodeRabbit API structure
type PRDetail struct {
	URL          string           `json:"url"`
	Author         string           `json:"author"`
	Metrics        PRMetrics        `json:"metrics"`
	DeveloperFeedbackAnalyses []ThreadAnalysis `json:"developer_feedback_analyses,omitempty"`
}

// Result represents the analysis result for CodeRabbit
type Result struct {
	Tool                string     `json:"ai_review_tool"`
	BotUsername         string     `json:"bot_username"`
	TotalPRs            int        `json:"total_prs"`
	AISuggestionThreads int        `json:"ai_suggestion_threads"`
	AcceptedSuggestions int        `json:"accepted_suggestions"`
	AutoResolvedByUser  int        `json:"auto_resolved_by_user"`
	RejectedComments    int        `json:"rejected_comments"`
	CriticalBugsFixed   int        `json:"critical_bugs_fixed"`
	ThumbsUp            int        `json:"thumbs_up"`
	ThumbsDown          int        `json:"thumbs_down"`
	AvgDeveloperFeedbackScore *int `json:"avg_developer_feedback_score,omitempty"`
	PRs                 []PRDetail `json:"prs,omitempty"`
}

// Analyzer handles CodeRabbit-specific analysis
type Analyzer struct {
	llmClient *llm.Client
	result    Result
}

// NewAnalyzer creates a new CodeRabbit analyzer
func NewAnalyzer(llmClient *llm.Client) *Analyzer {
	return &Analyzer{
		llmClient: llmClient,
		result: Result{
			Tool:        "coderabbit",
			BotUsername: BotUsername,
		},
	}
}

// AnalyzePR analyzes a single PR for CodeRabbit metrics
func (a *Analyzer) AnalyzePR(ctx context.Context, pr ghgql.PRData) {
	detail := PRDetail{
		URL:  pr.URL,
		Author: pr.Author,
	}

	// Calculate thread stats
	stats := ghgql.CalculateThreadStats(pr.Threads, BotUsername)

	// Count comments and reactions from bot-initiated threads
	for _, thread := range pr.Threads {
		if len(thread.Comments) == 0 || thread.Comments[0].AuthorLogin != botLogin() {
			continue
		}

		// First comment is from the bot
		detail.Metrics.AISuggestionThreads++
		a.result.AISuggestionThreads++

		// Count reactions on bot comments
		for _, comment := range thread.Comments {
			if comment.AuthorLogin == botLogin() {
				a.result.ThumbsUp += comment.ThumbsUp
				a.result.ThumbsDown += comment.ThumbsDown
			}
		}
	}

	// Auto-resolved = meaningful suggestions (will subtract accepted later)
	detail.Metrics.AutoResolvedByUser = stats.MeaningfulCount
	a.result.AutoResolvedByUser += stats.MeaningfulCount

	// Critical bugs fixed
	detail.Metrics.CriticalBugsFixed = stats.CriticalFixed
	a.result.CriticalBugsFixed += stats.CriticalFixed

	// Rejected comments (unresolved on merged PRs)
	if pr.IsMerged {
		detail.Metrics.RejectedComments = stats.UnresolvedThreads
		a.result.RejectedComments += stats.UnresolvedThreads
	}

	// Check commits for "Apply Suggestion" pattern
	for _, msg := range pr.CommitMsgs {
		msgLower := strings.ToLower(msg)
		if strings.Contains(msgLower, "apply suggestion") &&
			strings.Contains(msgLower, "coderabbit") {
			detail.Metrics.AcceptedSuggestions++
			a.result.AcceptedSuggestions++

			// Subtract from auto_resolved
			if detail.Metrics.AutoResolvedByUser > 0 {
				detail.Metrics.AutoResolvedByUser--
				a.result.AutoResolvedByUser--
			}
		}
	}

	// LLM analysis for threads with user replies
	if a.llmClient != nil {
		for _, thread := range pr.Threads {
			if len(thread.Comments) < 2 {
				continue // Skip threads without user replies
			}
			if thread.Comments[0].AuthorLogin != botLogin() {
				continue // Skip threads not started by bot
			}

			// Check if there's at least one non-bot reply
			hasUserReply := false
			for i, c := range thread.Comments {
				if i > 0 && c.AuthorLogin != botLogin() {
					hasUserReply = true
					break
				}
			}
			if !hasUserReply {
				continue
			}

			// Extract message bodies
			var messages []string
			for _, c := range thread.Comments {
				messages = append(messages, c.Body)
			}

			analysis, err := a.llmClient.AnalyzeThread(ctx, messages)
			if err != nil {
				continue // Skip on error
			}

			detail.DeveloperFeedbackAnalyses = append(detail.DeveloperFeedbackAnalyses, ThreadAnalysis{
				DeveloperFeedbackScore: analysis.DeveloperFeedbackScore,
				Summary:                analysis.Summary,
			})
		}
	}

	// Calculate average developer feedback score for this PR
	if len(detail.DeveloperFeedbackAnalyses) > 0 {
		var total int
		for _, ta := range detail.DeveloperFeedbackAnalyses {
			total += ta.DeveloperFeedbackScore
		}
		avg := total / len(detail.DeveloperFeedbackAnalyses)
		detail.Metrics.AvgDeveloperFeedbackScore = &avg
	}

	a.result.PRs = append(a.result.PRs, detail)
}

// GetResult returns the final analysis result
func (a *Analyzer) GetResult() Result {
	a.result.TotalPRs = len(a.result.PRs)

	var totalScore int
	var scoreCount int
	for _, detail := range a.result.PRs {
		for _, ta := range detail.DeveloperFeedbackAnalyses {
			totalScore += ta.DeveloperFeedbackScore
			scoreCount++
		}
	}

	if scoreCount > 0 {
		avg := totalScore / scoreCount
		a.result.AvgDeveloperFeedbackScore = &avg
	}

	return a.result
}

// HasReviews returns true if CodeRabbit reviewed any PRs
func (a *Analyzer) HasReviews() bool {
	return len(a.result.PRs) > 0
}
