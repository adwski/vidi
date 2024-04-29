package tool

import (
	"errors"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"net/url"
	"strings"
)

const (
	defaultEndpointURL = "http://localhost"
)

type (
	msgViDiURL string

	sConfig struct {
		form *huh.Form
		ep   *url.URL
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

func isURL(s string) error {
	u, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme == "" {
		return errors.New("invalid URL: missing scheme")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("invalid URL scheme: %s", u.Scheme)
	}
	if u.Hostname() == "" {
		return errors.New("invalid URL: missing hostname")
	}
	return nil
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
