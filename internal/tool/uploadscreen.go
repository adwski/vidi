package tool

import (
	"errors"
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/progress"
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
		filePicker   filepicker.Model
		form         *huh.Form
		progress     progress.Model
		videoName    string
		selectedFile string
		err          error
		uploading    bool
		done         bool
	}

	uploadControl struct {
		msg  int
		name string
		path string
	}

	uploadCompleted struct{}

	uploadProgress struct {
		completed uint64
		total     uint64
	}
)

func newUploadScreen() *sUpload {
	u := &sUpload{
		progress: progress.New(progress.WithDefaultGradient()),
	}
	u.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Description("Enter Name of the video").Value(&u.videoName),
			huh.NewConfirm(),
		),
	).WithTheme(defaultHuhTheme)
	u.filePicker = filepicker.New()
	u.filePicker.AllowedTypes = allowedFileTypes
	u.filePicker.Height = 20
	u.filePicker.ShowPermissions = false
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
	if !s.uploading {
		// user is still entering upload params, can go back with esc
		if m, ok := msg.(tea.KeyMsg); ok {
			switch m.String() {
			case "esc":
				return nil, &outerControl{data: uploadControl{}}
			}
		}

		if len(s.selectedFile) == 0 {
			// User chooses file
			var cmd tea.Cmd
			s.filePicker, cmd = s.filePicker.Update(msg)
			if ok, path := s.filePicker.DidSelectFile(msg); ok {
				s.selectedFile = path
			}
			if ok, _ := s.filePicker.DidSelectDisabledFile(msg); ok {
				s.err = errors.New("you can use only mp4 file")
				s.selectedFile = ""
			}
			return cmd, nil
		}

		// File is selected, user enters name
		form, cmd := s.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			s.form = f
		}
		if s.form.State == huh.StateCompleted {
			// form is confirmed, send msg to start upload
			s.uploading = true
			return nil, &outerControl{data: uploadControl{
				msg:  uploadControlMsgFileSelected,
				name: s.videoName,
				path: s.selectedFile,
			}}
		}
		return cmd, nil
	}

	// upload should start, wait for progress messages
	switch m := msg.(type) {
	case tea.KeyMsg:
		if s.done {
			// allow user to exit screen with any key after upload is finished
			return nil, &outerControl{data: uploadControl{msg: uploadControlMsgDone}}
		}
	case error:
		// error occurred, set finishing flag
		s.err = m
		s.done = true
	case uploadCompleted:
		// all ok, set finishing flag
		s.done = true
		cmd := s.progress.SetPercent(1.0) // just in case
		return cmd, nil
	case uploadProgress:
		// update progress bar percentage
		cmd := s.progress.SetPercent(float64(m.completed) / float64(m.total))
		return cmd, nil
	case progress.FrameMsg:
		// animations
		progressModel, cmd := s.progress.Update(msg)
		s.progress = progressModel.(progress.Model)
		return cmd, nil
	}
	return nil, nil
}

func (s *sUpload) view() string {
	var content string
	switch {
	case s.uploading:
		content = s.renderUploadProgress()
	case len(s.selectedFile) == 0:
		content = s.renderFilePicker()
	default:
		content = s.renderForm()
	}
	return containerWithBorder.Width(100).Render(content)
}

func (s *sUpload) renderUploadProgress() string {
	var sb strings.Builder
	sb.WriteString("Uploading file: \n\n")
	sb.WriteString("Name: ")
	sb.WriteString(s.videoName)
	sb.WriteString("\n\n")
	sb.WriteString("Path: ")
	sb.WriteString(s.selectedFile)
	sb.WriteString("\n\n")
	if s.err != nil {
		sb.WriteString("Upload error: ")
		sb.WriteString(s.err.Error())
		sb.WriteString("\nPress any key to continue...\n\n")
	} else if s.done {
		sb.WriteString("Upload completed successfully! Press any key to continue...\n\n")
	}
	return sb.String() + s.progress.View()
}

func (s *sUpload) renderForm() string {
	return "Selected file:\n\n" + s.selectedFile + "\n\n" + s.form.View()
}

func (s *sUpload) renderFilePicker() string {
	var sb strings.Builder
	sb.WriteString("Please select mp4 file")
	if s.err != nil {
		sb.WriteString(" (" + s.err.Error() + ")")
	}
	sb.WriteString("\n\n")

	return sb.String() + s.filePicker.View()
}
