// Package requestid implements request ID generator.
//
// Request IDs are generated using UUIDv4.
//
// Generator also implements GRPC server interceptor which
// has option to trust incoming request id or generate new one.
// Interceptor stores request id in context's values.
//
//nolint:wrapcheck // return grpc errors
package requestid

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	gstatus "google.golang.org/grpc/status"
)

type ctxKey int

const (
	xRequestIDField = "x-request-id"

	ctxKeyXRequestID ctxKey = iota

	cannotGenerateXRequestID = "#cannot generate request ID#"
)

// GetRequestIDFromContext retrieves request ID from request context.
func GetRequestIDFromContext(ctx context.Context) string {
	if reqID, ok := ctx.Value(ctxKeyXRequestID).(string); ok {
		return reqID
	}
	return ""
}

// Generator is request-id injector. It can work as GRPC unary interceptor
// or just provide generation func to be used externally.
type Generator struct {
	logger *zap.Logger
	gen    uuid.Generator
	trust  bool
}

// New creates request-id interceptor.
func New(logger *zap.Logger, trustRequestID bool) *Generator {
	return &Generator{
		gen:    uuid.NewGen(),
		logger: logger.With(zap.String("component", "request-id-generator")),
		trust:  trustRequestID,
	}
}

func (g *Generator) GenFunc() func() string {
	return func() string {
		uuidV4, err := g.gen.NewV4()
		if err != nil {
			g.logger.Error("cannot generate request id", zap.Error(err))
			return cannotGenerateXRequestID
		}
		return uuidV4.String()
	}
}

// InterceptorFunc returns UnaryServerInterceptor func that can be used by GRPC server.
func (g *Generator) InterceptorFunc() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		var newCtx context.Context
		if g.trust {
			reqID, ok := getXRequestIDFromMetadata(ctx)
			if ok {
				newCtx = context.WithValue(ctx, ctxKeyXRequestID, reqID)
			}
			g.logger.Debug("incoming request without id but trust is enabled")
		}

		if newCtx == nil {
			uuidV4, err := g.gen.NewV4()
			if err != nil {
				g.logger.Error("cannot generate request id", zap.Error(err))
				return nil, gstatus.Error(codes.Internal, "cannot generate request id")
			}
			newCtx = context.WithValue(ctx, ctxKeyXRequestID, uuidV4.String())
		}

		return handler(newCtx, req)
	}
}

func getXRequestIDFromMetadata(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}
	reqID := md.Get(xRequestIDField)
	if len(reqID) == 0 {
		return "", false
	}
	return reqID[0], true
}
