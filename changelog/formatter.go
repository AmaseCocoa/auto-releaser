package changelog

import (
	"fmt"
	"sort"
	"strings"
)

func Format(tag string, changes []Change, parser Parser) string {
	if len(changes) == 0 {
		return fmt.Sprintf("## %s\n\nNo notable changes.\n", tag)
	}

	sections := make(map[string][]string)

	for _, change := range changes {
		sections[change.Category] = append(sections[change.Category], change.Content)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("## %s\n\n", tag))

	categories := make([]string, 0, len(sections))
	for cat := range sections {
		categories = append(categories, cat)
	}
	sort.Strings(categories)

	for _, cat := range categories {
		title := parser.GetSectionTitle(cat)
		result.WriteString(fmt.Sprintf("### %s\n\n", title))

		items := sections[cat]
		sort.Strings(items)

		for _, item := range items {
			result.WriteString(fmt.Sprintf("- %s\n", item))
		}
		result.WriteString("\n")
	}

	return strings.TrimSpace(result.String())
}
