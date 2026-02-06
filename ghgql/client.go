package ghgql

import (
	"context"
	"strings"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// Client wraps the GitHub GraphQL client
type Client struct {
	client *githubv4.Client
}

// NewClient creates a new GraphQL client with the given token
func NewClient(token string) *Client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	return &Client{
		client: githubv4.NewClient(httpClient),
	}
}

// Comment represents a review comment
type Comment struct {
	AuthorLogin string
	Body        string
	ThumbsUp    int
	ThumbsDown  int
}

// ReviewThread represents a review thread with all its data
type ReviewThread struct {
	IsResolved bool
	IsOutdated bool
	ResolvedBy string // login of user who resolved
	Comments   []Comment
	IsCritical bool // first comment contains 🔴 emoji
}

// PRData represents combined REST and GraphQL data for a PR
type PRData struct {
	Number     int
	URL        string
	Author     string
	IsMerged   bool
	CommitMsgs []string
	Threads    []ReviewThread
}

// reviewThreadsQuery is the GraphQL query structure for fetching threads of a single PR
type reviewThreadsQuery struct {
	Repository struct {
		PullRequest struct {
			ReviewThreads struct {
				Nodes []struct {
					IsResolved bool
					IsOutdated bool
					ResolvedBy struct {
						Login string
					}
					Comments struct {
						Nodes []struct {
							Author struct {
								Login string
							}
							Body           string
							ReactionGroups []struct {
								Content  string
								Reactors struct {
									TotalCount int
								}
							}
						}
						PageInfo struct {
							HasNextPage bool
							EndCursor   string
						}
					} `graphql:"comments(first: 100, after: $commentsCursor)"`
				}
				PageInfo struct {
					HasNextPage bool
					EndCursor   string
				}
			} `graphql:"reviewThreads(first: 100, after: $threadsCursor)"`
		} `graphql:"pullRequest(number: $prNumber)"`
	} `graphql:"repository(owner: $owner, name: $repo)"`
}

// FetchReviewThreads fetches all review threads for a single PR with pagination
func (c *Client) FetchReviewThreads(ctx context.Context, owner, repo string, prNumber int) ([]ReviewThread, error) {
	var allThreads []ReviewThread
	var threadsCursor *githubv4.String

	for {
		var query reviewThreadsQuery
		variables := map[string]interface{}{
			"owner":          githubv4.String(owner),
			"repo":           githubv4.String(repo),
			"prNumber":       githubv4.Int(prNumber),
			"threadsCursor":  threadsCursor,
			"commentsCursor": (*githubv4.String)(nil),
		}

		err := c.client.Query(ctx, &query, variables)
		if err != nil {
			return nil, err
		}

		for _, node := range query.Repository.PullRequest.ReviewThreads.Nodes {
			rt := ReviewThread{
				IsResolved: node.IsResolved,
				IsOutdated: node.IsOutdated,
				ResolvedBy: node.ResolvedBy.Login,
			}

			// Collect comments from first page
			for i, comment := range node.Comments.Nodes {
				c := Comment{
					AuthorLogin: comment.Author.Login,
					Body:        comment.Body,
				}

				// Extract reactions
				for _, rg := range comment.ReactionGroups {
					switch rg.Content {
					case "THUMBS_UP":
						c.ThumbsUp = rg.Reactors.TotalCount
					case "THUMBS_DOWN":
						c.ThumbsDown = rg.Reactors.TotalCount
					}
				}

				rt.Comments = append(rt.Comments, c)

				// Check first comment for critical emoji
				if i == 0 && strings.Contains(comment.Body, "🔴") {
					rt.IsCritical = true
				}
			}

			// If there are more comments, fetch them (rare case)
			if node.Comments.PageInfo.HasNextPage {
				moreComments, err := c.fetchRemainingComments(ctx, owner, repo, prNumber, rt, node.Comments.PageInfo.EndCursor)
				if err != nil {
					return nil, err
				}
				rt.Comments = append(rt.Comments, moreComments...)
			}

			allThreads = append(allThreads, rt)
		}

		if !query.Repository.PullRequest.ReviewThreads.PageInfo.HasNextPage {
			break
		}
		threadsCursor = (*githubv4.String)(&query.Repository.PullRequest.ReviewThreads.PageInfo.EndCursor)
	}

	return allThreads, nil
}

// commentsQuery fetches additional comments for a thread
type commentsQuery struct {
	Repository struct {
		PullRequest struct {
			ReviewThreads struct {
				Nodes []struct {
					Comments struct {
						Nodes []struct {
							Author struct {
								Login string
							}
							Body           string
							ReactionGroups []struct {
								Content  string
								Reactors struct {
									TotalCount int
								}
							}
						}
						PageInfo struct {
							HasNextPage bool
							EndCursor   string
						}
					} `graphql:"comments(first: 100, after: $cursor)"`
				}
			} `graphql:"reviewThreads(first: 1)"`
		} `graphql:"pullRequest(number: $prNumber)"`
	} `graphql:"repository(owner: $owner, name: $repo)"`
}

// fetchRemainingComments fetches additional comments beyond the first page
func (c *Client) fetchRemainingComments(ctx context.Context, owner, repo string, prNumber int, rt ReviewThread, cursor string) ([]Comment, error) {
	// For simplicity in PoC, we don't implement deep comment pagination
	// Threads with 100+ comments are extremely rare
	return nil, nil
}

// ThreadStats contains aggregated thread statistics for a bot
type ThreadStats struct {
	TotalThreads      int
	ResolvedThreads   int
	MeaningfulCount   int
	FalsePositives    int
	BotResolvedCount  int
	CriticalFixed     int
	UnresolvedThreads int
}

// CalculateThreadStats calculates statistics for threads started by a specific bot
func CalculateThreadStats(threads []ReviewThread, botUsername string) ThreadStats {
	// GraphQL returns username without [bot] suffix
	botLogin := strings.TrimSuffix(botUsername, "[bot]")

	var stats ThreadStats

	for _, t := range threads {
		// Only count threads started by this bot
		if len(t.Comments) == 0 || t.Comments[0].AuthorLogin != botLogin {
			continue
		}

		stats.TotalThreads++

		if t.IsResolved {
			stats.ResolvedThreads++

			if t.IsOutdated {
				stats.FalsePositives++
			} else {
				stats.MeaningfulCount++
				if t.IsCritical {
					stats.CriticalFixed++
				}
			}

			if t.ResolvedBy == botLogin {
				stats.BotResolvedCount++
			}
		} else {
			stats.UnresolvedThreads++
		}
	}

	return stats
}
