package tool

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"io"
	"strings"
)

const (
	cursor            = ">>"
	defaultListHeight = 14
	defaultListWidth  = 20
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	fn := styleListItem.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return styleListSelected.Render(cursor + strings.Join(s, " "))
		}
	}
	_, _ = w.Write([]byte(fn(fmt.Sprintf(" %s", i))))
}
