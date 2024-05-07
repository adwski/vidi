//nolint:gomnd
package tool

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	sQuotas struct {
		table table.Model
	}

	quotasControl struct{}
)

func newQuotasScreen(quotas []QuotaParam) *sQuotas {
	var (
		columns = []table.Column{
			{Title: "Name", Width: 20},
			{Title: "Value", Width: 20},
		}
		rows = make([]table.Row, 0, len(quotas))
	)

	for _, p := range quotas {
		rows = append(rows, table.Row{p.Name, p.Value})
	}
	if len(quotas) == 0 {
		rows = append(rows, table.Row{"", "<cannot display quotas>"})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(defaultHuhTheme.Form.GetForeground()).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(defaultHuhTheme.Focused.FocusedButton.GetForeground()).
		Background(defaultHuhTheme.Focused.FocusedButton.GetBackground()).
		Bold(false)
	t.SetStyles(s)

	return &sQuotas{table: t}
}

func (s *sQuotas) init() tea.Cmd {
	return nil
}

func (s *sQuotas) name() string {
	return "quotaScreen"
}

func (s *sQuotas) update(msg tea.Msg) (tea.Cmd, *outerControl) {
	var cmd tea.Cmd
	if m, ok := msg.(tea.KeyMsg); ok {
		switch m.String() {
		case "space":
			if s.table.Focused() {
				s.table.Blur()
			} else {
				s.table.Focus()
			}
		case "backspace", "esc", "enter":
			return nil, &outerControl{data: quotasControl{}}
		}
	}
	s.table, cmd = s.table.Update(msg)
	return cmd, nil
}

func (s *sQuotas) view() string {
	return tableContainer.Render(s.table.View())
}
