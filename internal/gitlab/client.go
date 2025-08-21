package gitlab

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	baseURL string
	token   string
	client  *http.Client
}

type MergeRequest struct {
	ID          int    `json:"id"`
	IID         int    `json:"iid"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	Author      User   `json:"author"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

type DiffFile struct {
	OldPath     string `json:"old_path"`
	NewPath     string `json:"new_path"`
	AMode       string `json:"a_mode"`
	BMode       string `json:"b_mode"`
	NewFile     bool   `json:"new_file"`
	RenamedFile bool   `json:"renamed_file"`
	DeletedFile bool   `json:"deleted_file"`
	Diff        string `json:"diff"`
}

type Note struct {
	Body string `json:"body"`
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		token:   token,
		client:  &http.Client{},
	}
}

func (c *Client) GetMergeRequest(projectID, mrID string) (*MergeRequest, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests/%s", c.baseURL, url.QueryEscape(projectID), mrID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitLab API error %d: %s", resp.StatusCode, string(body))
	}
	
	var mr MergeRequest
	if err := json.NewDecoder(resp.Body).Decode(&mr); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	
	return &mr, nil
}

func (c *Client) GetMergeRequestDiff(projectID, mrID string) ([]DiffFile, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests/%s/changes", c.baseURL, url.QueryEscape(projectID), mrID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitLab API error %d: %s", resp.StatusCode, string(body))
	}
	
	var response struct {
		Changes []DiffFile `json:"changes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	
	return response.Changes, nil
}

func (c *Client) PostMergeRequestNote(projectID, mrID, body string) error {
	url := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests/%s/notes", c.baseURL, url.QueryEscape(projectID), mrID)
	
	note := Note{Body: body}
	jsonData, err := json.Marshal(note)
	if err != nil {
		return fmt.Errorf("marshaling note: %w", err)
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitLab API error %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// FormatDiffForReview formats the diff files into a readable format for Claude
func FormatDiffForReview(files []DiffFile) string {
	var builder strings.Builder
	
	for _, file := range files {
		if file.DeletedFile {
			continue // Skip deleted files for review
		}
		
		builder.WriteString(fmt.Sprintf("\n## File: %s\n", file.NewPath))
		
		if file.NewFile {
			builder.WriteString("**(New file)**\n")
		} else if file.RenamedFile {
			builder.WriteString(fmt.Sprintf("**(Renamed from: %s)**\n", file.OldPath))
		}
		
		if file.Diff != "" {
			builder.WriteString("```diff\n")
			builder.WriteString(file.Diff)
			builder.WriteString("\n```\n")
		}
	}
	
	return builder.String()
}