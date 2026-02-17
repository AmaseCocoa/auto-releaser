# Auto Releaser

Auto Releaser is a GitHub Action that automates the process of updating `CHANGELOG.md` via Pull Requests when a tag is pushed, and automatically creates a GitHub Release when the PR is merged.

## Features

* **Automated CHANGELOG Updates**: When a new tag is pushed, it analyzes commit messages since the last tag and creates a PR to update `CHANGELOG.md`.
* **Automated Release Creation**: Once the updated CHANGELOG PR is merged, it automatically creates a GitHub Release using the PR content as release notes.
* **Flexible Commit Parsing**: Supports Conventional Commits, Regex-based parsing, and a simple format.

## Usage

### Workflow Configuration

Add the following configuration to `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'
  pull_request:
    types:
      - closed

jobs:
  release:
    runs-on: ubuntu-latest
    permissions:
      contents: write
      pull-requests: write
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Auto Releaser
        uses: AmaseCocoa/auto-releaser@v1
        with:
          mode: ${{ github.event_name == 'push' && 'pr' || 'release' }}
          github_token: ${{ secrets.GITHUB_TOKEN }}
          main_branch: main

```

### Operational Flow

1. **Push a Tag**: Push a tag such as `v1.0.0`.
* The Action triggers (mode: `pr`) and generates a CHANGELOG from the commit history.
* A `chore/changelog-v1.0.0` branch is created, followed by a PR to update `CHANGELOG.md`.


2. **Review and Merge PR**: Review the generated PR and merge it.
3. **Create Release**: Upon merging (mode: `release`), the Action runs again to automatically create a GitHub Release.

## Inputs

| Input | Description | Default |
| --- | --- | --- |
| `mode` | Operation mode (`pr` or `release`). | **Required** |
| `adapter` | Commit parsing adapter (`conventional`, `regex`, `simple`). | `conventional` |
| `main_branch` | Target branch for creating Pull Requests. | `main` |
| `changelog_path` | File path to `CHANGELOG.md`. | `CHANGELOG.md` |
| `pattern` | Regex pattern used when the `regex` adapter is selected. | `""` |
| `github_token` | GitHub Token (requires `contents: write` and `pull-requests: write`). | `${{ github.token }}` |

## CHANGELOG Insertion Point

You can specify where to insert new release information by adding the following comment in your `CHANGELOG.md`:

```html
<!-- auto-releaser-start -->
```

If this comment is not found, new content will be prepended to the top of the file.

## Commit Parsing Adapters

### `conventional` (Default)

Parses the [Conventional Commits]() format. It categorizes the following types:

* `feat`: Features
* `fix`: Bug Fixes
* `docs`: Documentation
* `style`: Styles
* `refactor`: Code Refactoring
* `perf`: Performance Improvements
* `test`: Tests
* `chore`: Chores

### `simple`

Lists the first line of each commit message under a "Changes" section, excluding merge and revert commits.

### `regex`

Parses commits using a regular expression specified in the `pattern` input. The regex must contain two capture groups: the first for the category and the second for the content.

Example: `^\[(\w+)\]\s*(.*)$`