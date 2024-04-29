package tool

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"strconv"
)

const (
	optionUserSelectCurrent = iota - 2
	optionUserLogInNew
)

type (
	userSelectControl struct {
		option int
	}

	sUserSelect struct {
		form     *huh.Form
		password string
		option   int
	}
)

func (usc userSelectControl) String() string {
	switch usc.option {
	case optionUserSelectCurrent:
		return "option current user"
	case optionUserLogInNew:
		return "option new user"
	default:
		return "login as another: " + strconv.Itoa(usc.option)
	}
}

func newUserSelect(users []User, id int) *sUserSelect {
	var (
		us   sUserSelect
		opts []huh.Option[int]
	)
	opts = append(opts, huh.NewOption(fmt.Sprintf("Enter password for '%s'", users[id].Name), optionUserSelectCurrent))
	opts = append(opts, huh.NewOption("Select another user", optionUserLogInNew))
	for idx, u := range users {
		if idx != id {
			opts = append(opts, huh.NewOption(fmt.Sprintf("Login as '%s'", u.Name), id))
		}
	}
	f := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title("Current user has no valid token. "+
				"Enter your password again or select another user"),
			huh.NewSelect[int]().
				Title("Choose what to do").
				Options(opts...).Value(&us.option),
		)).WithTheme(defaultHuhTheme)
	us.form = f
	return &us
}

func (s *sUserSelect) init() tea.Cmd {
	return s.form.Init()
}

func (s *sUserSelect) name() string {
	return "userSelectScreen"
}

func (s *sUserSelect) update(msg tea.Msg) (tea.Cmd, *outerControl) {
	form, cmd := s.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		s.form = f
	}
	if s.form.State == huh.StateCompleted {
		return nil, &outerControl{data: userSelectControl{
			option: s.option,
		}}
	}
	return cmd, nil
}

func (s *sUserSelect) view() string {
	return s.form.View()
}
