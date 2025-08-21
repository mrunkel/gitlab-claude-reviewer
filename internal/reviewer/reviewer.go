package reviewer

import (
	"fmt"
	"log"
	"strings"

	"gitlab-claude-reviewer/internal/claude"
	"gitlab-claude-reviewer/internal/gitlab"
)

type Reviewer struct {
	gitlabClient *gitlab.Client
	claudeClient *claude.Client
	verbose      bool
}

func New(gitlabClient *gitlab.Client, claudeClient *claude.Client) *Reviewer {
	return &Reviewer{
		gitlabClient: gitlabClient,
		claudeClient: claudeClient,
		verbose:      false,
	}
}

func (r *Reviewer) SetVerbose(verbose bool) {
	r.verbose = verbose
}

func (r *Reviewer) log(format string, args ...interface{}) {
	if r.verbose {
		log.Printf(format, args...)
	}
}

func (r *Reviewer) ReviewMergeRequest(projectID, mrID string) error {
	r.log("Fetching merge request %s from project %s", mrID, projectID)
	
	// Get merge request details
	mr, err := r.gitlabClient.GetMergeRequest(projectID, mrID)
	if err != nil {
		return fmt.Errorf("fetching merge request: %w", err)
	}
	
	r.log("MR Title: %s", mr.Title)
	r.log("MR State: %s", mr.State)
	
	// Skip if MR is not open
	if mr.State != "opened" {
		return fmt.Errorf("merge request is not open (state: %s)", mr.State)
	}
	
	// Get the diff
	r.log("Fetching merge request diff")
	diffFiles, err := r.gitlabClient.GetMergeRequestDiff(projectID, mrID)
	if err != nil {
		return fmt.Errorf("fetching merge request diff: %w", err)
	}
	
	if len(diffFiles) == 0 {
		r.log("No changes found in merge request")
		return r.gitlabClient.PostMergeRequestNote(projectID, mrID, "🤖 **Claude Code Review**\n\nNo code changes detected in this merge request.")
	}
	
	r.log("Found %d changed files", len(diffFiles))
	
	// Format diff for Claude
	diffContent := gitlab.FormatDiffForReview(diffFiles)
	
	// Check if diff is too large
	if len(diffContent) > 50000 { // Rough limit to avoid hitting Claude's context window
		return r.handleLargeDiff(projectID, mrID, diffFiles)
	}
	
	// Prepare review request
	reviewReq := claude.ReviewRequest{
		Title:       mr.Title,
		Description: mr.Description,
		DiffContent: diffContent,
		Language:    r.detectLanguage(diffFiles),
	}
	
	r.log("Sending code to Claude for review")
	
	// Get review from Claude
	review, err := r.claudeClient.ReviewCode(reviewReq)
	if err != nil {
		return fmt.Errorf("getting review from Claude: %w", err)
	}
	
	// Post review as comment
	reviewComment := r.formatReviewComment(review)
	
	r.log("Posting review comment to merge request")
	
	err = r.gitlabClient.PostMergeRequestNote(projectID, mrID, reviewComment)
	if err != nil {
		return fmt.Errorf("posting review comment: %w", err)
	}
	
	return nil
}

func (r *Reviewer) handleLargeDiff(projectID, mrID string, diffFiles []gitlab.DiffFile) error {
	// For large diffs, provide a summary instead of detailed review
	summary := r.createDiffSummary(diffFiles)
	
	comment := fmt.Sprintf(`🤖 **Claude Code Review**

⚠️ **Large changeset detected** (%d files changed)

This merge request contains a significant number of changes. Here's a summary:

%s

**Recommendation:** Consider breaking this into smaller merge requests for better reviewability.

To get a detailed review of specific files, please mention me with the file paths you'd like me to focus on.`, len(diffFiles), summary)
	
	return r.gitlabClient.PostMergeRequestNote(projectID, mrID, comment)
}

func (r *Reviewer) createDiffSummary(diffFiles []gitlab.DiffFile) string {
	var summary strings.Builder
	
	fileTypes := make(map[string]int)
	newFiles := 0
	deletedFiles := 0
	modifiedFiles := 0
	
	for _, file := range diffFiles {
		if file.NewFile {
			newFiles++
		} else if file.DeletedFile {
			deletedFiles++
		} else {
			modifiedFiles++
		}
		
		// Count by file extension
		parts := strings.Split(file.NewPath, ".")
		if len(parts) > 1 {
			ext := parts[len(parts)-1]
			fileTypes[ext]++
		}
	}
	
	summary.WriteString("### Changes Summary\n")
	if newFiles > 0 {
		summary.WriteString(fmt.Sprintf("- 🆕 %d new files\n", newFiles))
	}
	if modifiedFiles > 0 {
		summary.WriteString(fmt.Sprintf("- ✏️ %d modified files\n", modifiedFiles))
	}
	if deletedFiles > 0 {
		summary.WriteString(fmt.Sprintf("- 🗑️ %d deleted files\n", deletedFiles))
	}
	
	if len(fileTypes) > 0 {
		summary.WriteString("\n### File Types\n")
		for ext, count := range fileTypes {
			summary.WriteString(fmt.Sprintf("- `.%s`: %d files\n", ext, count))
		}
	}
	
	return summary.String()
}

func (r *Reviewer) detectLanguage(diffFiles []gitlab.DiffFile) string {
	// Simple language detection based on file extensions
	extensions := make(map[string]int)
	
	for _, file := range diffFiles {
		if file.DeletedFile {
			continue
		}
		
		parts := strings.Split(file.NewPath, ".")
		if len(parts) > 1 {
			ext := strings.ToLower(parts[len(parts)-1])
			extensions[ext]++
		}
	}
	
	// Return the most common extension
	maxCount := 0
	mostCommon := "general"
	
	for ext, count := range extensions {
		if count > maxCount {
			maxCount = count
			mostCommon = ext
		}
	}
	
	// Map extensions to language names
	languageMap := map[string]string{
		"go":   "Go",
		"php":  "PHP",
		"js":   "JavaScript",
		"ts":   "TypeScript",
		"py":   "Python",
		"java": "Java",
		"rs":   "Rust",
		"cpp":  "C++",
		"c":    "C",
		"rb":   "Ruby",
		"sql":  "SQL",
		"html": "HTML",
		"css":  "CSS",
		"scss": "SCSS",
		"vue":  "Vue.js",
		"jsx":  "React JSX",
		"tsx":  "React TSX",
	}
	
	if lang, exists := languageMap[mostCommon]; exists {
		return lang
	}
	
	return "general"
}

func (r *Reviewer) formatReviewComment(review string) string {
	return fmt.Sprintf("🤖 **Claude Code Review**\n\n%s\n\n---\n*Automated review by Claude AI*", review)
}