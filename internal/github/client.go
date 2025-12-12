package github

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

const reviewRequestQuery = `
query {
  search(query: "is:pr is:open review-requested:@me", type: ISSUE, first: 100) {
    nodes {
      ... on PullRequest {
        number
        title
        url
        isDraft
        createdAt
        author { login }
        repository {
          name
          owner { login }
        }
        labels(first: 10) {
          nodes { name }
        }
        reviewRequests(first: 20) {
          nodes {
            requestedReviewer {
              ... on User { login }
            }
          }
        }
      }
    }
  }
}
`

const reviewedQuery = `
query {
  search(query: "is:pr is:open reviewed-by:@me -review-requested:@me -author:@me", type: ISSUE, first: 100) {
    nodes {
      ... on PullRequest {
        number
        title
        url
        isDraft
        createdAt
        updatedAt
        author { login }
        repository {
          name
          owner { login }
        }
        labels(first: 10) {
          nodes { name }
        }
        reviews(last: 100) {
          nodes {
            author { login }
            submittedAt
          }
        }
        commits(last: 100) {
          nodes {
            commit {
              committedDate
            }
          }
        }
        comments(last: 100) {
          nodes {
            createdAt
          }
        }
      }
    }
  }
}
`

const mentionsQuery = `
query {
  search(query: "is:pr is:open (mentions:@me OR commenter:@me) -author:@me", type: ISSUE, first: 100) {
    nodes {
      ... on PullRequest {
        number
        title
        url
        isDraft
        createdAt
        author { login }
        repository {
          name
          owner { login }
        }
        labels(first: 10) {
          nodes { name }
        }
      }
    }
  }
}
`

type graphQLResponse struct {
	Data struct {
		Search struct {
			Nodes []struct {
				Number    int    `json:"number"`
				Title     string `json:"title"`
				URL       string `json:"url"`
				IsDraft   bool   `json:"isDraft"`
				CreatedAt string `json:"createdAt"`
				Author    struct {
					Login string `json:"login"`
				} `json:"author"`
				Repository struct {
					Name  string `json:"name"`
					Owner struct {
						Login string `json:"login"`
					} `json:"owner"`
				} `json:"repository"`
				Labels struct {
					Nodes []struct {
						Name string `json:"name"`
					} `json:"nodes"`
				} `json:"labels"`
				ReviewRequests struct {
					Nodes []struct {
						RequestedReviewer struct {
							Login string `json:"login"`
						} `json:"requestedReviewer"`
					} `json:"nodes"`
				} `json:"reviewRequests"`
			} `json:"nodes"`
		} `json:"search"`
	} `json:"data"`
}

func FetchReviewRequests() ([]PR, error) {
	return fetchPRs(reviewRequestQuery, true)
}

func FetchMentions() ([]PR, error) {
	return fetchPRs(mentionsQuery, false)
}

func FetchAll() ([]PR, error) {
	pending, err := FetchReviewRequests()
	if err != nil {
		return nil, err
	}

	reviewed, err := FetchReviewed()
	if err != nil {
		return nil, err
	}

	seen := make(map[int]bool)
	var all []PR
	for _, pr := range pending {
		if !seen[pr.Number] {
			seen[pr.Number] = true
			all = append(all, pr)
		}
	}
	for _, pr := range reviewed {
		if !seen[pr.Number] {
			seen[pr.Number] = true
			all = append(all, pr)
		}
	}

	return all, nil
}

func FetchReviewed() ([]PR, error) {
	currentUser, err := getCurrentUser()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	cmd := exec.Command("gh", "api", "graphql", "-f", fmt.Sprintf("query=%s", reviewedQuery))
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to call gh api: %w", err)
	}

	var resp reviewedResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var prs []PR
	for _, node := range resp.Data.Search.Nodes {
		var lastReviewTime time.Time
		for _, review := range node.Reviews.Nodes {
			if review.Author.Login == currentUser {
				t, _ := time.Parse(time.RFC3339, review.SubmittedAt)
				if t.After(lastReviewTime) {
					lastReviewTime = t
				}
			}
		}

		newCommits := 0
		for _, commit := range node.Commits.Nodes {
			t, _ := time.Parse(time.RFC3339, commit.Commit.CommittedDate)
			if t.After(lastReviewTime) {
				newCommits++
			}
		}

		newComments := 0
		for _, comment := range node.Comments.Nodes {
			t, _ := time.Parse(time.RFC3339, comment.CreatedAt)
			if t.After(lastReviewTime) {
				newComments++
			}
		}

		activity := ""
		if newCommits > 0 || newComments > 0 {
			parts := []string{}
			if newCommits > 0 {
				parts = append(parts, fmt.Sprintf("%d commits", newCommits))
			}
			if newComments > 0 {
				parts = append(parts, fmt.Sprintf("%d comments", newComments))
			}
			activity = fmt.Sprintf("%s", joinStrings(parts, ", "))
		}

		labels := make([]string, len(node.Labels.Nodes))
		for i, l := range node.Labels.Nodes {
			labels[i] = l.Name
		}

		createdAt, _ := time.Parse(time.RFC3339, node.CreatedAt)

		prs = append(prs, PR{
			Number:    node.Number,
			Title:     node.Title,
			URL:       node.URL,
			Author:    node.Author.Login,
			Repo:      node.Repository.Name,
			Owner:     node.Repository.Owner.Login,
			IsDraft:   node.IsDraft,
			Labels:    labels,
			CreatedAt: createdAt,
			Activity:  activity,
		})
	}

	return prs, nil
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

type reviewedResponse struct {
	Data struct {
		Search struct {
			Nodes []struct {
				Number    int    `json:"number"`
				Title     string `json:"title"`
				URL       string `json:"url"`
				IsDraft   bool   `json:"isDraft"`
				CreatedAt string `json:"createdAt"`
				Author    struct {
					Login string `json:"login"`
				} `json:"author"`
				Repository struct {
					Name  string `json:"name"`
					Owner struct {
						Login string `json:"login"`
					} `json:"owner"`
				} `json:"repository"`
				Labels struct {
					Nodes []struct {
						Name string `json:"name"`
					} `json:"nodes"`
				} `json:"labels"`
				Reviews struct {
					Nodes []struct {
						Author struct {
							Login string `json:"login"`
						} `json:"author"`
						SubmittedAt string `json:"submittedAt"`
					} `json:"nodes"`
				} `json:"reviews"`
				Commits struct {
					Nodes []struct {
						Commit struct {
							CommittedDate string `json:"committedDate"`
						} `json:"commit"`
					} `json:"nodes"`
				} `json:"commits"`
				Comments struct {
					Nodes []struct {
						CreatedAt string `json:"createdAt"`
					} `json:"nodes"`
				} `json:"comments"`
			} `json:"nodes"`
		} `json:"search"`
	} `json:"data"`
}

func getCurrentUser() (string, error) {
	cmd := exec.Command("gh", "api", "user", "--jq", ".login")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output[:len(output)-1]), nil
}

func fetchPRs(query string, filterDirectReviewer bool) ([]PR, error) {
	var currentUser string
	if filterDirectReviewer {
		var err error
		currentUser, err = getCurrentUser()
		if err != nil {
			return nil, fmt.Errorf("failed to get current user: %w", err)
		}
	}

	cmd := exec.Command("gh", "api", "graphql", "-f", fmt.Sprintf("query=%s", query))
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to call gh api: %w", err)
	}

	var resp graphQLResponse
	if err := json.Unmarshal(output, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var prs []PR
	for _, node := range resp.Data.Search.Nodes {
		if filterDirectReviewer {
			isDirectReviewer := false
			for _, rr := range node.ReviewRequests.Nodes {
				if rr.RequestedReviewer.Login == currentUser {
					isDirectReviewer = true
					break
				}
			}
			if !isDirectReviewer {
				continue
			}
		}

		labels := make([]string, len(node.Labels.Nodes))
		for i, l := range node.Labels.Nodes {
			labels[i] = l.Name
		}

		createdAt, _ := time.Parse(time.RFC3339, node.CreatedAt)

		prs = append(prs, PR{
			Number:    node.Number,
			Title:     node.Title,
			URL:       node.URL,
			Author:    node.Author.Login,
			Repo:      node.Repository.Name,
			Owner:     node.Repository.Owner.Login,
			IsDraft:   node.IsDraft,
			Labels:    labels,
			CreatedAt: createdAt,
		})
	}

	return prs, nil
}
