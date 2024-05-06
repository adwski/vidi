package auth

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ctxKey int

const (
	authScheme = "bearer"

	ctxKeyClaims ctxKey = iota
)

func (a *Auth) GRPCAuthFunc(ctx context.Context) (context.Context, error) {
	token, err := auth.AuthFromMD(ctx, authScheme)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}
	claims, err := a.parseToken(token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}
	return context.WithValue(ctx, ctxKeyClaims, claims), nil
}

func GetClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(ctxKeyClaims).(*Claims)
	return claims, ok
}
