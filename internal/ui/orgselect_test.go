package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tpu-kanglabs/gh-member/internal/gh"
)

// TestOrgSelectModel_EnterSelectsItem: pressing Enter with at least one item sets done=true and chosen.
func TestOrgSelectModel_EnterSelectsItem(t *testing.T) {
	orgs := []gh.Org{
		{Login: "my-org", Name: "My Organization"},
	}

	m := newOrgSelectModel(orgs)

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	final, ok := result.(orgSelectModel)
	if !ok {
		t.Fatal("Update did not return orgSelectModel")
	}

	if !final.done {
		t.Error("expected done=true after Enter")
	}
	if final.chosen != "my-org" {
		t.Errorf("expected chosen=%q, got %q", "my-org", final.chosen)
	}
}

// TestOrgSelectModel_QuitOnQ: pressing q sets done=true and leaves chosen empty (cancel).
func TestOrgSelectModel_QuitOnQ(t *testing.T) {
	orgs := []gh.Org{
		{Login: "my-org", Name: "My Organization"},
	}

	m := newOrgSelectModel(orgs)

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	final, ok := result.(orgSelectModel)
	if !ok {
		t.Fatal("Update did not return orgSelectModel")
	}

	if !final.done {
		t.Error("expected done=true after q")
	}
	if final.chosen != "" {
		t.Errorf("expected chosen to be empty (cancel), got %q", final.chosen)
	}
}

// TestOrgSelectModel_EmptyOrgs: SelectOrg returns an error when orgs is empty.
func TestOrgSelectModel_EmptyOrgs(t *testing.T) {
	_, err := SelectOrg([]gh.Org{})
	if err == nil {
		t.Fatal("expected error for empty orgs, got nil")
	}
	if err.Error() != "no organizations found" {
		t.Errorf("expected error %q, got %q", "no organizations found", err.Error())
	}
}
