# Commands

This document covers the commands provided by `gh-member`, structured according to the [Diátaxis](https://diataxis.fr/) framework:

- [Tutorial](#tutorial-browsing-your-first-organization): step-by-step walkthrough for new users
- [How-to guides](#how-to-guides): recipes for common tasks
- [Reference](#reference): complete flag and argument descriptions
- [Explanation](#explanation): background on output modes and API design

---

## Tutorial: Browsing your first organization

This tutorial walks you through installing `gh-member` and using the interactive member browser for the first time. You will end the tutorial knowing how to navigate the TUI, search for a member, and open their GitHub profile.

### Prerequisites

- [GitHub CLI](https://cli.github.com/) installed and authenticated (`gh auth login`)
- `gh-member` installed (`gh extension install tpu-kanglabs/gh-member`)

### Step 1 — Open the interactive browser

Run the command without any arguments:

```bash
gh member list
```

If your account belongs to more than one organization, an organization selector appears. Use **↑/↓** to highlight an org and press **Enter** to confirm.

If you belong to only one organization, the member table opens immediately.

### Step 2 — Explore the member table

The table shows four columns for every member:

| Column  | Content                              |
|---------|--------------------------------------|
| NAME    | Display name (falls back to login)   |
| ID      | GitHub login (username)              |
| ROLE    | `ADMIN` or `MEMBER`                  |
| PROFILE | Profile URL                          |

Use **↑/↓** or **j/k** to move between rows.

### Step 3 — Search for a member

Start typing any part of a member's name or login. The table filters in real time. Press **Backspace** to delete characters and widen the results.

### Step 4 — Open a member's GitHub profile

Highlight any row and press **Enter**. The profile URL opens in your default browser.

### Step 5 — Quit

Press **q** (when the search box is empty), **Esc**, or **Ctrl+C** to exit.

---

## How-to guides

### List members of a specific organization

Pass the organization login as a positional argument to skip the interactive selector:

```bash
gh member list my-org
```

### Limit the number of results

Use `--limit` (or `-L`) to cap the number of members returned. The default is 30.

```bash
gh member list my-org --limit 10
```

To retrieve every member regardless of count, pass `-1`:

```bash
gh member list my-org --limit -1
```

### Filter by role

Use `--role` to show only admins or only regular members:

```bash
# Admins only
gh member list my-org --role admin

# Regular members only
gh member list my-org --role member
```

### Export data as JSON

Use `--json` with a comma-separated list of field names to write JSON to stdout. This bypasses the TUI entirely.

```bash
gh member list my-org --json name,login,role,url
```

Available fields: `name`, `login`, `role`, `databaseId`, `url`.

Omitting field names returns all fields:

```bash
gh member list my-org --json
```

### Filter JSON output with jq

Pipe the JSON through a [jq](https://jqlang.org/) expression using `--jq` (or `-q`):

```bash
# Print only the logins of admins
gh member list my-org --json login,role --jq '.[] | select(.role == "ADMIN") | .login'
```

### Format output with a Go template

Use `--template` (or `-t`) to apply a [Go text/template](https://pkg.go.dev/text/template) to the JSON output:

```bash
gh member list my-org --json login,role \
  --template '{{range .}}{{.login}} ({{.role}}){{"\n"}}{{end}}'
```

### Use in a script or CI pipeline

When stdout is not a terminal (e.g. piped to another command), the TUI is suppressed automatically and tab-separated output is written instead. The `--org` argument is required in this mode.

```bash
# Pipe into grep
gh member list my-org | grep ADMIN

# Redirect to a file
gh member list my-org --limit -1 > members.tsv
```

To force structured JSON in CI:

```bash
gh member list my-org --json name,login,role,url > members.json
```

---

## Reference

### `gh member list [org]`

List the members of a GitHub organization.

```
gh member list [org] [flags]
```

#### Arguments

| Argument | Required | Description                                                                 |
|----------|----------|-----------------------------------------------------------------------------|
| `org`    | No       | Login name of the organization. If omitted in a terminal, an interactive selector is shown. Required in non-TTY and structured-output modes. |

#### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--limit` | `-L` | `30` | Maximum number of members to fetch. Use `-1` to fetch all members. |
| `--role` | | `all` | Filter by role. One of `all`, `admin`, `member`. |
| `--json` | | `""` | Output as JSON. Optionally specify a comma-separated list of fields (e.g. `name,login,role`). When set, the TUI is not shown and `org` is required. |
| `--jq` | `-q` | `""` | Apply a jq expression to the JSON output. Implies `--json`. |
| `--template` | `-t` | `""` | Apply a Go template to the JSON output. Implies `--json`. |

#### Output columns (table and TUI mode)

| Column  | Source field | Notes                                    |
|---------|--------------|------------------------------------------|
| NAME    | `name`       | Falls back to `login` when name is empty |
| ID      | `login`      | GitHub username                          |
| ROLE    | `role`       | `ADMIN` or `MEMBER` (uppercase)          |
| PROFILE | `url`        | `https://github.com/<login>`             |

#### JSON fields

| Field      | Type   | Description                          |
|------------|--------|--------------------------------------|
| `name`     | string | Display name; empty string if not set |
| `login`    | string | GitHub username                      |
| `role`     | string | `"ADMIN"` or `"MEMBER"`              |
| `databaseId` | int  | Numeric GitHub user ID               |
| `url`      | string | Profile URL                          |

#### TUI key bindings

| Key         | Action                                              |
|-------------|-----------------------------------------------------|
| `↑` / `↓`  | Move selection up/down                              |
| Type        | Filter rows by name or login (incremental search)   |
| `Backspace` | Delete last character from search                   |
| `Enter`     | Open selected member's profile URL in browser       |
| `q`         | Quit (only when search box is empty)                |
| `Esc`       | Quit                                                |
| `Ctrl+C`    | Quit                                                |

#### Exit codes

| Code | Meaning                                        |
|------|------------------------------------------------|
| `0`  | Success                                        |
| `1`  | Error (invalid flag value, API error, etc.)    |

---

## Explanation

### How the output mode is selected

`gh member list` chooses between three output modes automatically based on the flags and environment:

1. **Structured mode** — activated when any of `--json`, `--jq`, or `--template` is set. The TUI is not shown. `org` must be provided as an argument. Output goes to stdout and is suitable for scripting and CI.

2. **TUI mode** — activated when stdout is a terminal and no structured flag is set. If `org` is omitted, an interactive organization selector is shown first. The full member browser then opens with search and browser-open support.

3. **Static table mode** — activated when stdout is not a terminal (e.g. piped or redirected) and no structured flag is set. `org` must be provided. Output is tab-separated with a header row, matching the behavior of other `gh` commands when piped.

This mirrors the output-mode convention used throughout the GitHub CLI.

### Why GraphQL instead of REST

The REST endpoint `GET /orgs/{org}/members` does not include display names or roles in its response. Fetching them would require an additional `GET /users/{login}` call for names and `GET /orgs/{org}/memberships/{login}` for roles — creating N+1 API calls that exhaust the rate limit quickly for large organizations.

The GraphQL query `organization.membersWithRole` returns name, login, role, and profile URL together in a single request, fetching up to 100 members per page. With the default `--limit 30`, only one API call is made.

### Role filtering

GitHub's `membersWithRole` API does not support server-side role filtering. The `--role` flag is therefore applied client-side after each page is fetched. When `--role admin` or `--role member` is set, the page size sent to the API is always the maximum (100) regardless of the `--limit` value, to avoid making many small requests when admins are sparse in a large organization.
