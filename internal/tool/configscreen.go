package tool

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

const (
	defaultEndpointURL = "http://localhost"
)

type (
	msgViDiURL string

	sConfig struct {
		form *huh.Form
	}
)

func newConfigScreen() *sConfig {
	return &sConfig{
		form: huh.NewForm(
			huh.NewGroup(
				huh.NewNote().Title("Seems like there's no valid config or I couldn't connect to ViDi endpoint"),
				huh.NewNote().Title("Configure ViDi endpoint URL"),
				huh.NewInput().
					Key("endpoint").
					Description("ViDi endpoint").
					Placeholder(defaultEndpointURL).
					Suggestions([]string{defaultEndpointURL}).
					Validate(isURL),
			),
		).WithTheme(defaultHuhTheme),
	}
}

func (s *sConfig) init() tea.Cmd {
	return s.form.Init()
}

func (s *sConfig) name() string {
	return "configScreen"
}

func (s *sConfig) update(msg tea.Msg) (tea.Cmd, *outerControl) {
	form, cmd := s.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		s.form = f
	}
	if s.form.State == huh.StateCompleted {
		ep := strings.TrimRight(s.form.GetString("endpoint"), "/")
		return nil, &outerControl{data: msgViDiURL(ep)}
	}
	return cmd, nil
}

func (s *sConfig) view() string {
	return s.form.View()
}
