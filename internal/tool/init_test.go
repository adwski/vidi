package tool

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:lll // certificate
const testRCFG = `{
  "user_api_url": "http://localhost/api/users",
  "video_api_url": "localhost:443",
  "vidi_ca": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUZoRENDQTJ5Z0F3SUJBZ0lVWTNSZmo0V1I3VkNwUEREZ2liR1Q0V1RySzlBd0RRWUpLb1pJaHZjTkFRRUwKQlFBd09qRUxNQWtHQTFVRUJoTUNVbFV4RFRBTEJnTlZCQW9NQkZacFJHa3hEVEFMQmdOVkJBc01CSFpwWkdreApEVEFMQmdOVkJBTU1CSFpwWkdrd0hoY05NalF3TlRFeE1qRXhOekEzV2hjTk1qVXdOVEV4TWpFeE56QTNXakE2Ck1Rc3dDUVlEVlFRR0V3SlNWVEVOTUFzR0ExVUVDZ3dFVm1sRWFURU5NQXNHQTFVRUN3d0VkbWxrYVRFTk1Bc0cKQTFVRUF3d0VkbWxrYVRDQ0FpSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnSVBBRENDQWdvQ2dnSUJBTGFOdjVPSgp6WmpYWGZ4enhaYWI3L2FJMjZkMWc4TEp1Y0R5SnpQY2JvUDdCSnh5RHFsUXAxT0RGZ2N0S2JmUWY4QlJnZG1oCjBKVFdJd0dyaWNwN2FJOEprd2lvQXZWRmZaOTNtTElCQUc1Zjc5RVp3Y0YyZnllSjNQOG5WdU5FRDVSclR2Z3EKaUNidU5EdkJQMG5uUEhWeDhDYnU1WFpyaGJVV2xZTjkvQW1kb0dFRTJXSlpVb2NJdGxKQnF3bmkzWVpaekxKZQpPZFV6Z00rY2loYjljV3N5OXBvMkZ0V0Z6YnRySnNCaXhLRElUNk9IYTFLTHdoS0RSUFJLSnNmNTNSQ3EvVUlJCkRVM2I1WGx1UFFvQVJvaHl0bzhYdDNDYkgxYWJtRStneUN4QzR6czVsL2dlRVU5N2FQKzRJQ1JlVGFMemdFckUKN21Nakd2NWFQTTFxYVc2MFZyaUwxMUs4ajlwMGpJeDdkRnBpeVZ5S1g1allCVG8zWXBKdTZIc2lldS92MEZyNQo4czdDcmQ2NlFxNmRXR2lpdlV3VXhlNVF5WVJYdkZxQ2hEOERvRHJadGE0UXpQaXhMcldaSjNhU0UwS1JJYyszCi8weE4rYVZhMXQwU0pNMXNJMkMyRVd5ZmtMVWNqcUpLaDNaY0RidHBsUFQ4OXEwV1VEOUVJVXNlU0xtVER4QXIKMURLaEdKaWxhcjUweG1GOVFxYkgrTHFNWWl6V1lPVTBqRUtlSWpsN1RuT2xXSy9WNC9BMDBvL0NnamNCcnVyRgprVS92bTNFbGpYeXNmMlh5UVFmbFdISExVN3hWSElUYmhucGFObXFDaENnQVRFTzhUemlpem1EMEp3TkFRTld1CmdpV3Q0SEVZbDJUdEFRdXdaaEJTNUdmbllBMld2VlQ2ZEs4MUFnTUJBQUdqZ1lFd2Z6QWRCZ05WSFE0RUZnUVUKTmQ5WWlHa0RZVlB2T2JNOEx6UGFOQjY5aDZvd0h3WURWUjBqQkJnd0ZvQVVOZDlZaUdrRFlWUHZPYk04THpQYQpOQjY5aDZvd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBc0JnTlZIUkVFSlRBamdnbHNiMk5oYkdodmMzU0hCSDhBCkFBR0hFQUFBQUFBQUFBQUFBQUFBQUFBQUFBRXdEUVlKS29aSWh2Y05BUUVMQlFBRGdnSUJBSHZOY2lReDd5dWoKUlBqelk0Y2c0dGtKbkNoZjFGcHJWbThYbC9GL1FteC9KK0JOWjNyTm5ZdWVFN1ZXSUV5bzdRQ3Y4NjI4TTF4MwpnRy9iODRSOFJ6ajhWeXhXZWpCdnZpTGNIZW5VZUJTbnVmS2JOUlBnRmFlSWk1L1k3akQ3Qk9vVDAzdzJUUGZYCmZUbG5IRTE5ZndvaDFuUFFRZXJEazhuSFVnZmdhUHIzWXNOWkdPL0MrbmFtcytoTGp4R0dEcWJPa1psQW8wU2YKaCtxQ3h4dVdwT3VJODBLajNOVVc5L0kwTERtTFRzeHBXcGxmT25ibm9RbmUyYXZLdEtadkxHa05vVDFZMTd0bQpSRFRERGhtYTBaTytTaTk0dlF6ZXhWUUg0QnBOd3QydDZnTVRHTW0yRTRBS0FFc2FXWHZ0Y0dPaGhudjJ5NzkyCjJUa3A3aEZmZEFjWlk2bDVpZjZHZGxtU1gwdUVFcWV4ZjRuNjJvaHF1bVp3aVZldHVEUndaMitnSEFwdGwyeG0KekNOUHZOSEY3emxpRGo2VE9OWXoxV0w0NGViNldGbTV6YThlTU4xNFNwd01mUTBRUTV4MjUwOCttNW4yM2JJZApmaUVnVytORTUzLzV5c2wyeWpVSGtiUWhpY1hqZVlsei9sNk9jRVFDN2pGaHVwcmZaWmR2Yzd1QWFJczBINER2CjdYOXdvRlh5OXNzUUwwZk1nemdIQ2M5VmQ4TDlabWNWazBhV3RncWxiRERYM2wreVFHUjBnNW5pcnhDUGZZNC8KWEF2R2p4ZFllcnBxNGFmNkNIeHRCZkdMaGszZnYwbmM4REtDRnZUZUhHTC9SbmkzZnB3SURHN2dTcUZaOFBtVQpTQStKN0Y0UUlNaVNrLy9JYUxlWThBdnRDUG1RY3VsZgotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
}`

const testRCFGInvalidCA = `{
  "user_api_url": "http://localhost/api/users",
  "video_api_url": "localhost:443",
  "vidi_ca": "asdasddsa"
}`

const testRCFGMissingCA = `{
  "user_api_url": "http://localhost/api/users",
  "video_api_url": "localhost:443"
}`

func TestInitStateDir(t *testing.T) {
	dir, err := initStateDir(t.TempDir())
	require.NoError(t, err)
	require.NotEmpty(t, dir)
}

func TestTool_initialize(t *testing.T) {
	tool, err := NewWithConfig(Config{EnforceHomeDir: t.TempDir()})
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, errW := w.Write([]byte(testRCFG))
		require.NoError(t, errW)
	}))
	defer srv.Close()

	tool.initialize()

	require.True(t, tool.state.noEndpoint())
	require.NoError(t, tool.err)
	require.NotNil(t, tool.state)

	tool.state.Endpoint = srv.URL
	require.False(t, tool.state.noEndpoint())

	err = tool.initClients(srv.URL)
	require.NoError(t, err)

	assert.NotNil(t, tool.videoapi, "videoAPI client should be initialized")
	assert.NotNil(t, tool.userapi, "userAPI client should be initialized")

	assert.False(t, tool.noClients())
}

func TestTool_initClientsInvalidRemoteConfig(t *testing.T) {
	type args struct {
		epContent string
		epURL     string
	}
	type want struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid endpoint",
			args: args{
				epContent: testRCFG,
			},
			want: want{},
		},
		{
			name: "unreachable endpoint",
			args: args{
				epURL: "http://127.0.0.1:12345",
			},
			want: want{err: errors.New("cannot contact ViDi endpoint")},
		},
		{
			name: "invalid ca",
			args: args{
				epContent: testRCFGInvalidCA,
			},
			want: want{err: errors.New("cannot decode vidi ca")},
		},
		{
			name: "empty ca",
			args: args{
				epContent: testRCFGMissingCA,
			},
			want: want{err: errors.New("vidi ca is empty")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool, err := NewWithConfig(Config{EnforceHomeDir: t.TempDir()})
			require.NoError(t, err)
			require.NotNil(t, tool)
			tool.initialize()
			require.NotNil(t, tool.state)

			epURL := tt.args.epURL
			if tt.args.epContent != "" {
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, errW := w.Write([]byte(tt.args.epContent))
					require.NoError(t, errW)
				}))
				defer srv.Close()
				epURL = srv.URL
			}

			err = tool.initClients(epURL)
			if tt.want.err != nil {
				require.Contains(t, err.Error(), tt.want.err.Error())
				assert.Nil(t, tool.videoapi)
				assert.Nil(t, tool.userapi)
				assert.True(t, tool.noClients())
			} else {
				require.NoError(t, err)
				assert.NotNil(t, tool.videoapi)
				assert.NotNil(t, tool.userapi)
				assert.False(t, tool.noClients())
			}
		})
	}
}
