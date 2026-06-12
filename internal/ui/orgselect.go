package ui

import (
	"errors"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/rmuraix/gh-member/internal/gh"
)

// orgItem implements list.Item for a GitHub Organization.
type orgItem struct {
	login string
	name  string
}

// Title shows the human-readable name first (more recognizable); falls back to login.
func (i orgItem) Title() string {
	if i.name != "" {
		return i.name
	}
	return i.login
}

// Description shows the login handle below the name.
func (i orgItem) Description() string { return i.login }
func (i orgItem) FilterValue() string { return i.login }

// orgSelectModel is the Bubble Tea model for the org selector.
type orgSelectModel struct {
	list   list.Model
	chosen string
	done   bool
}

func newOrgSelectModel(orgs []gh.Org) orgSelectModel {
	items := make([]list.Item, len(orgs))
	for i, org := range orgs {
		items[i] = orgItem{login: org.Login, name: org.Name}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select organization"

	return orgSelectModel{list: l}
}

func (m orgSelectModel) Init() tea.Cmd {
	return nil
}

func (m orgSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if item, ok := m.list.SelectedItem().(orgItem); ok {
				m.chosen = item.login
				m.done = true
				return m, tea.Quit
			}
		case "q", "esc", "ctrl+c":
			m.done = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height-2)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m orgSelectModel) View() string {
	return m.list.View()
}

// SelectOrg shows an interactive list of orgs and returns the login of the chosen one.
// Returns an empty string and nil when the user cancels with q/Ctrl+C/Esc.
// Returns an error if orgs is empty.
func SelectOrg(orgs []gh.Org) (string, error) {
	if len(orgs) == 0 {
		return "", errors.New("no organizations found")
	}

	m := newOrgSelectModel(orgs)
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return "", err
	}

	final, ok := result.(orgSelectModel)
	if !ok {
		return "", nil
	}

	return final.chosen, nil
}
