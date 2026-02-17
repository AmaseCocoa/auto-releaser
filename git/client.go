package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Client struct {
	token string
}

func NewClient() *Client {
	return &Client{
		token: os.Getenv("GITHUB_TOKEN"),
	}
}

func (c *Client) FetchUnshallow() error {
	cmd := exec.Command("git", "fetch", "--unshallow", "--tags")
	cmd.Env = append(os.Environ(), fmt.Sprintf("GIT_ASKPASS=echo"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "does not support") {
			cmd = exec.Command("git", "fetch", "--tags")
			output, err = cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("git fetch --tags failed: %w\n%s", err, output)
			}
			return nil
		}
		return fmt.Errorf("git fetch --unshallow failed: %w\n%s", err, output)
	}
	return nil
}

func (c *Client) GetPreviousTag(currentTag string) (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0", currentTag+"^")
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("git", "rev-list", "--max-parents=0", "HEAD")
		output, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to get first commit: %w", err)
		}
		return strings.TrimSpace(string(output)), nil
	}
	return strings.TrimSpace(string(output)), nil
}

func (c *Client) GetCommitsBetween(prevTag, currentTag string) ([]Commit, error) {
	cmd := exec.Command("git", "log", fmt.Sprintf("%s..%s", prevTag, currentTag), "--pretty=format:%H%x00%s")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	return ParseLog(string(output)), nil
}

func (c *Client) Checkout(branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout failed: %w\n%s", err, output)
	}
	return nil
}

func (c *Client) Pull(branch string) error {
	cmd := exec.Command("git", "pull", "origin", branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %w\n%s", err, output)
	}
	return nil
}

func (c *Client) Add(files ...string) error {
	args := append([]string{"add"}, files...)
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git add failed: %w\n%s", err, output)
	}
	return nil
}

func (c *Client) Commit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git commit failed: %w\n%s", err, output)
	}
	return nil
}

func (c *Client) Push(branch string) error {
	cmd := exec.Command("git", "push", "origin", branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push failed: %w\n%s", err, output)
	}
	return nil
}

func (c *Client) CreateBranch(branch string) error {
	cmd := exec.Command("git", "checkout", "-b", branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout -b failed: %w\n%s", err, output)
	}
	return nil
}

func (c *Client) CreateRelease(tag, notes string) error {
	cmd := exec.Command("gh", "release", "create", tag, "--notes", notes)
	cmd.Env = append(os.Environ(), fmt.Sprintf("GH_TOKEN=%s", c.token))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gh release create failed: %w\n%s", err, output)
	}
	return nil
}

func (c *Client) CreatePR(title, body, head, base string) (string, error) {
	var buf bytes.Buffer
	cmd := exec.Command("gh", "pr", "create", "--title", title, "--body", body, "--head", head, "--base", base)
	cmd.Env = append(os.Environ(), fmt.Sprintf("GH_TOKEN=%s", c.token))
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gh pr create failed: %w\n%s", err, buf.String())
	}
	return strings.TrimSpace(buf.String()), nil
}

func (c *Client) SetRemoteURL(repo string) error {
	url := fmt.Sprintf("https://x-access-token:%s@github.com/%s.git", c.token, repo)
	cmd := exec.Command("git", "remote", "set-url", "origin", url)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git remote set-url failed: %w\n%s", err, output)
	}
	return nil
}
