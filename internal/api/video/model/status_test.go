package model

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatus_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name   string
		arg    []byte
		err    string
		status Status
	}{
		{
			name:   "unmarshall string",
			arg:    []byte(`"ready"`),
			status: StatusReady,
		},
		{
			name:   "unmarshall num",
			arg:    []byte(`0`),
			status: StatusCreated,
		},
		{
			name: "unmarshall invalid num",
			arg:  []byte(`-1233`),
			err:  "unknown status num",
		},
		{
			name: "unmarshall invalid type",
			arg:  []byte(`{"a":"b"}`),
			err:  "invalid type",
		},
		{
			name: "unmarshall invalid json",
			arg:  []byte(`qweqwe1231`),
			err:  "invalid character",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s Status

			err := s.UnmarshalJSON(tt.arg)
			if tt.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.status, s)
			} else {
				assert.Contains(t, err.Error(), tt.err)
			}
		})
	}
}

func TestGetStatusFromName(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want Status
		err  error
	}{
		{
			name: "valid status",
			arg:  "uploaded",
			want: StatusUploaded,
		},
		{
			name: "invalid status",
			arg:  "qweqweqwasd",
			err:  errors.New("incorrect status name"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := GetStatusFromName(tt.arg)
			assert.Equal(t, tt.want, s)
			assert.Equal(t, tt.err, err)
		})
	}
}
