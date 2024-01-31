package video

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/adwski/vidi/internal/api/user/auth"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestService_getServiceSessionNoAuth(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	svc, err := NewService(&ServiceConfig{
		Logger:          logger,
		APIPrefix:       "/api/video",
		WatchURLPrefix:  "/watch",
		UploadURLPrefix: "/upload",
		AuthConfig: auth.Config{
			Secret:     "qweqweqwe",
			Domain:     "domain.com",
			HTTPS:      false,
			Expiration: time.Hour,
		},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	err = svc.getServiceSession(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, c.Response().Status)
}
