package git

import (
	"strings"
)

type Commit struct {
	Hash    string
	Message string
}

func ParseLog(logOutput string) []Commit {
	if strings.TrimSpace(logOutput) == "" {
		return nil
	}

	entries := strings.Split(logOutput, "\n")
	commits := make([]Commit, 0, len(entries))

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		parts := strings.SplitN(entry, "\x00", 2)
		if len(parts) != 2 {
			continue
		}

		commits = append(commits, Commit{
			Hash:    parts[0],
			Message: parts[1],
		})
	}

	return commits
}
