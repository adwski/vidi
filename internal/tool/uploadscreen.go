package tool

import (
	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"os"
	"strings"
)

const (
	uploadControlMsgFileSelected = iota + 1
	uploadControlMsgDone
)

var (
	allowedFileTypes = []string{".mp4"}
)

type (
	sUpload struct {
		filePicker    filepicker.Model
		form          *huh.Form
		renderEventCh chan any
		videoName     string
		selectedFile  string
		errMsg        string
	}

	uploadControl struct {
		msg           int
		renderEventCh chan any
		name          string
		path          string
	}
)

func newUploadScreen() *sUpload {
	u := &sUpload{}
	u.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Description("Enter Name of the video").Value(&u.videoName),
			huh.NewConfirm(),
		),
	).WithTheme(defaultHuhTheme)
	u.filePicker = filepicker.New()
	u.filePicker.AllowedTypes = allowedFileTypes
	u.filePicker.Height = 15
	u.filePicker.CurrentDirectory, _ = os.UserHomeDir()
	return u
}

func (s *sUpload) init() tea.Cmd {
	return tea.Batch(s.filePicker.Init(), s.form.Init())
}

func (s *sUpload) name() string {
	return "uploadScreen"
}

func (s *sUpload) update(msg tea.Msg) (tea.Cmd, *outerControl) {
	if m, ok := msg.(tea.KeyMsg); ok {
		switch m.String() {
		case "esc":
			return nil, &outerControl{data: uploadControl{}}
		}
	}

	var cmd tea.Cmd
	if len(s.selectedFile) == 0 {
		s.filePicker, cmd = s.filePicker.Update(msg)
		if ok, path := s.filePicker.DidSelectFile(msg); ok {
			s.selectedFile = path
		}
		if ok, _ := s.filePicker.DidSelectDisabledFile(msg); ok {
			s.errMsg = "You can use only mp4 file"
			s.selectedFile = ""
			return nil, nil
		}
	} else {
		var form tea.Model
		form, cmd = s.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			s.form = f
		}
		if s.form.State == huh.StateCompleted {
			s.renderEventCh = make(chan any)
			return nil, &outerControl{data: uploadControl{
				msg:  uploadControlMsgFileSelected,
				name: s.videoName,
				path: s.selectedFile,
			}}
		}
	}
	return cmd, nil
}

func (s *sUpload) view() string {
	if len(s.selectedFile) == 0 {
		return s.renderFilePicker()
	} else {
		return s.renderForm()
	}
}

func (s *sUpload) renderForm() string {
	return tableContainer.Render("Selected file:\n\n" + s.selectedFile + "\n\n" + s.form.View())
}

func (s *sUpload) renderFilePicker() string {
	var sb strings.Builder
	if len(s.errMsg) > 0 {
		sb.WriteString(s.errMsg + "\n\n")
	} else {
		sb.WriteString("Please select mp4 file\n\n")
	}
	return tableContainer.Render(sb.String() + s.filePicker.View())
}
