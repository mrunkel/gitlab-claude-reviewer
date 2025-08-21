# GitLab Claude Reviewer 🤖

An AI-powered code review tool that integrates Claude AI with GitLab merge requests to provide automated, intelligent code reviews.

## Features

- 🔍 **Automated Code Review**: Claude AI analyzes your merge requests and provides intelligent feedback
- 🚀 **GitLab CI/CD Integration**: Seamlessly integrates with your existing GitLab pipelines
- 🐳 **Docker Support**: Containerized for easy deployment across different environments
- 🔒 **Secure**: Uses GitLab and Claude API tokens for secure access
- 📊 **Smart Analysis**: Detects programming languages and provides context-aware reviews
- 💬 **GitLab Comments**: Posts reviews directly as merge request comments
- ⚡ **Fast**: Lightweight Go binary with minimal resource usage

## Quick Start

### Prerequisites

1. **GitLab Personal Access Token** with `api` scope
2. **Claude API Key** from Anthropic
3. Docker (for containerized usage) or Go 1.21+ (for building from source)

### Using Docker (Recommended)

```bash
# Build the Docker image
git clone <this-repo>
cd gitlab-claude-reviewer
docker build -t gitlab-claude-reviewer .

# Run a review
docker run --rm \
  -e GITLAB_TOKEN="your_gitlab_token" \
  -e CLAUDE_API_KEY="your_claude_api_key" \
  gitlab-claude-reviewer \
  --project-id="12345" \
  --mr-id="67" \
  --gitlab-url="https://gitlab.com" \
  --verbose
```

### Building from Source

```bash
# Clone and build
git clone <this-repo>
cd gitlab-claude-reviewer
go build -o gitlab-claude-reviewer ./cmd/gitlab-claude-reviewer

# Run a review
export GITLAB_TOKEN="your_gitlab_token"
export CLAUDE_API_KEY="your_claude_api_key"

./gitlab-claude-reviewer \
  --project-id="12345" \
  --mr-id="67" \
  --gitlab-url="https://gitlab.com"
```

## GitLab CI/CD Integration

Add this stage to your `.gitlab-ci.yml`:

```yaml
stages:
  - test
  - review

ai-review:
  stage: review
  image: your-registry/gitlab-claude-reviewer:latest
  script:
    - |
      if [ -n "$CI_MERGE_REQUEST_IID" ]; then
        gitlab-claude-reviewer \
          --project-id="$CI_PROJECT_ID" \
          --mr-id="$CI_MERGE_REQUEST_IID" \
          --gitlab-url="$CI_SERVER_URL" \
          --verbose
      fi
  only:
    - merge_requests
  when: manual  # Remove for automatic reviews
```

**Required CI/CD Variables** (Settings → CI/CD → Variables):
- `GITLAB_TOKEN`: Personal access token with `api` scope
- `CLAUDE_API_KEY`: Your Claude API key

## Command Line Options

```
Usage: gitlab-claude-reviewer [options]

Options:
  --gitlab-token string    GitLab personal access token (or GITLAB_TOKEN env var)
  --claude-token string    Claude API token (or CLAUDE_API_KEY env var)
  --gitlab-url string      GitLab instance URL (default "https://gitlab.com")
  --project-id string      GitLab project ID (required)
  --mr-id string          Merge request ID to review (required)
  --verbose               Enable verbose logging
```

## Configuration

### GitLab Token Setup

1. Go to GitLab → Settings → Access Tokens
2. Create a token with `api` scope
3. Set as `GITLAB_TOKEN` environment variable or use `--gitlab-token` flag

### Claude API Key Setup

1. Get an API key from [Anthropic Console](https://console.anthropic.com/)
2. Set as `CLAUDE_API_KEY` environment variable or use `--claude-token` flag

## What Claude Reviews

Claude provides intelligent feedback on:

- **Code Quality**: Readability, maintainability, and best practices
- **Security**: Potential vulnerabilities and security concerns
- **Performance**: Optimization opportunities and bottlenecks
- **Bugs**: Potential issues and edge cases
- **Architecture**: Design patterns and structural improvements
- **Language-Specific**: Framework and language best practices

## Examples

### Review a Specific MR

```bash
gitlab-claude-reviewer \
  --project-id="gitlab-org/gitlab" \
  --mr-id="123456" \
  --verbose
```

### Self-Hosted GitLab

```bash
gitlab-claude-reviewer \
  --gitlab-url="https://gitlab.company.com" \
  --project-id="internal/my-project" \
  --mr-id="42"
```

### Using with Different Projects

```bash
# PHP/Laravel project
gitlab-claude-reviewer --project-id="company/laravel-app" --mr-id="15"

# Go project  
gitlab-claude-reviewer --project-id="team/go-service" --mr-id="28"

# JavaScript/React project
gitlab-claude-reviewer --project-id="frontend/react-app" --mr-id="91"
```

## Docker Registry

### Building and Pushing

```bash
# Build for multiple architectures
docker buildx build --platform linux/amd64,linux/arm64 -t your-registry/gitlab-claude-reviewer:latest --push .

# Or build for single architecture
docker build -t your-registry/gitlab-claude-reviewer:latest .
docker push your-registry/gitlab-claude-reviewer:latest
```

### Using in CI/CD

```yaml
ai-review:
  stage: review
  image: your-registry/gitlab-claude-reviewer:latest
  # ... rest of configuration
```

## Cost Considerations

- **Claude API**: ~$0.008 per 200-line diff review (using Claude 3 Haiku pricing)
- **Typical Usage**: ~$2-5 per month for small to medium teams
- **Large Projects**: Consider reviewing only critical files for cost optimization

## Troubleshooting

### Common Issues

1. **"GitLab API error 401"**: Check your GitLab token permissions
2. **"Claude API error 401"**: Verify your Claude API key
3. **"No changes detected"**: MR might be empty or already merged
4. **"Large changeset detected"**: MR has too many changes; consider smaller MRs

### Debug Mode

Use `--verbose` flag to see detailed logging:

```bash
gitlab-claude-reviewer --verbose --project-id="123" --mr-id="45"
```

### Testing Locally

```bash
# Using docker-compose
cp docs/docker-compose.yml .
echo "GITLAB_TOKEN=your_token" > .env
echo "CLAUDE_API_KEY=your_key" >> .env
echo "PROJECT_ID=123" >> .env
echo "MR_ID=45" >> .env
docker-compose up
```

## Security

- API tokens are passed via environment variables
- No tokens are logged or stored
- All communication uses HTTPS
- Docker image runs as non-root user

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

[Add your license here]

## Support

- 🐛 **Issues**: Report bugs via GitLab issues
- 💡 **Feature Requests**: Use issue templates
- 📖 **Documentation**: Check the `docs/` folder
- 💬 **Discussions**: Use GitLab discussions for questions