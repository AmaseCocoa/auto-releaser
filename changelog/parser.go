package changelog

import (
	"github.com/AmaseCocoa/auto-releaser/git"
)

func ParseCommits(commits []git.Commit, parser Parser) []Change {
	changes := make([]Change, 0)

	for _, commit := range commits {
		if change, ok := parser.Parse(commit); ok {
			changes = append(changes, change)
		}
	}

	return changes
}
