# Olivetti — Create Jira Tickets with AI

> Named after the iconic Italian typewriter brand — because good writing deserves good tooling.

Olivetti is an [MCP](https://modelcontextprotocol.io/) server that lets AI assistants (like GitHub Copilot, Claude, Cursor, etc.) create well-structured tickets on your behalf.

You describe what you need in plain English. Olivetti picks the right template, validates the fields, and creates the ticket in your project tracker — all without leaving your editor.

> **Currently supported:** Jira. The provider architecture makes it straightforward to add Linear, GitHub Issues, and others.

---

## Why Olivetti?

Writing Jira tickets by hand is tedious. You have to open the browser, pick the right issue type, fill out a dozen fields, and format the description. Olivetti removes that friction:

- **Talk naturally** — just describe the bug, feature, or task
- **Consistent formatting** — every ticket follows your team's templates
- **Stays in your flow** — works inside any MCP-compatible tool (VS Code, Claude Desktop, etc.)

## How It Works

```
You: "Create a story for adding CSV export to the reports page"
        ↓
Olivetti picks the Story template
        ↓
Validates required fields (summary, description, acceptance criteria…)
        ↓
Renders a structured Jira description from the template
        ↓
Posts the ticket via the Jira REST API (or your configured provider)
        ↓
Returns: PROJ-42 → https://yourteam.atlassian.net/browse/PROJ-42
```

## Quick Start

### Prerequisites

- **Go 1.23+** — [install Go](https://go.dev/dl/)
- **Jira Cloud** account with API access
- **Jira API token** — [generate one here](https://id.atlassian.com/manage-profile/security/api-tokens)

### 1. Clone and install

```bash
git clone https://github.com/YOUR_USERNAME/olivetti-mcp.git
cd olivetti-mcp
go mod tidy
```

### 2. Configure

Copy the example env file and fill in your Jira credentials:

```bash
cp .env.example .env
```

Then edit `.env`:

```env
JIRA_BASE_URL=https://yourteam.atlassian.net
JIRA_EMAIL=you@example.com
JIRA_API_TOKEN=your_api_token_here
JIRA_PROJECT_KEY=DEV
```

> **Security note:** The `.env` file is git-ignored and should never be committed.

### 3. Build

```bash
go build -o olivetti-mcp .
```

### 4. Connect to your MCP client

Olivetti uses **stdio transport**, which means your MCP client launches it as a subprocess. Here's how to configure popular clients:

<details>
<summary><strong>VS Code (GitHub Copilot)</strong></summary>

Add to your `.vscode/mcp.json` (or user `settings.json`):

```json
{
  "servers": {
    "olivetti": {
      "type": "stdio",
      "command": "/absolute/path/to/olivetti-mcp",
      "env": {
        "JIRA_BASE_URL": "https://yourteam.atlassian.net",
        "JIRA_EMAIL": "you@example.com",
        "JIRA_API_TOKEN": "your_api_token_here",
        "JIRA_PROJECT_KEY": "DEV"
      }
    }
  }
}
```

</details>

<details>
<summary><strong>Claude Desktop</strong></summary>

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "olivetti": {
      "command": "/absolute/path/to/olivetti-mcp",
      "env": {
        "JIRA_BASE_URL": "https://yourteam.atlassian.net",
        "JIRA_EMAIL": "you@example.com",
        "JIRA_API_TOKEN": "your_api_token_here",
        "JIRA_PROJECT_KEY": "DEV"
      }
    }
  }
}
```

</details>

<details>
<summary><strong>Any MCP client (generic)</strong></summary>

Olivetti reads configuration from environment variables. Set them however your client supports, then launch the binary. It communicates over stdin/stdout using the MCP protocol.

</details>

## Usage

Once connected, your AI assistant will have access to the `create_ticket` tool. Just ask it to create a ticket:

> "Create a bug for the login page crashing on Safari when the password field is empty"

The tool accepts these parameters:

| Parameter             | Required | Description                                      |
| --------------------- | -------- | ------------------------------------------------ |
| `summary`             | Yes      | Short title for the ticket                       |
| `issue_type`          | Yes      | `Story`, `Bug`, or `Task`                        |
| `description`         | Yes      | What the ticket is about                         |
| `acceptance_criteria` | No       | Definition of done                               |
| `background`          | No       | Why this work is needed                          |
| `out_of_scope`        | No       | What this ticket does **not** cover              |
| `priority`            | No       | `High`, `Medium`, or `Low` (defaults to template) |
| `story_points`        | No       | Estimate: 1, 2, 3, 5, 8, or 13 (stories only)  |

## Templates

Templates live in the `templates/` folder as YAML files. They control how each issue type is structured.

### Included templates

- **Story** — for features and user-facing work
- **Bug** — for defects and issues
- **Task** — for technical or operational work

### Creating your own

Add a new `.yaml` file to `templates/`. For example, `templates/epic.yaml`:

```yaml
issue_type: Epic
default_priority: Medium
required_fields:
  - summary
  - description
field_labels:
  - epic
description_template: |
  ## Summary
  {description}

  ## Background
  {background}

  ## Acceptance Criteria
  {acceptance_criteria}
```

Then add the new type to the `Enum(...)` list in [server.go](server.go) and rebuild.

### Template reference

| Field                  | Description                                           |
| ---------------------- | ----------------------------------------------------- |
| `issue_type`           | Jira issue type name (must match your Jira config)    |
| `default_priority`     | Priority used when none is specified                  |
| `required_fields`      | Fields the user must provide                          |
| `field_labels`         | Labels automatically applied to the ticket            |
| `description_template` | Markdown body with `{placeholder}` substitution       |
| `story_points_options` | Allowed story point values (optional)                 |

## Project Structure

```
olivetti-mcp/
├── main.go              # Entry point — loads config, starts MCP server
├── server.go            # Registers the MCP tool and handler
├── ticket/
│   └── provider.go      # Provider interface and shared types
├── jira/
│   ├── client.go        # Jira provider (implements ticket.Provider)
│   ├── templates.go     # YAML template loader and renderer
│   └── templates_test.go
├── templates/
│   ├── story.yaml       # Story template
│   ├── bug.yaml         # Bug template
│   └── task.yaml        # Task template
├── .env.example         # Sample environment config
├── .gitignore
├── LICENSE
├── go.mod
└── README.md
```

## Running Tests

```bash
go test ./...
```

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

To add a new provider (e.g. Linear, GitHub Issues), create a new package that implements the `ticket.Provider` interface:

```go
type Provider interface {
    Name() string
    CreateTicket(ctx context.Context, req ticket.CreateRequest) (*ticket.CreateResponse, error)
}
```

See [jira/client.go](jira/client.go) for a working example.

## License

[MIT](LICENSE)
