//nolint:lll // have jwt tokens in test data
package tool

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_isURL(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "valid url",
			args: args{
				url: "http://example.com/123",
			},
			wantErr: assert.NoError,
		},
		{
			name: "invalid url",
			args: args{
				url: "://example.com",
			},
			wantErr: assert.Error,
		},
		{
			name: "no scheme",
			args: args{
				url: "example.com",
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid scheme",
			args: args{
				url: "qweqwe://example.com",
			},
			wantErr: assert.Error,
		},
		{
			name: "missing hostname",
			args: args{
				url: "http://:123",
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, isURL(tt.args.url), fmt.Sprintf("isURL(%v)", tt.args.url))
		})
	}
}

func TestState_activeUserUnsafe(t *testing.T) {
	type args struct {
		Users       []User
		CurrentUser int
	}
	type want struct {
		selectedUser string
		noUsers      bool
	}
	var tests = []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid active user",
			args: args{
				Users: []User{
					{
						Name: "test",
					},
				},
				CurrentUser: 0,
			},
			want: want{
				selectedUser: "test",
				noUsers:      false,
			},
		},
		{
			name: "no user selected",
			args: args{
				Users: []User{
					{
						Name: "test",
					},
				},
				CurrentUser: -1,
			},
			want: want{
				noUsers: false,
			},
		},
		{
			name: "no users",
			args: args{
				CurrentUser: -1,
			},
			want: want{
				noUsers: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &State{}
			s.Users = tt.args.Users
			s.CurrentUser = tt.args.CurrentUser

			if tt.want.noUsers {
				require.True(t, s.noUsers())
				require.Nil(t, s.activeUserUnsafe())
			} else {
				require.False(t, s.noUsers())
				u := s.activeUserUnsafe()
				if tt.want.selectedUser != "" {
					require.NotNil(t, u)
					require.Equal(t, tt.want.selectedUser, u.Name)
					require.Equal(t, tt.args.CurrentUser, s.userID(u.Name))
				} else {
					require.Nil(t, u)
				}
			}
		})
	}
}

const (
	stateFileValidContent = `{
  "users": [
    {
      "upload": null,
      "name": "user123",
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiJQQXg3b0RZOVJQZVJVS09FSG1GQXpnIiwibmFtZSI6InVzZXIxMjMiLCJleHAiOjE3MTU1NjI4MjV9.SdMNc0Pf5EWWf5SwPjVpy8PtLbnFT0U-XzoD7-Ayg6Y",
      "expires_at": 1715562825
    }
  ],
  "endpoint": "http://127.0.0.1:1234",
  "current_user": 0
}`

	stateFileCorruptedParams = `{
  "users": [
    {
      "name": "user1"
    },
    {
      "name": "u"
    },
    {
      "name": "user2"
    }
  ],
  "endpoint": "h:ttp://127weww.0.:0:.1:1234",
  "current_user": 10
}`

	stateFileNoUsersCorruptedCurrentUser = `{
  "current_user": 10
}`
)

func TestState_load(t *testing.T) {
	type args struct {
		stateFileContent string
	}
	type want struct {
		errMsg   string
		numUsers int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "state file not exists",
			args: args{},
			want: want{
				errMsg:   "cannot open state file",
				numUsers: 0,
			},
		},
		{
			name: "state file valid",
			args: args{
				stateFileContent: stateFileValidContent,
			},
			want: want{
				numUsers: 1,
			},
		},
		{
			name: "corrupted params",
			args: args{
				stateFileContent: stateFileCorruptedParams,
			},
			want: want{
				numUsers: 2,
			},
		},
		{
			name: "corrupted current user",
			args: args{
				stateFileContent: stateFileNoUsersCorruptedCurrentUser,
			},
			want: want{
				numUsers: 0,
			},
		},
		{
			name: "invalid json",
			args: args{
				stateFileContent: "asdasdasdsad",
			},
			want: want{
				errMsg:   "cannot unmarshal state file",
				numUsers: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newState(t.TempDir())

			if tt.args.stateFileContent != "" {
				err := os.WriteFile(s.dir+stateFile, []byte(tt.args.stateFileContent), 0600)
				require.NoError(t, err)
			}
			err := s.load()
			if tt.want.errMsg != "" {
				require.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.want.errMsg)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want.numUsers, len(s.Users))
			if tt.want.numUsers == 0 {
				assert.Equal(t, -1, s.CurrentUser)
			}
		})
	}
}

const (
	stateFileExpiredToken = `{
  "users": [
    {
      "name": "user123",
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiJQQXg3b0RZOVJQZVJVS09FSG1GQXpnIiwibmFtZSI6InVzZXIxMjMiLCJleHAiOjE3MTU1NjI4MjV9.SdMNc0Pf5EWWf5SwPjVpy8PtLbnFT0U-XzoD7-Ayg6Y",
      "expires_at": 1715562825
    }
  ],
  "current_user": 0
}`

	stateFileNoToken = `{
  "users": [
    {
      "name": "user123"
    }
  ],
  "current_user": 0
}`
	stateFileInvalidToken = `{
  "users": [
    {
      "name": "user123",
      "token": "asdad3q3e"
    }
  ],
  "current_user": 0
}`

	stateFileValidToken = `{
  "users": [
    {
      "name": "user123",
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiSm9obiBEb2UiLCJleHAiOjQwODI0Njc4Mzl9.TQy6X7dVkSRf92XjN-tRI9-fQjOOml6vcJn3Qb5iNt8"
    }
  ],
  "current_user": 0
}`

	stateFileInvalidTokenNoExp = `{
  "users": [
    {
      "name": "user123",
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiSm9obiBEb2UifQ.fpbbnpvxSC8wbH3TzgGYp14yUVb4WFY5kR5D3QALT60"
    }
  ],
  "current_user": 0
}`

	stateFileInvalidTokenInvalidExp = `{
  "users": [
    {
      "name": "user123",
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiSm9obiBEb2UiLCJleHAiOiJxcXdlcXdlcXcifQ.W49mBsvLxRiVHKoxzADqPNfRxMbavVryKO1BDBnBTZw"
    }
  ],
  "current_user": 0
}`
)

func TestState_checkToken(t *testing.T) {
	type args struct {
		stateFileContent string
	}
	type want struct {
		errMsg string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "token expired",
			args: args{
				stateFileContent: stateFileExpiredToken,
			},
			want: want{
				errMsg: "token is expired",
			},
		},
		{
			name: "token missing",
			args: args{
				stateFileContent: stateFileNoToken,
			},
			want: want{
				errMsg: "token is empty",
			},
		},
		{
			name: "token invalid",
			args: args{
				stateFileContent: stateFileInvalidToken,
			},
			want: want{
				errMsg: "cannot parse token",
			},
		},
		{
			name: "token valid",
			args: args{
				stateFileContent: stateFileValidToken,
			},
			want: want{},
		},
		{
			name: "token with no exp time",
			args: args{
				stateFileContent: stateFileInvalidTokenNoExp,
			},
			want: want{
				"token has no expiration time",
			},
		},
		{
			name: "token with invalid exp time",
			args: args{
				stateFileContent: stateFileInvalidTokenInvalidExp,
			},
			want: want{
				"unable to get expiration time from token",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newState(t.TempDir())

			err := os.WriteFile(s.dir+stateFile, []byte(tt.args.stateFileContent), 0600)
			require.NoError(t, err)
			err = s.load()
			require.NoError(t, err)

			err = s.checkToken()
			if tt.want.errMsg != "" {
				require.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.want.errMsg)
			} else {
				require.NoError(t, err)
				assert.Greater(t, s.activeUserUnsafe().TokenExpiresAt, int64(0))
			}
		})
	}
}
