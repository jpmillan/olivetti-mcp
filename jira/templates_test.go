package jira

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestTemplates(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	storyYAML := `issue_type: Story
default_priority: Medium
required_fields:
  - summary
  - description
  - acceptance_criteria
  - story_points
field_labels:
  - team-backend
description_template: |
  ## Summary
  {description}

  ## Acceptance Criteria
  {acceptance_criteria}
story_points_options: [1, 2, 3, 5, 8, 13]
`
	if err := os.WriteFile(filepath.Join(dir, "story.yaml"), []byte(storyYAML), 0644); err != nil {
		t.Fatal(err)
	}

	bugYAML := `issue_type: Bug
default_priority: High
required_fields:
  - summary
  - description
field_labels:
  - bug
description_template: |
  ## Bug
  {description}
`
	if err := os.WriteFile(filepath.Join(dir, "bug.yaml"), []byte(bugYAML), 0644); err != nil {
		t.Fatal(err)
	}

	return dir
}

func TestTemplateLoader_Load(t *testing.T) {
	dir := setupTestTemplates(t)
	loader := NewTemplateLoader(dir)

	if err := loader.Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	types := loader.AvailableTypes()
	if len(types) != 2 {
		t.Fatalf("expected 2 templates, got %d: %v", len(types), types)
	}
}

func TestTemplateLoader_Get(t *testing.T) {
	dir := setupTestTemplates(t)
	loader := NewTemplateLoader(dir)
	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	tmpl, err := loader.Get("Story")
	if err != nil {
		t.Fatalf("Get(Story) failed: %v", err)
	}
	if tmpl.IssueType != "Story" {
		t.Errorf("expected IssueType=Story, got %q", tmpl.IssueType)
	}
	if tmpl.DefaultPriority != "Medium" {
		t.Errorf("expected DefaultPriority=Medium, got %q", tmpl.DefaultPriority)
	}
}

func TestTemplateLoader_Get_CaseInsensitive(t *testing.T) {
	dir := setupTestTemplates(t)
	loader := NewTemplateLoader(dir)
	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	if _, err := loader.Get("story"); err != nil {
		t.Errorf("Get(story) should work case-insensitively: %v", err)
	}
	if _, err := loader.Get("BUG"); err != nil {
		t.Errorf("Get(BUG) should work case-insensitively: %v", err)
	}
}

func TestTemplateLoader_Get_NotFound(t *testing.T) {
	dir := setupTestTemplates(t)
	loader := NewTemplateLoader(dir)
	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	_, err := loader.Get("Epic")
	if err == nil {
		t.Fatal("expected error for missing template, got nil")
	}
}

func TestTemplate_ValidateFields(t *testing.T) {
	dir := setupTestTemplates(t)
	loader := NewTemplateLoader(dir)
	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	tmpl, _ := loader.Get("Story")

	// Missing required fields.
	err := tmpl.ValidateFields(map[string]string{
		"summary":     "A title",
		"description": "Some description",
	})
	if err == nil {
		t.Fatal("expected validation error for missing fields")
	}

	// All required fields present.
	err = tmpl.ValidateFields(map[string]string{
		"summary":             "A title",
		"description":         "Some description",
		"acceptance_criteria": "Done when X",
		"story_points":        "5",
	})
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestTemplate_RenderDescription(t *testing.T) {
	dir := setupTestTemplates(t)
	loader := NewTemplateLoader(dir)
	if err := loader.Load(); err != nil {
		t.Fatal(err)
	}

	tmpl, _ := loader.Get("Story")
	rendered := tmpl.RenderDescription(map[string]string{
		"description":         "Build the widget",
		"acceptance_criteria": "Widget works correctly",
	})

	if rendered == "" {
		t.Fatal("rendered description is empty")
	}
	if !strings.Contains(rendered, "Build the widget") {
		t.Error("rendered description should contain the description text")
	}
	if !strings.Contains(rendered, "Widget works correctly") {
		t.Error("rendered description should contain acceptance criteria")
	}
}
