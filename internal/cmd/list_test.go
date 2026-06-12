package cmd

import (
	"strings"
	"testing"
)

// TestListCmd_FlagParsing verifies that --limit, --role, and --json flags
// are correctly registered and have the expected default values.
func TestListCmd_FlagParsing(t *testing.T) {
	cmd := newListCmd()

	// --limit / -L: default 30
	limitFlag := cmd.Flags().Lookup("limit")
	if limitFlag == nil {
		t.Fatal("flag --limit not found")
	}
	if limitFlag.DefValue != "30" {
		t.Errorf("--limit default = %q, want %q", limitFlag.DefValue, "30")
	}

	// --role: default "all"
	roleFlag := cmd.Flags().Lookup("role")
	if roleFlag == nil {
		t.Fatal("flag --role not found")
	}
	if roleFlag.DefValue != "all" {
		t.Errorf("--role default = %q, want %q", roleFlag.DefValue, "all")
	}

	// --json: default ""
	jsonFlag := cmd.Flags().Lookup("json")
	if jsonFlag == nil {
		t.Fatal("flag --json not found")
	}
	if jsonFlag.DefValue != "" {
		t.Errorf("--json default = %q, want %q", jsonFlag.DefValue, "")
	}

	// --jq / -q: default ""
	jqFlag := cmd.Flags().Lookup("jq")
	if jqFlag == nil {
		t.Fatal("flag --jq not found")
	}
	if jqFlag.DefValue != "" {
		t.Errorf("--jq default = %q, want %q", jqFlag.DefValue, "")
	}

	// --template / -t: default ""
	templateFlag := cmd.Flags().Lookup("template")
	if templateFlag == nil {
		t.Fatal("flag --template not found")
	}
	if templateFlag.DefValue != "" {
		t.Errorf("--template default = %q, want %q", templateFlag.DefValue, "")
	}
}

// TestListCmd_FlagShorthands verifies that shorthand flags are registered correctly.
func TestListCmd_FlagShorthands(t *testing.T) {
	cmd := newListCmd()

	// -L for --limit
	if f := cmd.Flags().ShorthandLookup("L"); f == nil {
		t.Error("shorthand -L not found for --limit")
	}

	// -q for --jq
	if f := cmd.Flags().ShorthandLookup("q"); f == nil {
		t.Error("shorthand -q not found for --jq")
	}

	// -t for --template
	if f := cmd.Flags().ShorthandLookup("t"); f == nil {
		t.Error("shorthand -t not found for --template")
	}
}

// TestListCmd_ParseLimit verifies that --limit flag value is parsed correctly.
func TestListCmd_ParseLimit(t *testing.T) {
	cmd := newListCmd()

	if err := cmd.Flags().Set("limit", "50"); err != nil {
		t.Fatalf("failed to set --limit: %v", err)
	}

	limitFlag := cmd.Flags().Lookup("limit")
	if limitFlag.Value.String() != "50" {
		t.Errorf("--limit value = %q, want %q", limitFlag.Value.String(), "50")
	}
}

// TestListCmd_ParseRole verifies that --role flag value is parsed correctly.
func TestListCmd_ParseRole(t *testing.T) {
	cmd := newListCmd()

	for _, role := range []string{"all", "admin", "member"} {
		if err := cmd.Flags().Set("role", role); err != nil {
			t.Fatalf("failed to set --role=%s: %v", role, err)
		}
		if got := cmd.Flags().Lookup("role").Value.String(); got != role {
			t.Errorf("--role value = %q, want %q", got, role)
		}
	}
}

// TestListCmd_ParseJSON verifies that --json flag value is parsed correctly.
func TestListCmd_ParseJSON(t *testing.T) {
	cmd := newListCmd()

	if err := cmd.Flags().Set("json", "login,role,url"); err != nil {
		t.Fatalf("failed to set --json: %v", err)
	}

	if got := cmd.Flags().Lookup("json").Value.String(); got != "login,role,url" {
		t.Errorf("--json value = %q, want %q", got, "login,role,url")
	}
}

// --- RunE integration tests ---

// TestListCmd_InvalidRole verifies that an invalid --role value returns an error
// without making any API call.
func TestListCmd_InvalidRole(t *testing.T) {
	cmd := newListCmd()
	cmd.SetArgs([]string{"--role", "superuser", "myorg"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --role, got nil")
	}
	if !strings.Contains(err.Error(), "invalid role") {
		t.Errorf("expected 'invalid role' in error message, got: %v", err)
	}
}

// TestListCmd_OrgRequiredStructuredMode verifies that --json without org returns an error.
func TestListCmd_OrgRequiredStructuredMode(t *testing.T) {
	cmd := newListCmd()
	cmd.SetArgs([]string{"--json", "login,role"}) // no org

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when org is missing in structured mode, got nil")
	}
	if !strings.Contains(err.Error(), "organization is required") {
		t.Errorf("expected 'organization is required' in error message, got: %v", err)
	}
}
