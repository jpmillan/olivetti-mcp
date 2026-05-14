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
			mcp.Description("Acceptance criteria in Given/When/Then format"),
		),
		mcp.WithString("non_functional_requirements",
			mcp.Description("Non-functional requirements: performance, security, accessibility, audit/logging"),
		),
		mcp.WithString("dependencies",
			mcp.Description("Dependencies: related stories/bugs, external systems, design links, API contracts"),
		),
		mcp.WithString("uat_requirement",
			mcp.Description("UAT details: whether UAT is required, scenarios covered, and UAT owner"),
		),
		mcp.WithString("testing_notes",
			mcp.Description("Testing notes: required test data, negative scenarios, regression impact areas"),
		),
		mcp.WithString("target_users",
			mcp.Description("Target users or personas for this work"),
		),
		mcp.WithString("environment",
			mcp.Description("Environment where the bug occurs: Dev/QA/Staging/Production, browser, user role"),
		),
		mcp.WithString("bug_description",
			mcp.Description("Detailed bug description: what is happening and what is wrong or unexpected"),
		),
		mcp.WithString("steps_to_reproduce",
			mcp.Description("Exact steps to reproduce the bug"),
		),
		mcp.WithString("expected_result",
			mcp.Description("What should happen (expected behaviour)"),
		),
		mcp.WithString("actual_result",
			mcp.Description("What actually happens (actual behaviour)"),
		),
		mcp.WithString("evidence",
			mcp.Description("Evidence: screenshots, screen recordings, logs, error messages, network responses"),
		),
		mcp.WithString("impact_assessment",
			mcp.Description("Impact assessment: severity, frequency, affected users, business impact"),
		),
		mcp.WithString("priority",
			mcp.Description("Priority level (e.g. High, Medium, Low). Defaults to template default"),
		),
		mcp.WithNumber("story_points",
			mcp.Description("Story point estimate (for stories: 1, 2, 3, 5, 8, 13)"),
		),
		mcp.WithString("project_key",
			mcp.Description("Jira project key (e.g. DEV, PROJ). Overrides the default configured via JIRA_PROJECT_KEY"),
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
		nonFunctionalReqs, _ := args["non_functional_requirements"].(string)
		dependencies, _ := args["dependencies"].(string)
		uatRequirement, _ := args["uat_requirement"].(string)
		testingNotes, _ := args["testing_notes"].(string)
		targetUsers, _ := args["target_users"].(string)
		environment, _ := args["environment"].(string)
		bugDescription, _ := args["bug_description"].(string)
		stepsToReproduce, _ := args["steps_to_reproduce"].(string)
		expectedResult, _ := args["expected_result"].(string)
		actualResult, _ := args["actual_result"].(string)
		evidence, _ := args["evidence"].(string)
		impactAssessment, _ := args["impact_assessment"].(string)
		priority, _ := args["priority"].(string)
		projectKey, _ := args["project_key"].(string)

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
			"summary":                     summary,
			"description":                 description,
			"acceptance_criteria":         acceptanceCriteria,
			"non_functional_requirements": nonFunctionalReqs,
			"dependencies":                dependencies,
			"uat_requirement":             uatRequirement,
			"testing_notes":               testingNotes,
			"target_users":                targetUsers,
			"environment":                 environment,
			"bug_description":             bugDescription,
			"steps_to_reproduce":          stepsToReproduce,
			"expected_result":             expectedResult,
			"actual_result":               actualResult,
			"evidence":                    evidence,
			"impact_assessment":           impactAssessment,
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
			ProjectKey:  projectKey,
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
