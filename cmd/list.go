package cmd

import (
	"fmt"
	"sort"

	"github.com/amiraminb/plantir/internal/filter"
	"github.com/amiraminb/plantir/internal/github"
	"github.com/amiraminb/plantir/internal/output"
	"github.com/spf13/cobra"
)

var (
	repoFlag     string
	typeFlag     string
	jsonFlag     bool
	limitFlag    int
	staleFlag    int
	reviewedFlag bool
	pendingFlag  bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List PRs related to your reviews",
	Long:  `Fetches all open pull requests where you are requested as reviewer or have reviewed.`,
	Run: func(cmd *cobra.Command, args []string) {
		var prs []github.PR
		var err error
		var emptyMsg, headerMsg string

		if reviewedFlag && pendingFlag {
			fmt.Println("Error: cannot use both --reviewed and --pending")
			return
		}

		if reviewedFlag {
			prs, err = github.FetchReviewed()
			emptyMsg = "✨ No PRs you've reviewed!"
			headerMsg = "👀 PRs you've reviewed..."
		} else if pendingFlag {
			prs, err = github.FetchReviewRequests()
			emptyMsg = "✨ No PRs waiting for your review!"
			headerMsg = "🔍 PRs waiting for your review..."
		} else {
			prs, err = github.FetchAll()
			emptyMsg = "✨ No PRs related to you!"
			headerMsg = "🔮 All PRs (pending + reviewed)..."
		}

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		prs = filter.Apply(prs, filter.Options{
			Repo:      repoFlag,
			Type:      typeFlag,
			StaleDays: staleFlag,
		})

		sort.Slice(prs, func(i, j int) bool {
			return prs[i].CreatedAt.After(prs[j].CreatedAt)
		})

		totalCount := len(prs)

		if totalCount == 0 {
			if jsonFlag {
				fmt.Println("[]")
				return
			}
			fmt.Println(emptyMsg)
			return
		}

		hasMore := false
		if limitFlag > 0 && len(prs) > limitFlag {
			prs = prs[:limitFlag]
			hasMore = true
		}

		if jsonFlag {
			output.JSON(prs)
		} else {
			fmt.Println(headerMsg)
			if hasMore {
				fmt.Printf("\nShowing %d of %d PRs (use --limit to see more):\n\n", limitFlag, totalCount)
			} else {
				fmt.Printf("\nFound %d PRs:\n\n", totalCount)
			}
			output.Table(prs)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringVarP(&repoFlag, "repo", "r", "", "Filter by repository name")
	listCmd.Flags().StringVarP(&typeFlag, "type", "t", "", "Filter by PR type (feature, dependabot, bot)")
	listCmd.Flags().BoolVar(&jsonFlag, "json", false, "Output as JSON")
	listCmd.Flags().IntVarP(&limitFlag, "limit", "n", 20, "Maximum number of PRs to show (0 for unlimited)")
	listCmd.Flags().IntVarP(&staleFlag, "stale", "s", 0, "Only show PRs older than N days")
	listCmd.Flags().BoolVarP(&pendingFlag, "pending", "p", false, "Show only PRs waiting for your review")
	listCmd.Flags().BoolVar(&reviewedFlag, "reviewed", false, "Show only PRs you've already reviewed")
}
