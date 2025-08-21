# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GitLab Claude Reviewer is a Go-based AI-powered code review tool that integrates Claude AI with GitLab merge requests. It fetches merge request diffs from GitLab, sends them to Claude for analysis, and posts intelligent code review comments back to the merge request.

## Architecture

### Core Components

- **cmd/gitlab-claude-reviewer/main.go**: CLI entry point that handles command-line arguments and orchestrates the review process
- **internal/reviewer/reviewer.go**: Main business logic that coordinates GitLab and Claude API interactions
- **internal/gitlab/client.go**: GitLab API client for fetching merge requests, diffs, and posting comments
- **internal/claude/client.go**: Claude API client for sending code review requests

### Data Flow

1. CLI parses arguments and creates GitLab/Claude clients
2. Reviewer fetches merge request details and diff from GitLab
3. Diff is formatted and sent to Claude with review instructions
4. Claude's response is formatted and posted back to GitLab as a comment

## Development Commands

### Building
```bash
# Build the main binary
go build -o gitlab-claude-reviewer ./cmd/gitlab-claude-reviewer

# Build and test everything
./test.sh
```

### Testing
```bash
# Run the test script (includes build, help test, Docker build)
./test.sh

# Manual testing with real GitLab project
export GITLAB_TOKEN="your_token"
export CLAUDE_API_KEY="your_key"
./gitlab-claude-reviewer --project-id="123" --mr-id="45" --verbose
```

### Docker
```bash
# Build Docker image
docker build -t gitlab-claude-reviewer .

# Test Docker container
docker run --rm gitlab-claude-reviewer --help
```

## Key Configuration

### Environment Variables
- `GITLAB_TOKEN`: GitLab personal access token with `api` scope
- `CLAUDE_API_KEY`: Claude API key from Anthropic

### CLI Flags
- `--gitlab-token`: GitLab token (overrides env var)
- `--claude-token`: Claude API key (overrides env var)  
- `--gitlab-url`: GitLab instance URL (default: https://gitlab.com)
- `--project-id`: GitLab project ID (required)
- `--mr-id`: Merge request ID (required)
- `--verbose`: Enable detailed logging

### Constants (internal/claude/client.go)
- Model: `claude-3-5-sonnet-20241022`
- Max tokens: `4000`
- Large diff threshold: `50000` characters (reviewer.go:71)

## Key Behaviors

### Review Logic
- Only reviews merge requests in "opened" state
- Skips deleted files in diff analysis
- For large changesets (>50k chars), provides summary instead of detailed review
- Detects primary programming language based on file extensions
- Formats diff with file paths and context for Claude

### Language Detection
Maps file extensions to language names for context-aware reviews:
- `.go` → Go, `.php` → PHP, `.js` → JavaScript, `.ts` → TypeScript, etc.
- Falls back to "general" if no mapping found

### Error Handling
- Validates required CLI parameters before proceeding
- Handles GitLab API errors (401, etc.) with descriptive messages
- Handles Claude API errors and empty responses
- Large diff handling prevents context window overflow

## Docker Deployment

The project uses multi-stage Docker builds for minimal production images:
- Builder stage: Uses golang:1.21-alpine with full build tools
- Runtime stage: Uses alpine:latest with just the binary
- Runs as non-root user for security
- Includes ca-certificates for HTTPS API calls

## GitLab CI/CD Integration

Designed to run in GitLab pipelines using built-in variables:
- `$CI_PROJECT_ID`: Project ID
- `$CI_MERGE_REQUEST_IID`: MR ID  
- `$CI_SERVER_URL`: GitLab URL
- Supports both manual and automatic trigger modes