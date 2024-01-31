//nolint:dupl // similar test flows
package config

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupViper(t *testing.T, r io.Reader) *ViperEC {
	t.Helper()

	vec := NewViperEC()
	vec.SetConfigType("yaml")
	err := vec.ReadConfig(r)
	require.NoError(t, err)

	return vec
}

func TestViperEC_GetDuration(t *testing.T) {
	type args struct {
		config io.Reader
		key    string
	}
	tests := []struct {
		name string
		args args
		want time.Duration
		err  string
	}{
		{
			name: "get duration",
			args: args{
				key:    "duration",
				config: bytes.NewReader([]byte("duration: 5s")),
			},
			want: 5 * time.Second,
			err:  "",
		},
		{
			name: "get duration error",
			args: args{
				key:    "duration",
				config: bytes.NewReader([]byte("duration: sss")),
			},
			err: "invalid",
		},
		{
			name: "get duration zero",
			args: args{
				key:    "duration",
				config: bytes.NewReader([]byte("duration: 0")),
			},
			err: "cannot be zero",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vec := setupViper(t, tt.args.config)
			dur := vec.GetDuration(tt.args.key)
			assert.Equal(t, tt.want, dur)
			if tt.err == "" {
				assert.Empty(t, vec.Errors())
			} else {
				assert.True(t, vec.HasErrors())
				assert.Contains(t, vec.Errors()[tt.args.key].Error(), tt.err)
			}
		})
	}
}

func TestViperEC_GetBool(t *testing.T) {
	type args struct {
		config io.Reader
		key    string
	}
	tests := []struct {
		name string
		args args
		want bool
		err  string
	}{
		{
			name: "get bool",
			args: args{
				key:    "key",
				config: bytes.NewReader([]byte("key: true")),
			},
			want: true,
			err:  "",
		},
		{
			name: "get duration error",
			args: args{
				key:    "key",
				config: bytes.NewReader([]byte("key: sss")),
			},
			err: "invalid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vec := setupViper(t, tt.args.config)
			val := vec.GetBool(tt.args.key)
			assert.Equal(t, tt.want, val)
			if tt.err == "" {
				assert.Empty(t, vec.Errors())
			} else {
				assert.True(t, vec.HasErrors())
				assert.Contains(t, vec.Errors()[tt.args.key].Error(), tt.err)
			}
		})
	}
}

func TestViperEC_GetBoolNoError(t *testing.T) {
	type args struct {
		config io.Reader
		key    string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "get bool",
			args: args{
				key:    "key",
				config: bytes.NewReader([]byte("key: true")),
			},
			want: true,
		},
		{
			name: "get duration error",
			args: args{
				key:    "key",
				config: bytes.NewReader([]byte("key: sss")),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vec := setupViper(t, tt.args.config)
			val := vec.GetBool(tt.args.key)
			assert.Equal(t, tt.want, val)
		})
	}
}

func TestViperEC_GetURL(t *testing.T) {
	type args struct {
		config io.Reader
		key    string
	}
	tests := []struct {
		name string
		args args
		want string
		err  string
	}{
		{
			name: "get url",
			args: args{
				key:    "key",
				config: bytes.NewReader([]byte("key: http://wer.asd")),
			},
			want: "http://wer.asd",
			err:  "",
		},
		{
			name: "get url error",
			args: args{
				key:    "key",
				config: bytes.NewReader([]byte("key: :wer.asd")),
			},
			err: "is not valid url",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vec := setupViper(t, tt.args.config)
			dur := vec.GetURL(tt.args.key)
			assert.Equal(t, tt.want, dur)
			if tt.err == "" {
				assert.Empty(t, vec.Errors())
			} else {
				assert.True(t, vec.HasErrors())
				assert.Contains(t, vec.Errors()[tt.args.key].Error(), tt.err)
			}
		})
	}
}

func TestViperEC_GetURIPrefix(t *testing.T) {
	type args struct {
		config io.Reader
		key    string
	}
	tests := []struct {
		name string
		args args
		want string
		err  string
	}{
		{
			name: "get uri prefix",
			args: args{
				key:    "key",
				config: bytes.NewReader([]byte("key: /api")),
			},
			want: "/api",
		},
		{
			name: "get uri suffix error",
			args: args{
				key:    "key",
				config: bytes.NewReader([]byte("key:  /api/")),
			},
			err: "must not end with",
		},
		{
			name: "get uri empty",
			args: args{
				key:    "key",
				config: bytes.NewReader([]byte("key: ")),
			},
			err: "cannot be empty",
		},
		{
			name: "get uri prefix error",
			args: args{
				key:    "key",
				config: bytes.NewReader([]byte("key:  api/")),
			},
			err: "must start with",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vec := setupViper(t, tt.args.config)
			dur := vec.GetURIPrefix(tt.args.key)
			assert.Equal(t, tt.want, dur)
			if tt.err == "" {
				assert.Empty(t, vec.Errors())
			} else {
				assert.True(t, vec.HasErrors())
				assert.Contains(t, vec.Errors()[tt.args.key].Error(), tt.err)
			}
		})
	}
}

func TestViperEC_GetString(t *testing.T) {
	type args struct {
		config io.Reader
		key    string
	}
	tests := []struct {
		name string
		args args
		want string
		err  string
	}{
		{
			name: "get string",
			args: args{
				key:    "key",
				config: bytes.NewReader([]byte("key: qwe")),
			},
			want: "qwe",
		},
		{
			name: "get string empty",
			args: args{
				key:    "key",
				config: bytes.NewReader([]byte("key: ")),
			},
			err: "cannot be empty",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vec := setupViper(t, tt.args.config)
			dur := vec.GetString(tt.args.key)
			assert.Equal(t, tt.want, dur)
			if tt.err == "" {
				assert.Empty(t, vec.Errors())
			} else {
				assert.True(t, vec.HasErrors())
				assert.Contains(t, vec.Errors()[tt.args.key].Error(), tt.err)
			}
		})
	}
}
