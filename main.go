package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/server"
	"github.com/olivetti-mcp/jira"
)

func main() {
	// Load .env file if present (optional — env vars can also be set directly).
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, using environment variables directly")
	}

	baseURL := requireEnv("JIRA_BASE_URL")
	email := requireEnv("JIRA_EMAIL")
	apiToken := requireEnv("JIRA_API_TOKEN")
	projectKey := getEnv("JIRA_PROJECT_KEY", "")
	templatesDir := getEnv("TEMPLATES_DIR", "./templates")

	// Load YAML issue templates from disk.
	loader := jira.NewTemplateLoader(templatesDir)
	if err := loader.Load(); err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}
	log.Printf("Ready — loaded templates: %v", loader.AvailableTypes())

	jiraClient := jira.NewClient(baseURL, email, apiToken, projectKey)
	log.Printf("Provider: %s", jiraClient.Name())

	s := server.NewMCPServer(
		"Olivetti",
		"1.0.0",
		server.WithToolCapabilities(true),
	)
	registerTools(s, jiraClient, loader)

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return val
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
