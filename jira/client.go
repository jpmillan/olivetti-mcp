package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/olivetti-mcp/ticket"
)

// Client communicates with the Jira REST API v3.
type Client struct {
	baseURL    string
	email      string
	apiToken   string
	projectKey string
	httpClient *http.Client
}

// NewClient creates a Jira API client with basic auth credentials.
func NewClient(baseURL, email, apiToken, projectKey string) *Client {
	return &Client{
		baseURL:    baseURL,
		email:      email,
		apiToken:   apiToken,
		projectKey: projectKey,
		httpClient: &http.Client{},
	}
}

// Name returns the provider name.
func (c *Client) Name() string { return "Jira" }

// CreateTicket posts a new issue to Jira and returns the created ticket info.
func (c *Client) CreateTicket(_ context.Context, req ticket.CreateRequest) (*ticket.CreateResponse, error) {
	payload := c.buildPayload(req)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshalling request: %w", err)
	}

	url := fmt.Sprintf("%s/rest/api/3/issue", c.baseURL)
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	httpReq.SetBasicAuth(c.email, c.apiToken)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("sending request to Jira: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Jira API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		ID   string `json:"id"`
		Key  string `json:"key"`
		Self string `json:"self"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing Jira response: %w", err)
	}

	return &ticket.CreateResponse{
		Key: result.Key,
		URL: fmt.Sprintf("%s/browse/%s", c.baseURL, result.Key),
		ID:  result.ID,
	}, nil
}

func (c *Client) buildPayload(req ticket.CreateRequest) map[string]any {
	// Jira v3 uses Atlassian Document Format (ADF) for descriptions.
	descriptionADF := map[string]any{
		"version": 1,
		"type":    "doc",
		"content": []any{
			map[string]any{
				"type": "paragraph",
				"content": []any{
					map[string]any{
						"type": "text",
						"text": req.Description,
					},
				},
			},
		},
	}

	fields := map[string]any{
		"project": map[string]string{
			"key": c.projectKey,
		},
		"summary": req.Summary,
		"issuetype": map[string]string{
			"name": req.IssueType,
		},
		"description": descriptionADF,
	}

	if req.Priority != "" {
		fields["priority"] = map[string]string{
			"name": req.Priority,
		}
	}

	if len(req.Labels) > 0 {
		fields["labels"] = req.Labels
	}

	if req.StoryPoints != nil {
		fields["story_points"] = *req.StoryPoints
	}

	return map[string]any{
		"fields": fields,
	}
}
