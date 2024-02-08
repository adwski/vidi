//nolint:dupl //similar test cases
package video

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adwski/vidi/internal/api/user/auth"
	"github.com/adwski/vidi/internal/api/video/model"
	"github.com/adwski/vidi/internal/generators"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestService_getVideoNoSession(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	a, err := auth.NewAuth(&auth.Config{
		Secret: "qweqeqwe",
	})
	require.NoError(t, err)
	svc := Service{
		logger: logger,
		auth:   a,
	}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	ctx := echo.New().NewContext(r, w)

	err = svc.getVideo(ctx)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, ctx.Response().Status)
}

func TestService_getVideosNoSession(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	a, err := auth.NewAuth(&auth.Config{
		Secret: "qweqeqwe",
	})
	require.NoError(t, err)
	svc := Service{
		logger: logger,
		auth:   a,
		s:      NewMockStore(t),
	}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	ctx := echo.New().NewContext(r, w)

	err = svc.getVideos(ctx)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, ctx.Response().Status)
}

func TestService_getVideosError(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	a, err := auth.NewAuth(&auth.Config{
		Secret: "qweqeqwe",
	})
	require.NoError(t, err)

	s := NewMockStore(t)
	svc := Service{
		logger: logger,
		auth:   a,
		s:      s,
	}

	s.EXPECT().GetAll(mock.Anything, "qweqweqwe").Return(nil, errors.New("err"))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	ctx := echo.New().NewContext(r, w)
	ctx.Set("vUser", &jwt.Token{Claims: &auth.Claims{
		UserID: "qweqweqwe",
	}})

	err = svc.getVideos(ctx)
	require.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, ctx.Response().Status)
}

func TestService_watchVideoNoSession(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	a, err := auth.NewAuth(&auth.Config{
		Secret: "qweqeqwe",
	})
	require.NoError(t, err)

	svc := Service{
		logger: logger,
		auth:   a,
	}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	ctx := echo.New().NewContext(r, w)

	err = svc.watchVideo(ctx)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, ctx.Response().Status)
}

func TestService_watchVideoDBError(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	a, err := auth.NewAuth(&auth.Config{
		Secret: "qweqeqwe",
	})
	require.NoError(t, err)

	s := NewMockStore(t)
	svc := Service{
		logger: logger,
		auth:   a,
		s:      s,
	}

	s.EXPECT().Get(mock.Anything, mock.Anything, "qweqweqwe").Return(nil, errors.New("err"))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	ctx := echo.New().NewContext(r, w)
	ctx.Set("vUser", &jwt.Token{Claims: &auth.Claims{
		UserID: "qweqweqwe",
	}})

	err = svc.watchVideo(ctx)
	require.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, ctx.Response().Status)
}

func TestService_watchVideoError(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	a, err := auth.NewAuth(&auth.Config{
		Secret: "qweqeqwe",
	})
	require.NoError(t, err)

	s := NewMockStore(t)
	svc := Service{
		logger: logger,
		auth:   a,
		s:      s,
	}

	s.EXPECT().Get(mock.Anything, mock.Anything, "qweqweqwe").Return(&model.Video{
		Status: model.StatusError,
	}, nil)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	ctx := echo.New().NewContext(r, w)
	ctx.Set("vUser", &jwt.Token{Claims: &auth.Claims{
		UserID: "qweqweqwe",
	}})

	err = svc.watchVideo(ctx)
	require.NoError(t, err)

	assert.Equal(t, http.StatusMethodNotAllowed, ctx.Response().Status)
}

func TestService_deleteVideoNoSession(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	a, err := auth.NewAuth(&auth.Config{
		Secret: "qweqeqwe",
	})
	require.NoError(t, err)

	svc := Service{
		logger: logger,
		auth:   a,
	}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	ctx := echo.New().NewContext(r, w)

	err = svc.deleteVideo(ctx)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, ctx.Response().Status)
}

func TestService_deleteVideoDBError(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	a, err := auth.NewAuth(&auth.Config{
		Secret: "qweqeqwe",
	})
	require.NoError(t, err)

	s := NewMockStore(t)
	svc := Service{
		logger: logger,
		auth:   a,
		s:      s,
	}

	s.EXPECT().Delete(mock.Anything, mock.Anything, "qweqweqwe").Return(errors.New("err"))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	ctx := echo.New().NewContext(r, w)
	ctx.Set("vUser", &jwt.Token{Claims: &auth.Claims{
		UserID: "qweqweqwe",
	}})

	err = svc.deleteVideo(ctx)
	require.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, ctx.Response().Status)
}

func TestService_createVideoNoSession(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	a, err := auth.NewAuth(&auth.Config{
		Secret: "qweqeqwe",
	})
	require.NoError(t, err)

	svc := Service{
		logger: logger,
		auth:   a,
	}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	ctx := echo.New().NewContext(r, w)

	err = svc.createVideo(ctx)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, ctx.Response().Status)
}

func TestService_createVideoDBError(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	a, err := auth.NewAuth(&auth.Config{
		Secret: "qweqeqwe",
	})
	require.NoError(t, err)

	s := NewMockStore(t)
	svc := Service{
		logger: logger,
		auth:   a,
		s:      s,
		idGen:  generators.NewID(),
	}

	s.EXPECT().Create(mock.Anything, mock.Anything).Return(errors.New("err"))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	ctx := echo.New().NewContext(r, w)
	ctx.Set("vUser", &jwt.Token{Claims: &auth.Claims{
		UserID: "qweqweqwe",
	}})

	err = svc.createVideo(ctx)
	require.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, ctx.Response().Status)
}
