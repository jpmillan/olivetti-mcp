package ticket

import "context"

// Provider is the interface that any ticket-management backend must implement.
// Jira, Linear, GitHub Issues, etc. each get their own package with a type
// that satisfies this interface.
type Provider interface {
	// Name returns a human-readable name for this provider (e.g. "Jira", "Linear").
	Name() string

	// CreateTicket creates a new ticket/issue and returns its key and URL.
	CreateTicket(ctx context.Context, req CreateRequest) (*CreateResponse, error)
}

// CreateRequest holds the provider-agnostic fields for creating a ticket.
type CreateRequest struct {
	ProjectKey  string
	Summary     string
	IssueType   string
	Description string
	Priority    string
	Labels      []string
	StoryPoints *int
}

// CreateResponse holds the result returned after a ticket is created.
type CreateResponse struct {
	ID  string // Provider's internal ID (e.g. "10042")
	Key string // Human-readable key (e.g. "PROJ-42", "#42")
	URL string // Browseable link to the ticket
}
