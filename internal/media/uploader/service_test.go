package uploader

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func Test_checkHeader(t *testing.T) {
	type args struct {
		cType string
		cLen  int
	}
	tests := []struct {
		name          string
		args          args
		contentLength int
		err           string
	}{
		{
			name: "valid headers",
			args: args{
				cType: "application/x-vidi-mediapart",
				cLen:  100,
			},
			contentLength: 100,
		},
		{
			name: "wrong content-length",
			args: args{
				cType: "application/x-vidi-mediapart",
				cLen:  0,
			},
			err: "wrong or missing content length",
		},
		{
			name: "wrong content-type",
			args: args{
				cType: "qweqwe/asdas",
				cLen:  100,
			},
			err: "wrong content type",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{
				Request: fasthttp.Request{Header: fasthttp.RequestHeader{}},
			}
			ctx.Request.Header.SetContentLength(tt.args.cLen)
			ctx.Request.Header.SetContentType(tt.args.cType)

			ln, err := checkHeader(ctx)
			if tt.err != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.contentLength, ln)
			}
		})
	}
}
