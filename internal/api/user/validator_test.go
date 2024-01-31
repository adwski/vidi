package user

import (
	"testing"

	"github.com/adwski/vidi/internal/api/user/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRequestValidator_Validate(t *testing.T) {
	tests := []struct {
		name string
		args any
		err  string
	}{
		{
			name: "no error",
			args: &model.UserRequest{
				Username: "testUser",
				Password: "testPass",
			},
		},
		{
			name: "request validation error",
			args: &model.UserRequest{
				Username: "test",
				Password: "test",
			},
			err: "missing required params",
		},
		{
			name: "unknown error",
			args: nil,
			err:  "unknown error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := zap.NewDevelopment()
			require.NoError(t, err)
			rv := NewRequestValidator(logger)

			err = rv.Validate(tt.args)
			if tt.err != "" {
				assert.Contains(t, err.Error(), tt.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
