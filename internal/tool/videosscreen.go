//nolint:gomnd
package tool

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strconv"
)

type (
	sVideos struct {
		table table.Model
	}

	videosControl struct {
		vid string
	}
)

func newVideosScreen(videos []Video) *sVideos {
	var (
		columns = []table.Column{
			{Title: "N", Width: 2},
			{Title: "Name", Width: 30},
			{Title: "Status", Width: 10},
			{Title: "Size", Width: 10},
			{Title: "CreatedAt", Width: 25},
		}
		rows = make([]table.Row, 0, len(videos))
	)

	for i, v := range videos {
		rows = append(rows, table.Row{strconv.Itoa(i + 1), v.Name, v.Status, v.Size, v.CreatedAt})
	}
	if len(videos) == 0 {
		rows = append(rows, table.Row{"", "<no videos to show>", "", "", ""})
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

	return &sVideos{table: t}
}

func (s *sVideos) init() tea.Cmd {
	return nil
}

func (s *sVideos) name() string {
	return "videosScreen"
}

func (s *sVideos) update(msg tea.Msg) (tea.Cmd, *outerControl) {
	var cmd tea.Cmd
	if m, ok := msg.(tea.KeyMsg); ok {
		switch m.String() {
		case "space":
			if s.table.Focused() {
				s.table.Blur()
			} else {
				s.table.Focus()
			}
		case "backspace", "esc":
			return nil, &outerControl{data: videosControl{vid: ""}}
		case "enter":
			return nil, &outerControl{data: videosControl{vid: s.table.SelectedRow()[1]}}
		}
	}
	s.table, cmd = s.table.Update(msg)
	return cmd, nil
}

func (s *sVideos) view() string {
	return containerWithBorder.Render(s.table.View())
}
