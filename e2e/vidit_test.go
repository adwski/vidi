package e2e

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/adwski/vidi/internal/tool"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/labstack/gommon/random"
	"github.com/stretchr/testify/require"
)

const (
	VIDITe2eUserTmpl = "vidittestuser"
	VIDITe2ePassword = "vidittestpassword"

	testRCFG = `{
  "user_api_url": "http://localhost:18081/api/user",
  "video_api_url": "localhost:18092",
  "vidi_ca": "%s"
}`
)

var (
	VIDITe2eUser string
)

func init() {
	VIDITe2eUser = VIDITe2eUserTmpl + strconv.Itoa(int(time.Now().Unix()))
}

func TestVidit_MainFlow(t *testing.T) {
	// --------------------------------------------------------------------------------------
	// Prepare remote config and serve it
	b, err := os.ReadFile("cert.pem")
	require.NoError(t, err)

	remoteConfig := fmt.Sprintf(testRCFG, base64.StdEncoding.EncodeToString(b))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, errW := w.Write([]byte(remoteConfig))
		require.NoError(t, errW)
	}))
	defer srv.Close()

	// --------------------------------------------------------------------------------------
	// Init tool
	homeDir := t.TempDir()
	vidit, err := tool.NewWithConfig(tool.Config{
		EnforceHomeDir: homeDir,
		FilePickerDir:  homeDir,
		EarlyInit:      true,
	})
	require.NoError(t, err)
	require.NotNil(t, vidit)

	// --------------------------------------------------------------------------------------
	// Create teatest program
	tm := teatest.NewTestModel(t, vidit, teatest.WithInitialTermSize(300, 100))

	// --------------------------------------------------------------------------------------
	// Run tool
	errc := make(chan error)
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(1)
	go vidit.RunWithProgram(ctx, wg, errc, tm.GetProgram())
	go func() {
		for {
			select {
			case errR := <-errc:
				require.NoError(t, errR)
			case <-ctx.Done():
				return
			}
		}
	}()

	// --------------------------------------------------------------------------------------
	// Here should be config screen
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Configure ViDi endpoint URL"))
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second*3))
	t.Log("config screen showed")

	// enter endpoint
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(srv.URL),
	})
	time.Sleep(time.Millisecond * 200)
	// press enter
	tm.Send(tea.KeyMsg{
		Type: tea.KeyEnter,
	})

	// --------------------------------------------------------------------------------------
	// Here should be new user screen
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		// t.Log(string(bts))
		return bytes.Contains(bts, []byte("No locally stored users have found"))
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second*3))
	t.Log("new user screen showed")

	// choose register
	tm.Send(tea.KeyMsg{
		Type: tea.KeyDown,
	})
	time.Sleep(time.Millisecond * 200)
	tm.Send(tea.KeyMsg{
		Type: tea.KeyEnter,
	})

	// --------------------------------------------------------------------------------------
	// Here should be new user screen, second stage
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Provide user credentials"))
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second*3))
	t.Log("new user second screen showed")

	// enter creds
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(VIDITe2eUser),
	})
	time.Sleep(time.Millisecond * 200)
	tm.Send(tea.KeyMsg{
		Type: tea.KeyEnter,
	})
	time.Sleep(time.Millisecond * 200)
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(VIDITe2ePassword),
	})
	time.Sleep(time.Millisecond * 200)
	tm.Send(tea.KeyMsg{
		Type: tea.KeyEnter,
	})
	time.Sleep(time.Millisecond * 200)
	// confirm
	tm.Send(tea.KeyMsg{
		Type: tea.KeyLeft,
	})
	time.Sleep(time.Millisecond * 200)
	tm.Send(tea.KeyMsg{
		Type: tea.KeyEnter,
	})

	// --------------------------------------------------------------------------------------
	// Here should be main screen
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Welcome to Vidi terminal tool"))
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second*3))
	t.Log("main menu screen showed")

	// --------------------------------------------------------------------------------------
	// Copy mp4 file, so it would be easier to find it in file picker
	fileName := fmt.Sprintf("testvideo%s.mp4", random.String(5, "asdzxcqwe"))
	require.NoError(t, copyFile(homeDir+"/"+fileName, "../testfiles/test_seq_h264_high_uhd.mp4"),
		"copy test video file to home dir")

	// --------------------------------------------------------------------------------------
	// Upload file

	// select upload menu
	tm.Send(tea.KeyMsg{
		Type: tea.KeyDown,
	})
	time.Sleep(time.Millisecond * 200)
	tm.Send(tea.KeyMsg{
		Type: tea.KeyEnter,
	})
	time.Sleep(time.Millisecond * 200)

	// --------------------------------------------------------------------------------------
	// Here should be file picker
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte(fileName))
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second*3))
	t.Log("file picker showed")

	// select file
	tm.Send(tea.KeyMsg{
		Type: tea.KeyDown,
	})
	time.Sleep(time.Millisecond * 200)
	tm.Send(tea.KeyMsg{
		Type: tea.KeyDown,
	})
	time.Sleep(time.Millisecond * 200)
	tm.Send(tea.KeyMsg{
		Type: tea.KeyEnter,
	})
	time.Sleep(time.Millisecond * 200)

	// --------------------------------------------------------------------------------------
	// Here should be file picker second screen
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Enter Name of the video"))
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second*3))

	videoName := fmt.Sprintf("test video %s", random.String(5, "asdzxcqwe123"))
	// enter name
	tm.Send(tea.KeyMsg{
		Type: tea.KeyDown,
	})
	time.Sleep(time.Millisecond * 200)
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(videoName),
	})
	time.Sleep(time.Millisecond * 200)
	tm.Send(tea.KeyMsg{
		Type: tea.KeyEnter,
	})
	time.Sleep(time.Millisecond * 200)
	// confirm
	tm.Send(tea.KeyMsg{
		Type: tea.KeyLeft,
	})
	time.Sleep(time.Millisecond * 200)
	tm.Send(tea.KeyMsg{
		Type: tea.KeyEnter,
	})
	time.Sleep(time.Millisecond * 200)

	// --------------------------------------------------------------------------------------
	// Here should eventually be upload success message
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Upload completed successfully! Press any key to continue..."))
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second*3))

	tm.Send(tea.KeyMsg{
		Type: tea.KeyEnter,
	})
	time.Sleep(time.Millisecond * 200)

	// --------------------------------------------------------------------------------------
	// Here should be main screen
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Welcome to Vidi terminal tool"))
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second*3))
	t.Log("main menu screen showed")

	time.Sleep(time.Second * 15)
	tm.Send(tea.KeyMsg{
		Type: tea.KeyEnter,
	})
	time.Sleep(time.Millisecond * 200)

	// --------------------------------------------------------------------------------------
	// Here should be videos screen
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte(videoName)) && bytes.Contains(bts, []byte("ready"))
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second*1))
	t.Log("videos screen showed and uploaded video is ready")

	tm.Send(tea.KeyMsg{
		Type: tea.KeyBackspace,
	})
	time.Sleep(time.Millisecond * 200)

	// --------------------------------------------------------------------------------------
	// Here should be main screen
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Welcome to Vidi terminal tool"))
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second*3))
	t.Log("main menu screen showed")

	// goto quotas
	tm.Send(tea.KeyMsg{
		Type: tea.KeyDown,
	})
	time.Sleep(time.Millisecond * 200)
	tm.Send(tea.KeyMsg{
		Type: tea.KeyDown,
	})
	time.Sleep(time.Millisecond * 200)
	tm.Send(tea.KeyMsg{
		Type: tea.KeyEnter,
	})
	time.Sleep(time.Millisecond * 200)

	// --------------------------------------------------------------------------------------
	// Here should be quotas screen
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("size_quota")) &&
			bytes.Contains(bts, []byte("size_usage")) &&
			bytes.Contains(bts, []byte("size_remain")) &&
			bytes.Contains(bts, []byte("videos_quota")) &&
			bytes.Contains(bts, []byte("videos_usage")) &&
			bytes.Contains(bts, []byte("videos_remain"))
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second*3))
	t.Log("quotas screen showed")

	// go back
	tm.Send(tea.KeyMsg{
		Type: tea.KeyBackspace,
	})
	time.Sleep(time.Millisecond * 200)

	// --------------------------------------------------------------------------------------
	// Here should be main screen
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Welcome to Vidi terminal tool"))
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second*3))
	t.Log("main menu screen showed")

	// goto videos
	time.Sleep(time.Second * 15)
	tm.Send(tea.KeyMsg{
		Type: tea.KeyEnter,
	})
	time.Sleep(time.Millisecond * 200)

	// --------------------------------------------------------------------------------------
	// Here should be videos screen
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte(videoName)) && bytes.Contains(bts, []byte("ready"))
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second*1))
	t.Log("videos screen showed and uploaded video is ready")

	// gen watch url
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("w"),
	})
	time.Sleep(time.Millisecond * 200)

	// --------------------------------------------------------------------------------------
	// Here should be videos screen
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("/manifest.mpd"))
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second*1))
	t.Log("watch url was generated")

	// delete video
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("d"),
	})
	time.Sleep(time.Millisecond * 200)

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("y"),
	})
	time.Sleep(time.Millisecond * 200)

	// --------------------------------------------------------------------------------------
	// Here should be main screen
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Welcome to Vidi terminal tool"))
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second*3))
	t.Log("video deleted and main menu screen showed")

	tm.Send(tea.KeyMsg{
		Type: tea.KeyEnter,
	})
	time.Sleep(time.Millisecond * 200)

	// --------------------------------------------------------------------------------------
	// Here should be videos screen
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("<no videos to show>"))
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second*1))
	t.Log("videos screen showed and no videos are present")

	// --------------------------------------------------------------------------------------
	// Cleanup
	cancel()
	wg.Wait()

	b, err = os.ReadFile(homeDir + "/log.json")
	require.NoError(t, err)
	t.Log(string(b))
}

func copyFile(dst string, src string) error {
	fSrc, err := os.Open(src)
	if err != nil {
		return err //nolint:wrapcheck // unnecessary
	}
	fDst, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err //nolint:wrapcheck // unnecessary
	}
	_, err = fSrc.WriteTo(fDst)
	return err //nolint:wrapcheck // unnecessary
}
