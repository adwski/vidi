package tool

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sMainMenu struct {
	list   list.Model
	choice string
}

func newMainMenuScreen() *sMainMenu {
	items := []list.Item{
		item("Videos"),
		item("Upload"),
		item("Quotas"),
		item("Switch User"),
	}

	l := list.New(items, itemDelegate{}, defaultListWidth, defaultListHeight)
	l.Title = "Select action"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = styleListTitle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return &sMainMenu{list: l}
}

func (s *sMainMenu) init() tea.Cmd {
	return nil
}

func (s *sMainMenu) name() string {
	return "mainMenuScreen"
}

func (s *sMainMenu) update(msg tea.Msg) (tea.Cmd, *outerControl) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		s.list.SetWidth(m.Width)
		return nil, nil

	case tea.KeyMsg:
		switch keypress := m.String(); keypress {
		case "enter":
			i, ok := s.list.SelectedItem().(item)
			if ok {
				return nil, &outerControl{
					data: string(i),
				}
			}
		}
	}

	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return cmd, nil
}

func (s *sMainMenu) view() string {
	if s.choice != "" {
		return quitTextStyle.Render(fmt.Sprintf("Selected: %s", s.choice))
	}
	return lipgloss.JoinVertical(lipgloss.Left, styleTitle.Render("Vidi Terminal Menu"), s.list.View())
}
