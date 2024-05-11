//nolint:mnd // table sizes
package tool

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	sVideos struct { //nolint:govet // embedded structs are not aligned optimally
		err        error
		videos     [][2]string
		help       *help.Model
		watchURL   string
		table      table.Model
		keys       keyMap
		videoToDel int
	}

	videosControl struct {
		vid    string
		delete bool
		watch  bool
	}

	watchInfo struct {
		url string
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
		videoIDs = make([][2]string, 0, len(videos))
		rows     = make([]table.Row, 0, len(videos))
	)

	for i, v := range videos {
		rows = append(rows, table.Row{strconv.Itoa(i + 1), v.Name, v.Status, v.Size, v.CreatedAt})
		videoIDs = append(videoIDs, [2]string{v.ID, v.Name})
	}
	if len(videos) == 0 {
		rows = append(rows, table.Row{"", "<no videos to show>", "", "", ""})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
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

	km := keyMap{
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "move down"),
		),
		Del: key.NewBinding(
			key.WithKeys("d", "D"),
			key.WithHelp("D/d", "delete video"),
		),
		Watch: key.NewBinding(
			key.WithKeys("w", "W"),
			key.WithHelp("W/w", "get watch url"),
		),
		Return: key.NewBinding(
			key.WithKeys("←", "backspace", "esc"),
			key.WithHelp("←/esc/backspace", "go back"),
		),
	}
	km.kList = []key.Binding{km.Up, km.Down, km.Del, km.Watch, km.Return}

	return &sVideos{
		keys:       km,
		videos:     videoIDs,
		videoToDel: -1,
		table:      t,
		help: &help.Model{
			Width:          0,
			ShowAll:        false,
			ShortSeparator: " • ",
			FullSeparator:  " • ",
			Ellipsis:       " * ",
			Styles:         defaultHuhTheme.Help,
		},
	}
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
	case watchInfo:
		s.watchURL = m.url
	case error:
		s.err = m
	case tea.KeyMsg:
		if s.videoToDel > -1 {
			switch m.String() {
			case "Y", "y":
				return nil, &outerControl{data: videosControl{vid: s.videos[s.videoToDel][0], delete: true}}
			default:
				s.videoToDel = -1
			}
		}

		switch m.String() {
		case "space":
			if s.table.Focused() {
				s.table.Blur()
			} else {
				s.table.Focus()
			}
		case "d":
			// this is ugly, but I didn't find nicer way. Also, it'll always be a number.
			vNum, _ := strconv.Atoi(s.table.SelectedRow()[0])
			s.videoToDel = vNum - 1
			return nil, nil // prevent table updates
		case "w":
			vNum, _ := strconv.Atoi(s.table.SelectedRow()[0])
			return nil, &outerControl{data: videosControl{vid: s.videos[vNum-1][0], watch: true}}
		case "backspace", "esc", "left":
			return nil, &outerControl{data: videosControl{vid: ""}}
		case "enter":
			return nil, &outerControl{data: videosControl{vid: s.table.SelectedRow()[1]}}
		}
	}
	s.table, cmd = s.table.Update(msg)
	return cmd, nil
}

func (s *sVideos) view() string {
	var footer = "\n"
	switch {
	case s.videoToDel > -1:
		footer = confirmStyle.Render(fmt.Sprintf(">> Delete video %d: '%s'? Press [Y]es or any key to cancel\n",
			s.videoToDel+1, s.videos[s.videoToDel][1]))
	case len(s.watchURL) > 0:
		footer = confirmStyle.Render(fmt.Sprintf(">> Watch URL: %s\n", s.watchURL))
	case s.err != nil:
		footer = errStyleFooter.Render(fmt.Sprintf(">> Error: %s\n", s.err.Error()))
	}
	return containerWithBorder.Render(s.table.View()) + "\n" + footer + "\n\n" + s.help.View(s.keys)
}
