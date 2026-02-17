package adapters

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/AmaseCocoa/auto-releaser/changelog"
	"github.com/AmaseCocoa/auto-releaser/git"
)

type ConventionalCommitAdapter struct {
	sectionTitles map[string]string
}

func NewConventionalCommitAdapter() *ConventionalCommitAdapter {
	return &ConventionalCommitAdapter{
		sectionTitles: map[string]string{
			"feat":     "Features",
			"fix":      "Bug Fixes",
			"docs":     "Documentation",
			"style":    "Styles",
			"refactor": "Code Refactoring",
			"perf":     "Performance Improvements",
			"test":     "Tests",
			"chore":    "Chores",
		},
	}
}

func (a *ConventionalCommitAdapter) Parse(c git.Commit) (changelog.Change, bool) {
	message := strings.TrimSpace(c.Message)

	re := regexp.MustCompile(`^(\w+)(?:\([^)]+\))?:\s*(.+)$`)
	matches := re.FindStringSubmatch(message)

	if len(matches) < 3 {
		return changelog.Change{}, false
	}

	category := matches[1]
	content := matches[2]

	if _, exists := a.sectionTitles[category]; !exists {
		category = "other"
	}

	return changelog.Change{
		Category: category,
		Content:  content,
	}, true
}

func (a *ConventionalCommitAdapter) GetSectionTitle(category string) string {
	if title, exists := a.sectionTitles[category]; exists {
		return title
	}
	return strings.Title(category)
}

type RegexAdapter struct {
	pattern       *regexp.Regexp
	sectionTitles map[string]string
}

func NewRegexAdapter(pattern string) (*RegexAdapter, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	return &RegexAdapter{
		pattern:       re,
		sectionTitles: make(map[string]string),
	}, nil
}

func (a *RegexAdapter) Parse(c git.Commit) (changelog.Change, bool) {
	matches := a.pattern.FindStringSubmatch(c.Message)
	if len(matches) < 3 {
		return changelog.Change{}, false
	}

	category := matches[1]
	content := matches[2]

	return changelog.Change{
		Category: category,
		Content:  content,
	}, true
}

func (a *RegexAdapter) GetSectionTitle(category string) string {
	if title, exists := a.sectionTitles[category]; exists {
		return title
	}
	return strings.Title(category)
}

type SimpleAdapter struct{}

func NewSimpleAdapter() *SimpleAdapter {
	return &SimpleAdapter{}
}

func (a *SimpleAdapter) Parse(c git.Commit) (changelog.Change, bool) {
	message := strings.TrimSpace(c.Message)

	if strings.HasPrefix(message, "Merge ") || strings.HasPrefix(message, "Revert ") {
		return changelog.Change{}, false
	}

	lines := strings.Split(message, "\n")
	firstLine := strings.TrimSpace(lines[0])

	if firstLine == "" {
		return changelog.Change{}, false
	}

	return changelog.Change{
		Category: "changes",
		Content:  firstLine,
	}, true
}

func (a *SimpleAdapter) GetSectionTitle(category string) string {
	return "Changes"
}

func New(adapterType, pattern string) (changelog.Parser, error) {
	switch adapterType {
	case "conventional":
		return NewConventionalCommitAdapter(), nil
	case "regex":
		if pattern == "" {
			return nil, fmt.Errorf("pattern is required for regex adapter")
		}
		return NewRegexAdapter(pattern)
	case "simple":
		return NewSimpleAdapter(), nil
	default:
		return nil, fmt.Errorf("unknown adapter type: %s", adapterType)
	}
}
