package tool

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"strings"
)

const (
	newUserOptionLogin = iota
	newUserOptionRegister
)

type userControl struct {
	username string
	password string
	option   int
}

func (uc userControl) String() string {
	return fmt.Sprintf("username: %s, password: %s, option: %d",
		uc.username, strings.Repeat("*", len(uc.password)), uc.option)
}

type sNewUser struct {
	form     *huh.Form
	username string
	password string
	option   int
}

func newUserScreen() *sNewUser {
	var cfg sNewUser
	f := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title("No locally stored users have found"),
			huh.NewSelect[int]().
				Title("Choose what to do").
				Options(
					huh.NewOption("Login with existing account", newUserOptionLogin),
					huh.NewOption("Register", newUserOptionRegister),
				).Value(&cfg.option),
		),
		huh.NewGroup(
			huh.NewNote().Title("Provide user credentials"),
			huh.NewInput().Title("Enter username").Key("username").
				Suggestions([]string{"username"}).Validate(validateUsername),
			huh.NewInput().Title("Enter password").Key("password").
				Suggestions([]string{"password"}).Validate(validatePassword).Password(true),
			huh.NewConfirm().Title("Proceed with entered credentials"),
		),
	).WithTheme(defaultHuhTheme)
	cfg.form = f
	return &cfg
}

func validateUsername(s string) error {
	if len(s) < 3 {
		return fmt.Errorf("username should not be less than 3 letters")
	}
	return nil
}

func validatePassword(s string) error {
	if len(s) < 8 {
		return fmt.Errorf("password should not be less than 8 letters")
	}
	return nil
}

func (s *sNewUser) init() tea.Cmd {
	return s.form.Init()
}

func (s *sNewUser) name() string {
	return "newUserScreen"
}

func (s *sNewUser) update(msg tea.Msg) (tea.Cmd, *outerControl) {
	form, cmd := s.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		s.form = f
	}
	if s.form.State == huh.StateCompleted {
		return nil, &outerControl{data: userControl{
			username: s.form.GetString("username"),
			password: s.form.GetString("password"),
			option:   s.option,
		}}
	}
	return cmd, nil
}

func (s *sNewUser) view() string {
	return s.form.View()
}
