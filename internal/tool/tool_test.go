//nolint:lll // long config samples
package tool

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/stretchr/testify/require"
)

func TestTool_StartNoEndpointFailedConnect(t *testing.T) {
	tool, err := New()
	require.NoError(t, err)

	tool.dir = t.TempDir()
	tool.initialize()
	require.NoError(t, tool.err)

	tm := teatest.NewTestModel(t, tool, teatest.WithInitialTermSize(300, 100))
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Configure ViDi endpoint URL"))
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second*3))

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("http://127.0.0.1:12345"),
	})
	tm.Send(tea.KeyMsg{
		Type: tea.KeyEnter,
	})

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("cannot contact ViDi endpoint"))
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second*3))

	tm.Send(struct{}{})
	tm.Send(tea.KeyMsg{
		Type: tea.KeyEnter,
	})

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Configure ViDi endpoint URL"))
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second*3))

	tm.Send(tea.KeyMsg{
		Type: tea.KeyCtrlC,
	})
}

const validRemoveConfig = `{
  "user_api_url": "http://localhost/api/users",
  "video_api_url": "localhost:443",
  "vidi_ca": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUZoRENDQTJ5Z0F3SUJBZ0lVWTNSZmo0V1I3VkNwUEREZ2liR1Q0V1RySzlBd0RRWUpLb1pJaHZjTkFRRUwKQlFBd09qRUxNQWtHQTFVRUJoTUNVbFV4RFRBTEJnTlZCQW9NQkZacFJHa3hEVEFMQmdOVkJBc01CSFpwWkdreApEVEFMQmdOVkJBTU1CSFpwWkdrd0hoY05NalF3TlRFeE1qRXhOekEzV2hjTk1qVXdOVEV4TWpFeE56QTNXakE2Ck1Rc3dDUVlEVlFRR0V3SlNWVEVOTUFzR0ExVUVDZ3dFVm1sRWFURU5NQXNHQTFVRUN3d0VkbWxrYVRFTk1Bc0cKQTFVRUF3d0VkbWxrYVRDQ0FpSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnSVBBRENDQWdvQ2dnSUJBTGFOdjVPSgp6WmpYWGZ4enhaYWI3L2FJMjZkMWc4TEp1Y0R5SnpQY2JvUDdCSnh5RHFsUXAxT0RGZ2N0S2JmUWY4QlJnZG1oCjBKVFdJd0dyaWNwN2FJOEprd2lvQXZWRmZaOTNtTElCQUc1Zjc5RVp3Y0YyZnllSjNQOG5WdU5FRDVSclR2Z3EKaUNidU5EdkJQMG5uUEhWeDhDYnU1WFpyaGJVV2xZTjkvQW1kb0dFRTJXSlpVb2NJdGxKQnF3bmkzWVpaekxKZQpPZFV6Z00rY2loYjljV3N5OXBvMkZ0V0Z6YnRySnNCaXhLRElUNk9IYTFLTHdoS0RSUFJLSnNmNTNSQ3EvVUlJCkRVM2I1WGx1UFFvQVJvaHl0bzhYdDNDYkgxYWJtRStneUN4QzR6czVsL2dlRVU5N2FQKzRJQ1JlVGFMemdFckUKN21Nakd2NWFQTTFxYVc2MFZyaUwxMUs4ajlwMGpJeDdkRnBpeVZ5S1g1allCVG8zWXBKdTZIc2lldS92MEZyNQo4czdDcmQ2NlFxNmRXR2lpdlV3VXhlNVF5WVJYdkZxQ2hEOERvRHJadGE0UXpQaXhMcldaSjNhU0UwS1JJYyszCi8weE4rYVZhMXQwU0pNMXNJMkMyRVd5ZmtMVWNqcUpLaDNaY0RidHBsUFQ4OXEwV1VEOUVJVXNlU0xtVER4QXIKMURLaEdKaWxhcjUweG1GOVFxYkgrTHFNWWl6V1lPVTBqRUtlSWpsN1RuT2xXSy9WNC9BMDBvL0NnamNCcnVyRgprVS92bTNFbGpYeXNmMlh5UVFmbFdISExVN3hWSElUYmhucGFObXFDaENnQVRFTzhUemlpem1EMEp3TkFRTld1CmdpV3Q0SEVZbDJUdEFRdXdaaEJTNUdmbllBMld2VlQ2ZEs4MUFnTUJBQUdqZ1lFd2Z6QWRCZ05WSFE0RUZnUVUKTmQ5WWlHa0RZVlB2T2JNOEx6UGFOQjY5aDZvd0h3WURWUjBqQkJnd0ZvQVVOZDlZaUdrRFlWUHZPYk04THpQYQpOQjY5aDZvd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBc0JnTlZIUkVFSlRBamdnbHNiMk5oYkdodmMzU0hCSDhBCkFBR0hFQUFBQUFBQUFBQUFBQUFBQUFBQUFBRXdEUVlKS29aSWh2Y05BUUVMQlFBRGdnSUJBSHZOY2lReDd5dWoKUlBqelk0Y2c0dGtKbkNoZjFGcHJWbThYbC9GL1FteC9KK0JOWjNyTm5ZdWVFN1ZXSUV5bzdRQ3Y4NjI4TTF4MwpnRy9iODRSOFJ6ajhWeXhXZWpCdnZpTGNIZW5VZUJTbnVmS2JOUlBnRmFlSWk1L1k3akQ3Qk9vVDAzdzJUUGZYCmZUbG5IRTE5ZndvaDFuUFFRZXJEazhuSFVnZmdhUHIzWXNOWkdPL0MrbmFtcytoTGp4R0dEcWJPa1psQW8wU2YKaCtxQ3h4dVdwT3VJODBLajNOVVc5L0kwTERtTFRzeHBXcGxmT25ibm9RbmUyYXZLdEtadkxHa05vVDFZMTd0bQpSRFRERGhtYTBaTytTaTk0dlF6ZXhWUUg0QnBOd3QydDZnTVRHTW0yRTRBS0FFc2FXWHZ0Y0dPaGhudjJ5NzkyCjJUa3A3aEZmZEFjWlk2bDVpZjZHZGxtU1gwdUVFcWV4ZjRuNjJvaHF1bVp3aVZldHVEUndaMitnSEFwdGwyeG0KekNOUHZOSEY3emxpRGo2VE9OWXoxV0w0NGViNldGbTV6YThlTU4xNFNwd01mUTBRUTV4MjUwOCttNW4yM2JJZApmaUVnVytORTUzLzV5c2wyeWpVSGtiUWhpY1hqZVlsei9sNk9jRVFDN2pGaHVwcmZaWmR2Yzd1QWFJczBINER2CjdYOXdvRlh5OXNzUUwwZk1nemdIQ2M5VmQ4TDlabWNWazBhV3RncWxiRERYM2wreVFHUjBnNW5pcnhDUGZZNC8KWEF2R2p4ZFllcnBxNGFmNkNIeHRCZkdMaGszZnYwbmM4REtDRnZUZUhHTC9SbmkzZnB3SURHN2dTcUZaOFBtVQpTQStKN0Y0UUlNaVNrLy9JYUxlWThBdnRDUG1RY3VsZgotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
}`

func TestTool_StartNoUsers(t *testing.T) {
	tool, err := New()
	require.NoError(t, err)

	tool.dir = t.TempDir()
	tool.initialize()
	require.NoError(t, tool.err)

	tm := teatest.NewTestModel(t, tool, teatest.WithInitialTermSize(300, 100))

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Configure ViDi endpoint URL"))
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second*3))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, errW := w.Write([]byte(validRemoveConfig))
		require.NoError(t, errW)
	}))
	defer srv.Close()

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(srv.URL),
	})
	tm.Send(tea.KeyMsg{
		Type: tea.KeyEnter,
	})

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("No locally stored users have found"))
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second*3))
}

const toolStateValidActiveUser = `{
  "endpoint": "%s",
  "users": [
    {
      "name": "user123",
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiSm9obiBEb2UiLCJleHAiOjQwODI0Njc4Mzl9.TQy6X7dVkSRf92XjN-tRI9-fQjOOml6vcJn3Qb5iNt8"
    }
  ],
  "current_user": 0
}`

func TestTool_StartValidActiveUser(t *testing.T) {
	tool, err := New()
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, errW := w.Write([]byte(validRemoveConfig))
		require.NoError(t, errW)
	}))
	defer srv.Close()

	tool.dir = t.TempDir()
	err = os.WriteFile(tool.dir+stateFile, []byte(fmt.Sprintf(toolStateValidActiveUser, srv.URL)), 0600)
	require.NoError(t, err)

	tool.initialize()
	require.NoError(t, tool.err)

	tm := teatest.NewTestModel(t, tool, teatest.WithInitialTermSize(300, 100))

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte(greetMessageTxt))
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second*3))
}

const toolStateValidActiveUserExpiredToken = `{
  "endpoint": "%s",
  "users": [
    {
      "name": "testUser",
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiJQQXg3b0RZOVJQZVJVS09FSG1GQXpnIiwibmFtZSI6InVzZXIxMjMiLCJleHAiOjE3MTU1NjI4MjV9.SdMNc0Pf5EWWf5SwPjVpy8PtLbnFT0U-XzoD7-Ayg6Y"
    }
  ],
  "current_user": 0
}`

func TestTool_StartValidActiveUserExpiredToken(t *testing.T) {
	tool, err := New()
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, errW := w.Write([]byte(validRemoveConfig))
		require.NoError(t, errW)
	}))
	defer srv.Close()

	tool.dir = t.TempDir()
	err = os.WriteFile(tool.dir+stateFile, []byte(fmt.Sprintf(toolStateValidActiveUserExpiredToken, srv.URL)), 0600)
	require.NoError(t, err)

	tool.initialize()
	require.NoError(t, tool.err)

	tm := teatest.NewTestModel(t, tool, teatest.WithInitialTermSize(300, 100))

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Enter your password again or select another user")) &&
			bytes.Contains(bts, []byte("Enter password for 'testUser'")) &&
			bytes.Contains(bts, []byte("Login with another user"))
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second*3))

	tm.Send(tea.KeyMsg{
		Type: tea.KeyEnter,
	})

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Enter password for 'testUser'")) &&
			!bytes.Contains(bts, []byte("Login with another user"))
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second*3))
}
