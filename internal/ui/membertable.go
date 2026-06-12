package ui

import (
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

// memberTableModel は Bubble Tea のモデル。
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
			if row != nil && len(row) >= 4 {
				url := row[3]
				b := browser.New("", os.Stdout, os.Stderr)
				_ = b.Browse(url)
			}
			return m, nil

		case "ctrl+c", "q", "esc":
			return m, tea.Quit

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
	footer := " ↑/↓: move  enter: open profile  q: quit"
	return m.search.View() + "\n" + tableStyle.Render(m.table.View()) + "\n" + footer
}

// filterMembers は query に部分一致する members を返す（大文字小文字無視）。
// Name または Login のどちらかに query が含まれれば一致。
// query が空なら全件返す。
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

// membersToRows は []gh.Member を []table.Row に変換する。
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

// BrowseMembers はメンバー一覧を Bubble Tea テーブルで表示する。
// - textinput で上部に検索ボックスを表示し、タイプで絞り込み（Name または Login に部分一致）
// - テーブル列: NAME | ID | ROLE | PROFILE (URL)
// - Enter キーで選択された行のプロフィール URL をブラウザで開く
// - q / Esc / Ctrl+C で終了
// - ウィンドウサイズに応じてテーブル幅が変わる
func BrowseMembers(members []gh.Member) error {
	m := newMemberTableModel(members)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
