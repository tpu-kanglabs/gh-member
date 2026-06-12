package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tpu-kanglabs/gh-member/internal/gh"
)

// TestFilterMembers_Empty: an empty query returns all members.
func TestFilterMembers_Empty(t *testing.T) {
	members := []gh.Member{
		{Name: "Alice", Login: "alice", Role: "MEMBER", URL: "https://github.com/alice"},
		{Name: "Bob", Login: "bob", Role: "ADMIN", URL: "https://github.com/bob"},
	}

	result := filterMembers(members, "")
	if len(result) != len(members) {
		t.Errorf("expected %d members, got %d", len(members), len(result))
	}
}

// TestFilterMembers_Match: only members whose Name or Login contains the query are returned.
func TestFilterMembers_Match(t *testing.T) {
	members := []gh.Member{
		{Name: "Alice Smith", Login: "asmith", Role: "MEMBER", URL: "https://github.com/asmith"},
		{Name: "Bob Jones", Login: "bjones", Role: "ADMIN", URL: "https://github.com/bjones"},
		{Name: "", Login: "charlie", Role: "MEMBER", URL: "https://github.com/charlie"},
	}

	t.Run("match by Name", func(t *testing.T) {
		result := filterMembers(members, "Alice")
		if len(result) != 1 {
			t.Errorf("expected 1 result, got %d", len(result))
		}
		if result[0].Login != "asmith" {
			t.Errorf("expected login %q, got %q", "asmith", result[0].Login)
		}
	})

	t.Run("match by Login", func(t *testing.T) {
		result := filterMembers(members, "bjones")
		if len(result) != 1 {
			t.Errorf("expected 1 result, got %d", len(result))
		}
		if result[0].Login != "bjones" {
			t.Errorf("expected login %q, got %q", "bjones", result[0].Login)
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		result := filterMembers(members, "ALICE")
		if len(result) != 1 {
			t.Errorf("expected 1 result, got %d", len(result))
		}
	})

	t.Run("no match", func(t *testing.T) {
		result := filterMembers(members, "xyz-no-match")
		if len(result) != 0 {
			t.Errorf("expected 0 results, got %d", len(result))
		}
	})

	t.Run("match in Login when Name is empty", func(t *testing.T) {
		result := filterMembers(members, "charlie")
		if len(result) != 1 {
			t.Errorf("expected 1 result, got %d", len(result))
		}
		if result[0].Login != "charlie" {
			t.Errorf("expected login %q, got %q", "charlie", result[0].Login)
		}
	})
}

// TestMembersToRows_EmptyName: when Name is empty, the NAME column shows Login instead.
func TestMembersToRows_EmptyName(t *testing.T) {
	members := []gh.Member{
		{Name: "", Login: "noname", Role: "MEMBER", URL: "https://github.com/noname"},
		{Name: "HasName", Login: "hasname", Role: "ADMIN", URL: "https://github.com/hasname"},
	}

	rows := membersToRows(members)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	// Empty Name: Login is shown in the NAME column.
	if rows[0][0] != "noname" {
		t.Errorf("expected NAME column to be %q, got %q", "noname", rows[0][0])
	}

	// Non-empty Name: Name is shown.
	if rows[1][0] != "HasName" {
		t.Errorf("expected NAME column to be %q, got %q", "HasName", rows[1][0])
	}
}

// TestBuildColumns_ProfileWidth: verifies PROFILE column width calculation in buildColumns.
func TestBuildColumns_ProfileWidth(t *testing.T) {
	// windowWidth=0 → default width (20)
	cols := buildColumns(0)
	if cols[3].Width != 20 {
		t.Errorf("windowWidth=0: expected PROFILE width 20, got %d", cols[3].Width)
	}

	// windowWidth=52 exactly (name20+id20+role8+padding4=52) → default width (20)
	cols = buildColumns(52)
	if cols[3].Width != 20 {
		t.Errorf("windowWidth=52: expected PROFILE width 20, got %d", cols[3].Width)
	}

	// windowWidth=100 → 100-20-20-8-4=48
	cols = buildColumns(100)
	if cols[3].Width != 48 {
		t.Errorf("windowWidth=100: expected PROFILE width 48, got %d", cols[3].Width)
	}
}

// TestMemberTableModel_SearchFiltersRows: typing in the search box filters the table rows.
func TestMemberTableModel_SearchFiltersRows(t *testing.T) {
	members := []gh.Member{
		{Name: "Alice", Login: "alice", Role: "MEMBER", URL: "https://github.com/alice"},
		{Name: "Bob", Login: "bob", Role: "ADMIN", URL: "https://github.com/bob"},
	}
	m := newMemberTableModel(members)

	// Initially all rows are shown.
	if len(m.table.Rows()) != 2 {
		t.Fatalf("initial: expected 2 rows, got %d", len(m.table.Rows()))
	}

	// Typing 'a' matches "Alice" but not "Bob".
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m2, ok := updated.(memberTableModel)
	if !ok {
		t.Fatal("Update did not return memberTableModel")
	}
	if len(m2.table.Rows()) != 1 {
		t.Errorf("after typing 'a': expected 1 row, got %d", len(m2.table.Rows()))
	}
}

// TestMemberTableModel_QuitOnQ: pressing q when the search box is empty returns a Quit command.
func TestMemberTableModel_QuitOnQ(t *testing.T) {
	members := []gh.Member{
		{Name: "Alice", Login: "alice", Role: "MEMBER", URL: "https://github.com/alice"},
	}

	m := newMemberTableModel(members)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatal("expected a command to be returned for q key")
	}

	// Confirm the returned command is tea.Quit by inspecting its message type.
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}
