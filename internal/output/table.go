package output

import (
	"os"
	"strconv"
	"time"

	"github.com/amiraminb/plantir/internal/github"
	"github.com/olekukonko/tablewriter"
)

func age(t time.Time) string {
	d := time.Since(t)

	if d.Hours() >= 24 {
		days := int(d.Hours() / 24)
		return strconv.Itoa(days) + "d"
	}
	if d.Hours() >= 1 {
		return strconv.Itoa(int(d.Hours())) + "h"
	}
	return strconv.Itoa(int(d.Minutes())) + "m"
}

func Table(prs []github.PR) {
	hasActivity := false
	for _, pr := range prs {
		if pr.Activity != "" {
			hasActivity = true
			break
		}
	}

	table := tablewriter.NewTable(os.Stdout)
	if hasActivity {
		table.Header("Repo", "PR#", "Title", "Author", "Age", "Type", "Activity")
	} else {
		table.Header("Repo", "PR#", "Title", "Author", "Age", "Type")
	}

	for _, pr := range prs {
		title := pr.Title
		if len(title) > 45 {
			title = title[:42] + "..."
		}

		if hasActivity {
			activity := pr.Activity
			if activity == "" {
				activity = "-"
			}
			table.Append([]string{
				pr.Repo,
				"#" + strconv.Itoa(pr.Number),
				title,
				pr.Author,
				age(pr.CreatedAt),
				pr.Type(),
				activity,
			})
		} else {
			table.Append([]string{
				pr.Repo,
				"#" + strconv.Itoa(pr.Number),
				title,
				pr.Author,
				age(pr.CreatedAt),
				pr.Type(),
			})
		}
	}

	table.Render()
}
