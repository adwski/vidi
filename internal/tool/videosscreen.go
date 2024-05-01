package tool

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
			{Title: "Id", Width: 4},
			{Title: "Name", Width: 10},
			{Title: "Status", Width: 10},
			{Title: "Size", Width: 10},
			{Title: "CreatedAt", Width: 10},
		}
		rows []table.Row
	)

	for _, v := range videos {
		rows = append(rows, table.Row{v.ID, v.Name, v.Status, v.Size, v.CreatedAt})
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
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "space":
			if s.table.Focused() {
				s.table.Blur()
			} else {
				s.table.Focus()
			}
		case "enter":
			return nil, &outerControl{data: videosControl{vid: s.table.SelectedRow()[1]}}
		}
	}
	s.table, cmd = s.table.Update(msg)
	return cmd, nil
}

func (s *sVideos) view() string {
	return defaultHuhTheme.Form.Render(s.table.View())
}
