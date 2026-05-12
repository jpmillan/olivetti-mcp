package jira

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Template represents a Jira issue template loaded from a YAML file.
type Template struct {
	IssueType           string   `yaml:"issue_type"`
	DefaultPriority     string   `yaml:"default_priority"`
	RequiredFields      []string `yaml:"required_fields"`
	FieldLabels         []string `yaml:"field_labels"`
	DescriptionTemplate string   `yaml:"description_template"`
	StoryPointsOptions  []int    `yaml:"story_points_options"`
}

// TemplateLoader loads and caches issue templates from a directory.
type TemplateLoader struct {
	dir       string
	templates map[string]*Template
}

// NewTemplateLoader creates a loader that reads YAML files from dir.
func NewTemplateLoader(dir string) *TemplateLoader {
	return &TemplateLoader{
		dir:       dir,
		templates: make(map[string]*Template),
	}
}

// Load reads all YAML templates from the configured directory.
func (tl *TemplateLoader) Load() error {
	entries, err := os.ReadDir(tl.dir)
	if err != nil {
		return fmt.Errorf("reading templates directory %q: %w", tl.dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !isYAMLFile(entry.Name()) {
			continue
		}

		tmpl, err := parseTemplateFile(filepath.Join(tl.dir, entry.Name()))
		if err != nil {
			return fmt.Errorf("parsing template %q: %w", entry.Name(), err)
		}

		key := strings.ToLower(tmpl.IssueType)
		tl.templates[key] = tmpl
	}

	return nil
}

// Get returns the template for the given issue type (case-insensitive).
func (tl *TemplateLoader) Get(issueType string) (*Template, error) {
	tmpl, ok := tl.templates[strings.ToLower(issueType)]
	if !ok {
		available := tl.AvailableTypes()
		return nil, fmt.Errorf("no template found for issue type %q; available types: %s",
			issueType, strings.Join(available, ", "))
	}
	return tmpl, nil
}

// AvailableTypes returns a sorted list of loaded issue type names.
func (tl *TemplateLoader) AvailableTypes() []string {
	types := make([]string, 0, len(tl.templates))
	for _, tmpl := range tl.templates {
		types = append(types, tmpl.IssueType)
	}
	return types
}

// ValidateFields checks that all required fields for a template are present in the input.
func (t *Template) ValidateFields(input map[string]string) error {
	var missing []string
	for _, field := range t.RequiredFields {
		if val, ok := input[field]; !ok || strings.TrimSpace(val) == "" {
			missing = append(missing, field)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}
	return nil
}

// RenderDescription substitutes {placeholders} in the description template with input values.
func (t *Template) RenderDescription(input map[string]string) string {
	result := t.DescriptionTemplate
	for key, value := range input {
		placeholder := "{" + key + "}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

func parseTemplateFile(path string) (*Template, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var tmpl Template
	if err := yaml.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("unmarshalling YAML: %w", err)
	}

	if tmpl.IssueType == "" {
		return nil, fmt.Errorf("template at %q is missing issue_type", path)
	}

	return &tmpl, nil
}

func isYAMLFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".yaml" || ext == ".yml"
}
