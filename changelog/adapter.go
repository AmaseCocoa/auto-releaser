package changelog

import (
	"github.com/AmaseCocoa/auto-releaser/git"
)

type Change struct {
	Category string
	Content  string
}

type Parser interface {
	Parse(c git.Commit) (change Change, ok bool)
	GetSectionTitle(category string) string
}
