package tool

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

const (
	mainMenuOptionVideos = iota + 1
	mainMenuOptionUpload
	mainMenuOptionQuotas
	mainMenuOptionSwitchUser
	mainMenuOptionResumeUpload
)

type (
	sMainMenu struct {
		form   *huh.Form
		option int
	}

	mainMenuControl struct {
		option int
	}
)

func (mmc mainMenuControl) String() string {
	switch mmc.option {
	case mainMenuOptionSwitchUser:
		return "switch user"
	case mainMenuOptionVideos:
		return "videos"
	case mainMenuOptionUpload:
		return "upload"
	case mainMenuOptionQuotas:
		return "quota"
	case mainMenuOptionResumeUpload:
		return "resume"
	default:
		return "unknown"
	}
}

func newMainMenuScreen(user string, resumableUpload bool) *sMainMenu {
	var opts []huh.Option[int]
	opts = append(opts,
		huh.NewOption("Videos", mainMenuOptionVideos),
		huh.NewOption("Upload", mainMenuOptionUpload),
		huh.NewOption("Quotas", mainMenuOptionQuotas),
		huh.NewOption("Switch User", mainMenuOptionSwitchUser))
	if resumableUpload {
		opts = append(opts, huh.NewOption("Resume Current Upload", mainMenuOptionResumeUpload))
	}
	smm := &sMainMenu{}
	f := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title(fmt.Sprintf("Logged as %s", user)),
			huh.NewSelect[int]().
				Title("Choose what to do").
				Options(opts...).Value(&smm.option),
		),
	).WithTheme(defaultHuhTheme)
	smm.form = f
	return smm
}

func (s *sMainMenu) init() tea.Cmd {
	return s.form.Init()
}

func (s *sMainMenu) name() string {
	return "mainMenuScreen"
}

func (s *sMainMenu) update(msg tea.Msg) (tea.Cmd, *outerControl) {
	form, cmd := s.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		s.form = f
	}
	if s.form.State == huh.StateCompleted {
		return nil, &outerControl{data: mainMenuControl{
			option: s.option,
		}}
	}
	return cmd, nil
}

func (s *sMainMenu) view() string {
	return greetStyle.Render(vidiSplashText+"\n\n"+greetMessage) + "\n\n" + s.form.View()
}
