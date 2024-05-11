package tool

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

type (
	reLogControl struct {
		password string
	}

	sReLog struct {
		form *huh.Form
	}
)

func newReLogScreen(username string) *sReLog {
	var cfg sReLog
	f := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title(fmt.Sprintf("Enter password for '%s'", username)),
			huh.NewInput().Title("Enter password").Key("password").
				Suggestions([]string{"password"}).Validate(validatePassword).Password(true),
			huh.NewConfirm().Title("Proceed with entered password"),
		),
	).WithTheme(defaultHuhTheme)
	cfg.form = f
	return &cfg
}

func (s *sReLog) init() tea.Cmd {
	return s.form.Init()
}

func (s *sReLog) name() string {
	return "reLogScreen"
}

func (s *sReLog) update(msg tea.Msg) (tea.Cmd, *outerControl) {
	form, cmd := s.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		s.form = f
	}
	if s.form.State == huh.StateCompleted {
		return nil, &outerControl{data: reLogControl{
			password: s.form.GetString("password"),
		}}
	}
	return cmd, nil
}

func (s *sReLog) view() string {
	return s.form.View()
}
