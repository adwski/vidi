//nolint:lll // jwt tokens in test cases
package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestAuth_GRPCAuthFunc(t *testing.T) {
	type args struct {
		ctxFunc func() context.Context
	}
	type want struct {
		claims *Claims
		errMsg string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid token",
			args: args{
				ctxFunc: func() context.Context {
					return metadata.NewIncomingContext(
						context.Background(),
						metadata.Pairs("authorization", "bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiJ0ZXN0dXNlcmlkIiwibmFtZSI6InRlc3RVc2VyIiwiZXhwIjo0NzE1ODM5MzY0fQ.QpWCWosIauu56J4LsEvpxBImwvl_mUoUMVyEeSRKL-M"))
				},
			},
			want: want{
				claims: &Claims{
					UserID: "testuserid",
					Name:   "testUser",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: &jwt.NumericDate{Time: time.Unix(4715839364, 0)},
					},
				},
			},
		},
		{
			name: "invalid token",
			args: args{
				ctxFunc: func() context.Context {
					return metadata.NewIncomingContext(
						context.Background(),
						metadata.Pairs("authorization", "bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQ"))
				},
			},
			want: want{
				errMsg: "invalid token",
			},
		},
		{
			name: "missing token",
			args: args{
				ctxFunc: func() context.Context {
					return metadata.NewIncomingContext(
						context.Background(),
						metadata.Pairs("qwe", "bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQ"))
				},
			},
			want: want{
				errMsg: "missing token",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := NewAuth(&Config{Secret: "superSecret"})
			require.NoError(t, err)
			require.NotNil(t, a)

			outCtx, err := a.GRPCAuthFunc(tt.args.ctxFunc())
			if tt.want.errMsg == "" {
				require.NoError(t, err)
				claims, ok := GetClaimsFromContext(outCtx)
				require.True(t, ok)
				require.Equal(t, tt.want.claims, claims)
			} else {
				require.Error(t, err)
				require.Nil(t, outCtx)
			}
		})
	}
}
