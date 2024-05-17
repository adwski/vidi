package tool

import (
	"errors"

	tea "github.com/charmbracelet/bubbletea"
)

type (
	msgErrorScreenDone int

	sErr struct {
		err error
	}
)

func newErrorScreen(err error) *sErr {
	return &sErr{err: err}
}

func (s *sErr) init() tea.Cmd {
	return nil
}

func (s *sErr) name() string {
	return "errorScreen"
}

func (s *sErr) update(msg tea.Msg) (tea.Cmd, *outerControl) {
	if _, ok := msg.(tea.KeyMsg); ok {
		return nil, &outerControl{data: msgErrorScreenDone(1)}
	}
	return nil, nil
}

func (s *sErr) view() string {
	if s.err == nil {
		s.err = errors.New("<no error>")
	}
	return errStyleBorder.Render(s.err.Error() + "\n\npress any key to continue...")
}
