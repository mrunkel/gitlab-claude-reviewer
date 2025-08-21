# Setup Guide

This guide walks you through setting up GitLab Claude Reviewer in your GitLab project.

## Prerequisites

### 1. GitLab Personal Access Token

1. Go to your GitLab instance → **Settings** → **Access Tokens**
2. Create a new token with these scopes:
   - `api` (required for reading MR data and posting comments)
3. Copy the token and store it securely

### 2. Claude API Key

1. Visit [Anthropic Console](https://console.anthropic.com/)
2. Create an account or sign in
3. Navigate to API keys section
4. Create a new API key
5. Copy the key and store it securely

## Method 1: Docker Image (Recommended)

### Step 1: Build or Pull Docker Image

#### Option A: Build from Source
```bash
git clone https://github.com/your-org/gitlab-claude-reviewer.git
cd gitlab-claude-reviewer
docker build -t gitlab-claude-reviewer .
```

#### Option B: Use Pre-built Image (when available)
```bash
docker pull your-registry/gitlab-claude-reviewer:latest
```

### Step 2: Configure GitLab CI/CD Variables

In your GitLab project:

1. Go to **Settings** → **CI/CD** → **Variables**
2. Add these variables (mark as **Protected** and **Masked**):
   - `GITLAB_TOKEN`: Your GitLab personal access token
   - `CLAUDE_API_KEY`: Your Claude API key

### Step 3: Update .gitlab-ci.yml

Add this stage to your `.gitlab-ci.yml`:

```yaml
stages:
  - test
  - review  # Add this stage

ai-review:
  stage: review
  image: gitlab-claude-reviewer:latest  # Or your registry path
  script:
    - |
      if [ -n "$CI_MERGE_REQUEST_IID" ]; then
        gitlab-claude-reviewer \
          --project-id="$CI_PROJECT_ID" \
          --mr-id="$CI_MERGE_REQUEST_IID" \
          --gitlab-url="$CI_SERVER_URL" \
          --verbose
      else
        echo "Not a merge request, skipping AI review"
      fi
  only:
    - merge_requests
  when: manual  # Remove this line for automatic reviews
```

## Method 2: Using GitLab Container Registry

### Step 1: Build and Push to GitLab Registry

```bash
# In your project repository
git clone https://github.com/your-org/gitlab-claude-reviewer.git
cd gitlab-claude-reviewer

# Login to GitLab registry
docker login $CI_REGISTRY -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD

# Build and push
docker build -t $CI_REGISTRY/your-group/your-project/gitlab-claude-reviewer:latest .
docker push $CI_REGISTRY/your-group/your-project/gitlab-claude-reviewer:latest
```

### Step 2: Use in CI/CD

```yaml
ai-review:
  stage: review
  image: $CI_REGISTRY/your-group/your-project/gitlab-claude-reviewer:latest
  script:
    - |
      if [ -n "$CI_MERGE_REQUEST_IID" ]; then
        gitlab-claude-reviewer \
          --project-id="$CI_PROJECT_ID" \
          --mr-id="$CI_MERGE_REQUEST_IID" \
          --gitlab-url="$CI_SERVER_URL"
      fi
  only:
    - merge_requests
```

## Method 3: Manual/Local Usage

### Step 1: Download or Build Binary

```bash
# Option A: Build from source
git clone https://github.com/your-org/gitlab-claude-reviewer.git
cd gitlab-claude-reviewer
go build -o gitlab-claude-reviewer ./cmd/gitlab-claude-reviewer

# Option B: Download release binary (when available)
# wget https://github.com/your-org/gitlab-claude-reviewer/releases/latest/download/gitlab-claude-reviewer-linux-amd64
```

### Step 2: Set Environment Variables

```bash
export GITLAB_TOKEN="your_gitlab_personal_access_token"
export CLAUDE_API_KEY="your_claude_api_key"
```

### Step 3: Run Manual Review

```bash
./gitlab-claude-reviewer \
  --project-id="12345" \
  --mr-id="67" \
  --gitlab-url="https://gitlab.com" \
  --verbose
```

## Configuration Options

### Advanced CI/CD Configuration

#### Conditional Reviews
Only review certain file types:
```yaml
ai-review:
  stage: review
  image: gitlab-claude-reviewer:latest
  script:
    - |
      if [ -n "$CI_MERGE_REQUEST_IID" ]; then
        # Only review if certain files changed
        if git diff --name-only HEAD~1 | grep -E '\.(php|js|go|py)$'; then
          gitlab-claude-reviewer \
            --project-id="$CI_PROJECT_ID" \
            --mr-id="$CI_MERGE_REQUEST_IID" \
            --gitlab-url="$CI_SERVER_URL"
        else
          echo "No reviewable files changed"
        fi
      fi
  only:
    - merge_requests
```

#### Review on Schedule
Run reviews on a schedule instead of every MR:
```yaml
ai-review-scheduled:
  stage: review
  image: gitlab-claude-reviewer:latest
  script:
    - |
      # Review all open MRs (requires custom script)
      ./review-all-mrs.sh
  only:
    - schedules
```

### Environment-Specific Configuration

#### Self-Hosted GitLab
```yaml
ai-review:
  stage: review
  image: gitlab-claude-reviewer:latest
  variables:
    GITLAB_URL: "https://gitlab.company.com"
  script:
    - |
      gitlab-claude-reviewer \
        --project-id="$CI_PROJECT_ID" \
        --mr-id="$CI_MERGE_REQUEST_IID" \
        --gitlab-url="$GITLAB_URL"
  only:
    - merge_requests
```

## Testing Your Setup

### Step 1: Create Test MR

1. Create a small test merge request in your project
2. Make sure it has some code changes

### Step 2: Trigger Review

#### If using `when: manual`:
1. Go to your MR's pipeline
2. Click the manual "ai-review" job
3. Check the job logs

#### If automatic:
1. The review should start automatically
2. Check the pipeline status

### Step 3: Verify Results

1. Check the MR comments for Claude's review
2. Look for the 🤖 **Claude Code Review** comment
3. Verify the feedback makes sense

## Troubleshooting

### Common Setup Issues

1. **"GitLab API error 401"**
   - Check your `GITLAB_TOKEN` is correct
   - Verify token has `api` scope
   - Ensure token hasn't expired

2. **"Claude API error 401"**
   - Check your `CLAUDE_API_KEY` is correct
   - Verify you have API credits

3. **"Container not found"**
   - Build the Docker image first
   - Check image name matches in CI/CD config

4. **"No merge request found"**
   - Ensure you're testing on an actual merge request
   - Check project ID is correct

### Debug Mode

Enable verbose logging:

```yaml
ai-review:
  # ... other config
  script:
    - |
      gitlab-claude-reviewer \
        --project-id="$CI_PROJECT_ID" \
        --mr-id="$CI_MERGE_REQUEST_IID" \
        --gitlab-url="$CI_SERVER_URL" \
        --verbose  # Add this for detailed logs
```

### Local Testing

Test locally before deploying:

```bash
# Set your tokens
export GITLAB_TOKEN="your_token"
export CLAUDE_API_KEY="your_key"

# Test with a real MR
./gitlab-claude-reviewer \
  --project-id="your-project-id" \
  --mr-id="existing-mr-id" \
  --verbose
```

## Security Considerations

1. **Token Storage**: Always use GitLab CI/CD variables, never hardcode tokens
2. **Token Scope**: Use minimal required scopes (`api` for GitLab)
3. **Token Rotation**: Regularly rotate your API tokens
4. **Access Control**: Limit who can modify CI/CD variables
5. **Container Security**: Keep Docker images updated

## Next Steps

Once setup is complete:

1. **Customize Reviews**: Modify the review prompts in the source code
2. **Add Rules**: Create custom rules for when to run reviews
3. **Monitor Costs**: Track Claude API usage
4. **Team Training**: Train your team on interpreting AI feedback
5. **Feedback Loop**: Collect feedback and improve the tool