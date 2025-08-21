#!/bin/bash

# Test script for gitlab-claude-reviewer
# This script tests the basic functionality without making actual API calls

echo "🔨 Building gitlab-claude-reviewer..."

# Build the binary
go build -o gitlab-claude-reviewer ./cmd/gitlab-claude-reviewer

if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi

echo "✅ Build successful"

# Test help command
echo ""
echo "🧪 Testing help command..."
./gitlab-claude-reviewer --help

# Test missing required parameters
echo ""
echo "🧪 Testing missing parameters..."
./gitlab-claude-reviewer 2>&1 | grep -q "required"
if [ $? -eq 0 ]; then
    echo "✅ Correctly shows required parameter errors"
else
    echo "❌ Should show required parameter errors"
fi

# Test Docker build
echo ""
echo "🐳 Testing Docker build..."
docker build -t gitlab-claude-reviewer-test . > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo "✅ Docker build successful"
    
    # Test Docker run
    echo ""
    echo "🐳 Testing Docker run..."
    docker run --rm gitlab-claude-reviewer-test --help > /dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        echo "✅ Docker container runs successfully"
    else
        echo "❌ Docker container failed to run"
    fi
    
    # Clean up test image
    docker rmi gitlab-claude-reviewer-test > /dev/null 2>&1
else
    echo "❌ Docker build failed"
fi

echo ""
echo "🎉 All tests completed!"
echo ""
echo "📋 To test with real GitLab project:"
echo "   1. Set GITLAB_TOKEN and CLAUDE_API_KEY environment variables"
echo "   2. Run: ./gitlab-claude-reviewer --project-id=PROJECT_ID --mr-id=MR_ID"

# Clean up
rm -f gitlab-claude-reviewer