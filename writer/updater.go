package writer

import (
	"fmt"
	"os"
	"strings"

	"github.com/AmaseCocoa/auto-releaser/git"
)

type Updater struct {
	gitClient     *git.Client
	mainBranch    string
	changelogPath string
	token         string
}

func NewUpdater(gitClient *git.Client, mainBranch, changelogPath, token string) *Updater {
	return &Updater{
		gitClient:     gitClient,
		mainBranch:    mainBranch,
		changelogPath: changelogPath,
		token:         token,
	}
}

func (u *Updater) Update(content, tag string) error {
	if err := u.gitClient.Checkout(u.mainBranch); err != nil {
		return err
	}

	if err := u.gitClient.Pull(u.mainBranch); err != nil {
		return err
	}

	if err := u.insertChangelog(content); err != nil {
		return fmt.Errorf("failed to update changelog file: %w", err)
	}

	if err := u.gitClient.Add(u.changelogPath); err != nil {
		return err
	}

	commitMsg := fmt.Sprintf("docs: release %s [skip ci]", tag)
	if err := u.gitClient.Commit(commitMsg); err != nil {
		return err
	}

	if err := u.gitClient.Push(u.mainBranch); err != nil {
		return u.createPR(content, tag)
	}

	return nil
}

func (u *Updater) insertChangelog(content string) error {
	existingContent := ""
	if data, err := os.ReadFile(u.changelogPath); err == nil {
		existingContent = string(data)
	}

	insertPosition := u.findInsertPosition(existingContent)

	var result strings.Builder
	if insertPosition > 0 {
		result.WriteString(existingContent[:insertPosition])
		result.WriteString("\n")
		result.WriteString(content)
		result.WriteString(existingContent[insertPosition:])
	} else {
		result.WriteString("<!-- auto-releaser-start -->\n\n")
		result.WriteString(content)
		result.WriteString("\n\n")
		if existingContent != "" {
			result.WriteString(existingContent)
		}
	}

	return os.WriteFile(u.changelogPath, []byte(result.String()), 0644)
}

func (u *Updater) findInsertPosition(content string) int {
	anchor := "<!-- auto-releaser-start -->"
	if idx := strings.Index(content, anchor); idx != -1 {
		return idx + len(anchor)
	}
	return -1
}

func (u *Updater) CreatePRForChangelog(content, tag string) error {
	branchName := fmt.Sprintf("chore/changelog-%s", tag)

	if err := u.gitClient.Checkout(u.mainBranch); err != nil {
		return err
	}

	if err := u.gitClient.Pull(u.mainBranch); err != nil {
		return err
	}

	if err := u.gitClient.CreateBranch(branchName); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	if err := u.insertChangelog(content); err != nil {
		return fmt.Errorf("failed to update changelog file: %w", err)
	}

	if err := u.gitClient.Add(u.changelogPath); err != nil {
		return err
	}

	commitMsg := fmt.Sprintf("docs: release %s [skip ci]", tag)
	if err := u.gitClient.Commit(commitMsg); err != nil {
		return err
	}

	if err := u.gitClient.Push(branchName); err != nil {
		return fmt.Errorf("failed to push branch: %w", err)
	}

	title := fmt.Sprintf("docs: update changelog for %s", tag)
	body := fmt.Sprintf("Automated changelog update for release %s\n\n## Changes:\n\n%s", tag, content)

	prURL, err := u.gitClient.CreatePR(title, body, branchName, u.mainBranch)
	if err != nil {
		return fmt.Errorf("failed to create PR: %w", err)
	}

	fmt.Printf("Created PR for changelog update: %s\n", prURL)
	return nil
}

func (u *Updater) createPR(content, tag string) error {
	branchName := fmt.Sprintf("chore/changelog-%s", tag)

	if err := u.gitClient.CreateBranch(branchName); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	if err := u.gitClient.Push(branchName); err != nil {
		return fmt.Errorf("failed to push branch: %w", err)
	}

	title := fmt.Sprintf("docs: update changelog for %s", tag)
	body := fmt.Sprintf("Automated changelog update for release %s\n\n## Changes:\n\n%s", tag, content)

	prURL, err := u.gitClient.CreatePR(title, body, branchName, u.mainBranch)
	if err != nil {
		return fmt.Errorf("failed to create PR: %w", err)
	}

	fmt.Printf("Created PR for changelog update: %s\n", prURL)
	return nil
}
