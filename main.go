package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/AmaseCocoa/auto-releaser/changelog"
	"github.com/AmaseCocoa/auto-releaser/changelog/adapters"
	"github.com/AmaseCocoa/auto-releaser/git"
	"github.com/AmaseCocoa/auto-releaser/writer"
)

type Config struct {
	Mode          string
	Adapter       string
	MainBranch    string
	ChangelogPath string
	Pattern       string
	Token         string
	CurrentTag    string
	Repository    string
	EventName     string
	EventAction   string
	IsMerged      bool
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	switch cfg.Mode {
	case "pr":
		return runPRCreation(cfg)
	case "release":
		return runReleaseCreation(cfg)
	default:
		return fmt.Errorf("unknown mode: %s", cfg.Mode)
	}
}

func runPRCreation(cfg *Config) error {
	gitClient := git.NewClient()

	if err := gitClient.SetRemoteURL(cfg.Repository); err != nil {
		return fmt.Errorf("failed to set remote URL: %w", err)
	}

	if err := gitClient.FetchUnshallow(); err != nil {
		return fmt.Errorf("failed to fetch tags: %w", err)
	}

	prevTag, err := gitClient.GetPreviousTag(cfg.CurrentTag)
	if err != nil {
		return fmt.Errorf("failed to get previous tag: %w", err)
	}

	commits, err := gitClient.GetCommitsBetween(prevTag, cfg.CurrentTag)
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}

	parser, err := adapters.New(cfg.Adapter, cfg.Pattern)
	if err != nil {
		return fmt.Errorf("failed to create parser: %w", err)
	}

	changes := changelog.ParseCommits(commits, parser)
	formatted := changelog.Format(cfg.CurrentTag, changes, parser)

	updater := writer.NewUpdater(gitClient, cfg.MainBranch, cfg.ChangelogPath, cfg.Token)
	if err := updater.CreatePRForChangelog(formatted, cfg.CurrentTag); err != nil {
		return fmt.Errorf("failed to create PR: %w", err)
	}

	fmt.Println("Successfully created PR for changelog update")
	return nil
}

func runReleaseCreation(cfg *Config) error {
	gitClient := git.NewClient()

	if err := gitClient.SetRemoteURL(cfg.Repository); err != nil {
		return fmt.Errorf("failed to set remote URL: %w", err)
	}

	if err := gitClient.FetchUnshallow(); err != nil {
		return fmt.Errorf("failed to fetch tags: %w", err)
	}

	prevTag, err := gitClient.GetPreviousTag(cfg.CurrentTag)
	if err != nil {
		return fmt.Errorf("failed to get previous tag: %w", err)
	}

	commits, err := gitClient.GetCommitsBetween(prevTag, cfg.CurrentTag)
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}

	parser, err := adapters.New(cfg.Adapter, cfg.Pattern)
	if err != nil {
		return fmt.Errorf("failed to create parser: %w", err)
	}

	changes := changelog.ParseCommits(commits, parser)
	formatted := changelog.Format(cfg.CurrentTag, changes, parser)

	if err := gitClient.CreateRelease(cfg.CurrentTag, formatted); err != nil {
		return fmt.Errorf("failed to create release: %w", err)
	}

	fmt.Println("Successfully created release")
	return nil
}

func loadConfig() (*Config, error) {
	eventName := os.Getenv("GITHUB_EVENT_NAME")
	eventAction := os.Getenv("GITHUB_EVENT_ACTION")
	isMerged := os.Getenv("GITHUB_EVENT_PULL_REQUEST_MERGED") == "true"

	mode := determineMode()

	cfg := &Config{
		Mode:          mode,
		Adapter:       getEnv("INPUT_ADAPTER", "conventional"),
		MainBranch:    getEnv("INPUT_MAIN_BRANCH", "main"),
		ChangelogPath: getEnv("INPUT_CHANGELOG_PATH", "CHANGELOG.md"),
		Pattern:       os.Getenv("INPUT_PATTERN"),
		Token:         os.Getenv("GITHUB_TOKEN"),
		CurrentTag:    os.Getenv("GITHUB_REF"),
		Repository:    os.Getenv("GITHUB_REPOSITORY"),
		EventName:     eventName,
		EventAction:   eventAction,
		IsMerged:      isMerged,
	}

	if cfg.Token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN is required")
	}

	cfg.CurrentTag = strings.TrimPrefix(cfg.CurrentTag, "refs/tags/")
	if cfg.CurrentTag == "" && mode == "pr" {
		return nil, fmt.Errorf("GITHUB_REF is required for PR creation")
	}

	if cfg.Repository == "" {
		return nil, fmt.Errorf("GITHUB_REPOSITORY is required")
	}

	return cfg, nil
}

func determineMode() string {
	return os.Getenv("INPUT_MODE")
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
