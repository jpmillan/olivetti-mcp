package main

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/olivetti-mcp/jira"
	"github.com/olivetti-mcp/ticket"
)

// registerTools adds all MCP tools to the server.
func registerTools(s *server.MCPServer, provider ticket.Provider, loader *jira.TemplateLoader) {
	tool := mcp.NewTool("create_ticket",
		mcp.WithDescription(fmt.Sprintf("Create a %s ticket from a plain English description using predefined templates", provider.Name())),
		mcp.WithString("summary",
			mcp.Required(),
			mcp.Description("Short title for the Jira ticket"),
		),
		mcp.WithString("issue_type",
			mcp.Required(),
			mcp.Description("Type of issue: Story, Bug, or Task"),
			mcp.Enum("Story", "Bug", "Task"),
		),
		mcp.WithString("description",
			mcp.Required(),
			mcp.Description("Plain English description of the issue"),
		),
		mcp.WithString("acceptance_criteria",
			mcp.Description("Definition of done for this ticket"),
		),
		mcp.WithString("background",
			mcp.Description("Context or motivation behind this ticket"),
		),
		mcp.WithString("out_of_scope",
			mcp.Description("What this ticket does NOT cover"),
		),
		mcp.WithString("priority",
			mcp.Description("Priority level (e.g. High, Medium, Low). Defaults to template default"),
		),
		mcp.WithNumber("story_points",
			mcp.Description("Story point estimate (for stories: 1, 2, 3, 5, 8, 13)"),
		),
	)

	s.AddTool(tool, createTicketHandler(provider, loader))
}

func createTicketHandler(provider ticket.Provider, loader *jira.TemplateLoader) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract parameters from the request.
		args := request.GetArguments()

		summary, _ := args["summary"].(string)
		issueType, _ := args["issue_type"].(string)
		description, _ := args["description"].(string)
		acceptanceCriteria, _ := args["acceptance_criteria"].(string)
		background, _ := args["background"].(string)
		outOfScope, _ := args["out_of_scope"].(string)
		priority, _ := args["priority"].(string)

		var storyPoints *int
		if sp, ok := args["story_points"].(float64); ok {
			v := int(sp)
			storyPoints = &v
		}

		// Load the matching template.
		tmpl, err := loader.Get(issueType)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Build input map for validation and rendering.
		input := map[string]string{
			"summary":             summary,
			"description":         description,
			"acceptance_criteria": acceptanceCriteria,
			"background":          background,
			"out_of_scope":        outOfScope,
		}
		if storyPoints != nil {
			input["story_points"] = strconv.Itoa(*storyPoints)
		}

		// Validate required fields.
		if err := tmpl.ValidateFields(input); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Validation failed: %s", err.Error())), nil
		}

		// Validate story points against allowed options.
		if storyPoints != nil && len(tmpl.StoryPointsOptions) > 0 {
			if !slices.Contains(tmpl.StoryPointsOptions, *storyPoints) {
				return mcp.NewToolResultError(fmt.Sprintf(
					"Invalid story points %d; allowed values: %v", *storyPoints, tmpl.StoryPointsOptions,
				)), nil
			}
		}

		// Determine priority.
		if priority == "" {
			priority = tmpl.DefaultPriority
		}

		// Render the description from the template.
		renderedDescription := tmpl.RenderDescription(input)

		// Create the ticket via the configured provider.
		resp, err := provider.CreateTicket(ctx, ticket.CreateRequest{
			Summary:     summary,
			IssueType:   tmpl.IssueType,
			Description: renderedDescription,
			Priority:    priority,
			Labels:      tmpl.FieldLabels,
			StoryPoints: storyPoints,
		})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create ticket: %s", err.Error())), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("%s: %s", resp.Key, resp.URL)), nil
	}
}
