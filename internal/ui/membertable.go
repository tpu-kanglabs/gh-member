package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cli/go-gh/v2/pkg/browser"

	"github.com/rmuraix/gh-member/internal/gh"
)

var (
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)
	tableStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))
)

// memberTableModel is the Bubble Tea model for the member table browser.
type memberTableModel struct {
	allMembers  []gh.Member
	table       table.Model
	search      textinput.Model
	windowWidth int
}

func newMemberTableModel(members []gh.Member) memberTableModel {
	ti := textinput.New()
	ti.Placeholder = "Search by name or login..."
	ti.Focus()

	cols := buildColumns(0)
	rows := membersToRows(members)

	s := table.DefaultStyles()
	s.Selected = selectedStyle

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(20),
	)
	t.SetStyles(s)

	return memberTableModel{
		allMembers: members,
		table:      t,
		search:     ti,
	}
}

func buildColumns(windowWidth int) []table.Column {
	const (
		nameWidth = 20
		idWidth   = 20
		roleWidth = 8
		padding   = 4
	)
	profileWidth := 20
	if windowWidth > nameWidth+idWidth+roleWidth+padding {
		profileWidth = windowWidth - nameWidth - idWidth - roleWidth - padding
	}
	return []table.Column{
		{Title: "NAME", Width: nameWidth},
		{Title: "ID", Width: idWidth},
		{Title: "ROLE", Width: roleWidth},
		{Title: "PROFILE", Width: profileWidth},
	}
}

func (m memberTableModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m memberTableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.table.SetColumns(buildColumns(m.windowWidth))
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			row := m.table.SelectedRow()
			if len(row) >= 4 {
				url := row[3]
				b := browser.New("", os.Stdout, os.Stderr)
				if err := b.Browse(url); err != nil {
					fmt.Fprintf(os.Stderr, "error opening browser: %v\n", err)
				}
			}
			return m, nil

		case "ctrl+c", "esc":
			return m, tea.Quit
		case "q":
			// Quit only when the search box is empty; otherwise pass to textinput.
			if m.search.Value() == "" {
				return m, tea.Quit
			}

		default:
			prevQuery := m.search.Value()
			var tiCmd tea.Cmd
			m.search, tiCmd = m.search.Update(msg)
			cmds = append(cmds, tiCmd)

			if m.search.Value() != prevQuery {
				filtered := filterMembers(m.allMembers, m.search.Value())
				m.table.SetRows(membersToRows(filtered))
			}

			var tblCmd tea.Cmd
			m.table, tblCmd = m.table.Update(msg)
			cmds = append(cmds, tblCmd)

			return m, tea.Batch(cmds...)
		}
	}

	// Pass through other messages to table (e.g. up/down navigation)
	var tblCmd tea.Cmd
	m.table, tblCmd = m.table.Update(msg)
	cmds = append(cmds, tblCmd)

	return m, tea.Batch(cmds...)
}

func (m memberTableModel) View() string {
	footer := " ↑/↓: move  enter: open profile  q: quit (when search empty)  esc: quit"
	return m.search.View() + "\n" + tableStyle.Render(m.table.View()) + "\n" + footer
}

// filterMembers returns members whose Name or Login contains query (case-insensitive).
// Returns all members when query is empty.
func filterMembers(members []gh.Member, query string) []gh.Member {
	if query == "" {
		return members
	}
	q := strings.ToLower(query)
	result := make([]gh.Member, 0)
	for _, m := range members {
		if strings.Contains(strings.ToLower(m.Name), q) ||
			strings.Contains(strings.ToLower(m.Login), q) {
			result = append(result, m)
		}
	}
	return result
}

// membersToRows converts a []gh.Member slice to []table.Row.
func membersToRows(members []gh.Member) []table.Row {
	rows := make([]table.Row, 0, len(members))
	for _, m := range members {
		name := m.Name
		if name == "" {
			name = m.Login
		}
		rows = append(rows, table.Row{name, m.Login, m.Role, m.URL})
	}
	return rows
}

// BrowseMembers displays the member list in an interactive Bubble Tea table.
// Type to filter by Name or Login; Enter opens the selected profile URL in a browser; q/Esc/Ctrl+C quits.
func BrowseMembers(members []gh.Member) error {
	m := newMemberTableModel(members)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
