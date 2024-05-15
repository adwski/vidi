package requestid

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"testing"
)

func TestGenerator_InterceptorFunc(t *testing.T) {
	type args struct {
		trust bool
		reqID string
		empty bool
	}
	type want struct {
		reqID string
		rand  bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "trust request id",
			args: args{
				reqID: "qweqwe",
				trust: true,
			},
			want: want{
				reqID: "qweqwe",
			},
		},
		{
			name: "trust but no incoming req id",
			args: args{
				trust: true,
			},
			want: want{
				rand: true,
			},
		},
		{
			name: "trust but empty req id",
			args: args{
				trust: true,
				empty: true,
			},
			want: want{
				rand: true,
			},
		},
		{
			name: "do not trust request id",
			args: args{
				reqID: "qweqwe",
				trust: false,
			},
			want: want{
				rand: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := zap.NewDevelopment()
			require.NoError(t, err)

			reqID := New(logger, tt.args.trust)
			require.NotNil(t, reqID)

			interceptor := reqID.InterceptorFunc()
			require.NotNil(t, interceptor)

			var ctx context.Context
			if tt.args.reqID != "" {
				ctx = metadata.NewIncomingContext(context.Background(), metadata.Pairs(xRequestIDField, tt.args.reqID))
			} else if tt.args.empty {
				ctx = metadata.NewIncomingContext(context.Background(), metadata.Pairs(xRequestIDField, ""))
			} else {
				ctx = context.Background()
			}

			handler := func(ctx context.Context, req any) (any, error) {
				rID, ok := ctx.Value(ctxKeyXRequestID).(string)
				require.True(t, ok)
				if !tt.want.rand {
					assert.Equal(t, tt.want.reqID, rID)
				} else {
					assert.NotEqual(t, tt.want.reqID, rID)
					assert.NotEmpty(t, rID)
				}
				return req, nil
			}

			_, err = interceptor(ctx, nil, nil, handler)
			require.NoError(t, err)
		})
	}
}
