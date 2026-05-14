# Olivetti — Create Jira Tickets with AI

> Named after the iconic Italian typewriter brand — because good writing deserves good tooling.

Olivetti is an [MCP](https://modelcontextprotocol.io/) server that lets AI assistants (like GitHub Copilot, Claude, Cursor, etc.) create well-structured Jira tickets on your behalf — without leaving your editor.

You describe what you need in plain English. Olivetti picks the right template, validates the fields, formats the description, and creates the ticket via the Jira API.

> **Currently supported:** Jira Cloud. The provider architecture makes it straightforward to add Linear, GitHub Issues, and others.

---

## Table of Contents

- [Why Olivetti?](#why-olivetti)
- [How It Works](#how-it-works)
- [Quick Start](#quick-start)
  - [Prerequisites](#prerequisites)
  - [1. Clone and build](#1-clone-and-build)
  - [2. Configure Jira credentials](#2-configure-jira-credentials)
  - [3. Connect to VS Code (GitHub Copilot)](#3-connect-to-vs-code-github-copilot)
  - [Other MCP clients](#other-mcp-clients)
- [Usage](#usage)
  - [Examples](#examples)
  - [Tool parameters](#tool-parameters)
- [Configuration Reference](#configuration-reference)
- [Templates](#templates)
- [Project Structure](#project-structure)
- [Running Tests](#running-tests)
- [Contributing](#contributing)

---

## Why Olivetti?

Writing Jira tickets by hand is tedious — open the browser, pick the right issue type, fill out a dozen fields, format the description. Olivetti removes that friction:

- **Talk naturally** — just describe the bug, feature, or task in plain English
- **Consistent formatting** — every ticket follows your team's YAML templates
- **Stays in your flow** — works inside VS Code, Claude Desktop, or any MCP-compatible tool
- **Flexible project targeting** — set a default project or specify one per ticket

---

## How It Works

```
You (in AI chat): "Create a story for adding CSV export to the reports page"
                                        ↓
                    Olivetti picks the Story template
                                        ↓
                Validates required fields (summary, description, etc.)
                                        ↓
            Renders a structured Jira description from the template
                                        ↓
              Posts the ticket via the Jira REST API (v3)
                                        ↓
        Returns: PROJ-42 → https://yourteam.atlassian.net/browse/PROJ-42
```

---

## Quick Start

### Prerequisites

| Requirement | Details |
|---|---|
| **Go 1.23+** | [Download Go](https://go.dev/dl/) |
| **Jira Cloud account** | With API access enabled for your user |
| **Jira API token** | [Generate one here](https://id.atlassian.com/manage-profile/security/api-tokens) |

### 1. Clone and build

```bash
git clone https://github.com/YOUR_USERNAME/olivetti-mcp.git
cd olivetti-mcp
go mod tidy
go build -o olivetti-mcp .
```

This produces an `olivetti-mcp` binary (or `olivetti-mcp.exe` on Windows) in the project directory. Note the **absolute path** to this file — you'll need it in the next step.

### 2. Configure Jira credentials

Olivetti reads Jira credentials from environment variables. You have two options:

**Option A — Pass them inline** (recommended for MCP client configs like VS Code):
Set the env vars directly in your MCP client configuration (see step 3 below).

**Option B — Use a `.env` file** (useful for local development/testing):

```bash
cp .env.example .env
```

Then edit `.env` with your values:

```env
JIRA_BASE_URL=https://yourteam.atlassian.net
JIRA_EMAIL=you@example.com
JIRA_API_TOKEN=your_api_token_here
JIRA_PROJECT_KEY=DEV                    # optional — can be provided per ticket instead
# TEMPLATES_DIR=./templates             # optional — defaults to ./templates
```

> **Security:** The `.env` file is git-ignored and should **never** be committed. Keep your API token secret.

### 3. Connect to VS Code (GitHub Copilot)

This is the most common setup. Olivetti runs as a subprocess launched by VS Code via the MCP protocol.

**Step 1 —** Open your project in VS Code.

**Step 2 —** Create the file `.vscode/mcp.json` in your workspace root with the following content:

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

Replace the placeholder values:

| Field | What to put |
|---|---|
| `command` | The **absolute path** to the `olivetti-mcp` binary you built in step 1 |
| `JIRA_BASE_URL` | Your Jira Cloud URL, e.g. `https://yourteam.atlassian.net` |
| `JIRA_EMAIL` | The email you log in to Jira with |
| `JIRA_API_TOKEN` | The API token you generated in the prerequisites |
| `JIRA_PROJECT_KEY` | Your default Jira project key (e.g. `DEV`, `PROJ`). Optional — you can provide it per ticket instead |

> **Windows users:** Use forward slashes in the `command` path (e.g. `C:/Users/you/olivetti-mcp.exe`).

#### Global setup (all workspaces)

If you want Olivetti available in **every** VS Code workspace without creating a `.vscode/mcp.json` in each project:

1. Press `Ctrl+Shift+P` (or `Cmd+Shift+P` on Mac) and select **Preferences: Open User Settings (JSON)**.
2. Add the `mcp` block at the top level of the JSON:

```json
{
  "mcp": {
    "servers": {
      "olivetti": {
        "type": "stdio",
        "command": "C:/absolute/path/to/olivetti-mcp.exe",
        "env": {
          "JIRA_BASE_URL": "https://yourteam.atlassian.net",
          "JIRA_EMAIL": "you@example.com",
          "JIRA_API_TOKEN": "your_api_token_here",
          "JIRA_PROJECT_KEY": "DEV"
        }
      }
    }
  }
}
```

3. Save and reload VS Code.

> **Note:** If a workspace also has its own `.vscode/mcp.json` with an `olivetti` entry, the workspace-level config takes precedence over the user-level one for that workspace.

**Step 3 —** Reload VS Code: press `Ctrl+Shift+P` (or `Cmd+Shift+P` on Mac) and run **Developer: Reload Window**.

**Step 4 —** Open **Copilot Chat** and switch to **Agent mode**. You should see `create_ticket` in the available tools list. If you don't, check the Output panel (`Ctrl+Shift+U` → select "MCP" from the dropdown) for error messages.

**Step 5 —** Try it! Type something like:

> *"Create a bug for the login page crashing on Safari when the password field is empty"*

Copilot will call Olivetti, which validates the input, renders the description, and creates the Jira ticket. You'll get back the ticket key and a link.

### Other MCP clients

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

Olivetti communicates over **stdin/stdout** using the MCP protocol. Set the required environment variables however your client supports them, then launch the `olivetti-mcp` binary as a subprocess.

</details>

---

## Usage

Once connected, your AI assistant has access to the `create_ticket` tool. Just describe the ticket in natural language — the AI will extract the structured fields and call the tool.

### Examples

| You say | What happens |
|---|---|
| *"Create a bug for the login page crashing on Safari"* | Creates a **Bug** ticket with the default priority (High) |
| *"Story for adding CSV export to the reports page, 5 story points"* | Creates a **Story** with 5 story points |
| *"Task to upgrade Node.js to v20 in the CI pipeline, low priority"* | Creates a **Task** with Low priority |
| *"Create a bug in the MOBILE project for push notifications not arriving"* | Creates a **Bug** in the `MOBILE` project (overrides the default project key) |

### Tool parameters

| Parameter | Required | Description |
|---|:---:|---|
| `summary` | Yes | Short title for the ticket |
| `issue_type` | Yes | `Story`, `Bug`, or `Task` |
| `description` | Yes | Plain English description of the issue |
| `acceptance_criteria` | No | Definition of done for this ticket |
| `background` | No | Context or motivation behind this ticket |
| `out_of_scope` | No | What this ticket does **not** cover |
| `priority` | No | `High`, `Medium`, or `Low` (defaults to the template's default) |
| `story_points` | No | Estimate: `1`, `2`, `3`, `5`, `8`, or `13` (stories only) |
| `project_key` | No | Jira project key (e.g. `DEV`). Overrides the default `JIRA_PROJECT_KEY` |

> **Tip:** You don't need to specify every parameter. The AI assistant will map your natural language to the right fields. Required fields that are missing will produce a clear validation error.

---

## Configuration Reference

Olivetti is configured via environment variables. Set them in your MCP client config, a `.env` file, or your shell environment.

| Variable | Required | Default | Description |
|---|:---:|---|---|
| `JIRA_BASE_URL` | Yes | — | Your Jira Cloud URL (e.g. `https://yourteam.atlassian.net`) |
| `JIRA_EMAIL` | Yes | — | Email address for Jira authentication |
| `JIRA_API_TOKEN` | Yes | — | Jira API token ([generate here](https://id.atlassian.com/manage-profile/security/api-tokens)) |
| `JIRA_PROJECT_KEY` | No | — | Default Jira project key (e.g. `DEV`). Can be overridden per ticket via the `project_key` parameter |
| `TEMPLATES_DIR` | No | `./templates` | Path to the directory containing YAML issue templates |

---

## Templates

Templates live in the `templates/` folder as YAML files. They define the structure, required fields, and description format for each issue type.

### Included templates

| Template | File | Default Priority | Required Fields |
|---|---|---|---|
| **Story** | `templates/story.yaml` | Medium | summary, description, acceptance_criteria, story_points |
| **Bug** | `templates/bug.yaml` | High | summary, description |
| **Task** | `templates/task.yaml` | Medium | summary, description |

### Creating your own

1. Add a new `.yaml` file to `templates/`. For example, `templates/epic.yaml`:

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

2. In `server.go`, add the new type to the `Enum(...)` list on the `issue_type` parameter:

    ```go
    mcp.Enum("Story", "Bug", "Task", "Epic"),
    ```

3. Rebuild: `go build -o olivetti-mcp .`

### Template field reference

| Field | Description |
|---|---|
| `issue_type` | Jira issue type name — must match your Jira project's configured issue types |
| `default_priority` | Priority used when the user doesn't specify one |
| `required_fields` | List of fields the user must provide (validation fails if missing) |
| `field_labels` | Labels automatically applied to the created ticket |
| `description_template` | Markdown body with `{placeholder}` tokens that get replaced with user input |
| `story_points_options` | Allowed story point values, e.g. `[1, 2, 3, 5, 8, 13]` (optional) |

Available placeholders: `{summary}`, `{description}`, `{acceptance_criteria}`, `{background}`, `{out_of_scope}`, `{story_points}`

---

## Project Structure

```
olivetti-mcp/
├── main.go                # Entry point — loads config, boots MCP server
├── server.go              # MCP tool definition and handler logic
├── ticket/
│   └── provider.go        # Provider interface (swap Jira for any backend)
├── jira/
│   ├── client.go          # Jira REST API v3 client
│   ├── templates.go       # YAML template loader and renderer
│   └── templates_test.go  # Template tests
├── templates/
│   ├── story.yaml
│   ├── bug.yaml
│   └── task.yaml
├── .env.example           # Sample environment config
├── go.mod
└── README.md
```

---

## Running Tests

```bash
go test ./...
```

---

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

To add a new ticket provider (e.g. Linear, GitHub Issues), create a new package that implements the `ticket.Provider` interface:

```go
type Provider interface {
    Name() string
    CreateTicket(ctx context.Context, req CreateRequest) (*CreateResponse, error)
}
```

Then wire it up in `main.go` as an alternative to `jira.NewClient`.

See [jira/client.go](jira/client.go) for a working example.

## License

[MIT](LICENSE)
