package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rmuraix/gh-member/internal/gh"
)

// TestOrgSelectModel_EnterSelectsItem: list が 1 アイテム以上あるとき Enter で done になること。
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

// TestOrgSelectModel_QuitOnQ: q キーで done になりキャンセルになること。
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

// TestOrgSelectModel_EmptyOrgs: orgs が空のとき SelectOrg がエラーを返すこと。
func TestOrgSelectModel_EmptyOrgs(t *testing.T) {
	_, err := SelectOrg([]gh.Org{})
	if err == nil {
		t.Fatal("expected error for empty orgs, got nil")
	}
	if err.Error() != "no organizations found" {
		t.Errorf("expected error %q, got %q", "no organizations found", err.Error())
	}
}
