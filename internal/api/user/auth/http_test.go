package auth

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestGetClaimFromEchoContext(t *testing.T) {
	type args struct {
		ctxFunc func() echo.Context
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
			name: "success",
			args: args{
				ctxFunc: func() echo.Context {
					e := echo.New()
					ctx := e.NewContext(nil, nil)
					ctx.Set(sessionContextKey, &jwt.Token{Claims: &Claims{}})
					return ctx
				},
			},
		},
		{
			name: "no token",
			args: args{
				ctxFunc: func() echo.Context {
					e := echo.New()
					ctx := e.NewContext(nil, nil)
					ctx.Set(sessionContextKey, "tuytu")
					return ctx
				},
			},
			want: want{errMsg: "cannot get jwt token from session context"},
		},
		{
			name: "no claims",
			args: args{
				ctxFunc: func() echo.Context {
					e := echo.New()
					ctx := e.NewContext(nil, nil)
					ctx.Set(sessionContextKey, &jwt.Token{Claims: Claims{}})
					return ctx
				},
			},
			want: want{errMsg: "cannot get claims from jwt token"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := GetClaimFromEchoContext(tt.args.ctxFunc())
			if tt.want.errMsg != "" {
				assert.EqualError(t, err, tt.want.errMsg)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
			}
		})
	}
}
