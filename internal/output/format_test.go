package output_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/tpu-kanglabs/gh-member/internal/gh"
	"github.com/tpu-kanglabs/gh-member/internal/output"
)

// testMembers is a shared set of test members.
var testMembers = []gh.Member{
	{Name: "Alice Smith", Login: "alice", Role: "ADMIN", DatabaseID: 1, URL: "https://github.com/alice"},
	{Name: "", Login: "bob", Role: "MEMBER", DatabaseID: 2, URL: "https://github.com/bob"},
}

// --- PrintTable tests ---

// TestPrintTable_TTY_ContainsHeaders checks that TTY output includes column headers.
func TestPrintTable_TTY_ContainsHeaders(t *testing.T) {
	var buf bytes.Buffer
	if err := output.PrintTable(&buf, true, 120, testMembers); err != nil {
		t.Fatalf("PrintTable returned error: %v", err)
	}
	got := buf.String()
	for _, header := range []string{"NAME", "ID", "ROLE", "PROFILE"} {
		if !strings.Contains(got, header) {
			t.Errorf("expected header %q in output, got:\n%s", header, got)
		}
	}
}

// TestPrintTable_TTY_EmptyNameFallsBackToLogin checks that when Name is empty, Login is shown.
func TestPrintTable_TTY_EmptyNameFallsBackToLogin(t *testing.T) {
	var buf bytes.Buffer
	if err := output.PrintTable(&buf, true, 120, testMembers); err != nil {
		t.Fatalf("PrintTable returned error: %v", err)
	}
	got := buf.String()
	// bob has no Name, so "bob" (Login) must appear
	if !strings.Contains(got, "bob") {
		t.Errorf("expected login %q to appear when Name is empty, got:\n%s", "bob", got)
	}
	// alice has a Name, so "Alice Smith" should appear
	if !strings.Contains(got, "Alice Smith") {
		t.Errorf("expected name %q to appear, got:\n%s", "Alice Smith", got)
	}
}

// TestPrintTable_NonTTY_TSV checks that non-TTY output is tab-separated.
func TestPrintTable_NonTTY_TSV(t *testing.T) {
	var buf bytes.Buffer
	if err := output.PrintTable(&buf, false, 0, testMembers); err != nil {
		t.Fatalf("PrintTable returned error: %v", err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) == 0 {
		t.Fatal("expected at least one line of output")
	}
	for i, line := range lines {
		if !strings.Contains(line, "\t") {
			t.Errorf("line %d is not tab-separated: %q", i, line)
		}
	}
}

// TestPrintTable_NonTTY_NoHeaders checks that non-TTY output has no headers.
func TestPrintTable_NonTTY_NoHeaders(t *testing.T) {
	var buf bytes.Buffer
	if err := output.PrintTable(&buf, false, 0, testMembers); err != nil {
		t.Fatalf("PrintTable returned error: %v", err)
	}
	got := buf.String()
	// tsvTablePrinter ignores AddHeader, so no header line should appear.
	if strings.Contains(got, "NAME") || strings.Contains(got, "ROLE") {
		t.Errorf("non-TTY output should not contain headers, got:\n%s", got)
	}
}

// --- PrintJSON tests ---

// TestPrintJSON_AllFields checks that all fields are present when fields is nil.
func TestPrintJSON_AllFields(t *testing.T) {
	var buf bytes.Buffer
	if err := output.PrintJSON(&buf, testMembers, nil); err != nil {
		t.Fatalf("PrintJSON returned error: %v", err)
	}
	var got []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	if len(got) != len(testMembers) {
		t.Fatalf("expected %d items, got %d", len(testMembers), len(got))
	}
	for _, expectedField := range []string{"name", "login", "role", "databaseId", "url"} {
		if _, ok := got[0][expectedField]; !ok {
			t.Errorf("expected field %q in JSON, not found in: %v", expectedField, got[0])
		}
	}
}

// TestPrintJSON_SelectedFields checks that only specified fields are present when fields is provided.
func TestPrintJSON_SelectedFields(t *testing.T) {
	var buf bytes.Buffer
	if err := output.PrintJSON(&buf, testMembers, []string{"name", "login"}); err != nil {
		t.Fatalf("PrintJSON returned error: %v", err)
	}
	var got []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	if len(got) != len(testMembers) {
		t.Fatalf("expected %d items, got %d", len(testMembers), len(got))
	}
	item := got[0]
	if _, ok := item["name"]; !ok {
		t.Errorf("expected field %q in JSON", "name")
	}
	if _, ok := item["login"]; !ok {
		t.Errorf("expected field %q in JSON", "login")
	}
	// Fields NOT requested must not be present.
	for _, unexpected := range []string{"role", "databaseId", "url"} {
		if _, ok := item[unexpected]; ok {
			t.Errorf("unexpected field %q found in JSON output", unexpected)
		}
	}
}

// TestPrintJSON_EmptySliceAllFields checks that an empty (non-nil) fields slice produces all fields.
func TestPrintJSON_EmptySliceAllFields(t *testing.T) {
	var buf bytes.Buffer
	if err := output.PrintJSON(&buf, testMembers[:1], []string{}); err != nil {
		t.Fatalf("PrintJSON returned error: %v", err)
	}
	var got []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	for _, expectedField := range []string{"name", "login", "role", "databaseId", "url"} {
		if _, ok := got[0][expectedField]; !ok {
			t.Errorf("empty fields slice: expected field %q in JSON, not found in: %v", expectedField, got[0])
		}
	}
}

// --- ApplyJQ tests ---

// TestApplyJQ_LoginFilter checks that a simple jq filter works.
func TestApplyJQ_LoginFilter(t *testing.T) {
	jsonInput := `[{"login":"alice"},{"login":"bob"}]`
	var buf bytes.Buffer
	if err := output.ApplyJQ(&buf, strings.NewReader(jsonInput), ".[] | .login"); err != nil {
		t.Fatalf("ApplyJQ returned error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "alice") {
		t.Errorf("expected %q in output, got: %q", "alice", got)
	}
	if !strings.Contains(got, "bob") {
		t.Errorf("expected %q in output, got: %q", "bob", got)
	}
}

// TestApplyJQ_InvalidFilter checks that an invalid jq filter returns an error.
func TestApplyJQ_InvalidFilter(t *testing.T) {
	jsonInput := `[{"login":"alice"}]`
	var buf bytes.Buffer
	err := output.ApplyJQ(&buf, strings.NewReader(jsonInput), "this is not valid jq |||")
	if err == nil {
		t.Error("expected error for invalid jq filter, got nil")
	}
}

// --- ApplyTemplate tests ---

// TestApplyTemplate_LoginRange checks that a Go template iterating over items works.
func TestApplyTemplate_LoginRange(t *testing.T) {
	jsonInput := `[{"login":"alice"},{"login":"bob"}]`
	var buf bytes.Buffer
	if err := output.ApplyTemplate(&buf, strings.NewReader(jsonInput), `{{range .}}{{.login}}{{end}}`); err != nil {
		t.Fatalf("ApplyTemplate returned error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "alice") {
		t.Errorf("expected %q in output, got: %q", "alice", got)
	}
	if !strings.Contains(got, "bob") {
		t.Errorf("expected %q in output, got: %q", "bob", got)
	}
}

// TestApplyTemplate_InvalidSyntax checks that an invalid template syntax returns an error.
func TestApplyTemplate_InvalidSyntax(t *testing.T) {
	jsonInput := `[{"login":"alice"}]`
	var buf bytes.Buffer
	err := output.ApplyTemplate(&buf, strings.NewReader(jsonInput), `{{range .}{{.login}}{{end}}`)
	if err == nil {
		t.Error("expected error for invalid template syntax, got nil")
	}
}
