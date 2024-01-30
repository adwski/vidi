//nolint:wrapcheck  // we use echo-style handler returns, i.e. return c.JSON(..)
package video

import (
	"errors"
	"net/http"

	"github.com/adwski/vidi/internal/session"
	sessionStore "github.com/adwski/vidi/internal/session/store"

	common "github.com/adwski/vidi/internal/api/model"
	"github.com/adwski/vidi/internal/api/user/auth"
	user "github.com/adwski/vidi/internal/api/user/model"
	"github.com/adwski/vidi/internal/api/video/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func (svc *Service) getUserSession(c echo.Context) (*user.User, error, bool) {
	claims, err := auth.GetClaimFromContext(c)
	if err != nil {
		svc.logger.Debug("cannot get user session", zap.Error(err))
		return nil, c.JSON(http.StatusUnauthorized, &common.Response{
			Error: "unauthorized",
		}), false
	}
	return &user.User{
		ID:   claims.UserID,
		Name: claims.Name,
	}, nil, true
}

func (svc *Service) getVideo(c echo.Context) error {
	u, err, ok := svc.getUserSession(c)
	if !ok {
		return err
	}
	video, errV := svc.s.Get(c.Request().Context(), c.Param("id"), u.ID)
	if errV == nil {
		return c.JSON(http.StatusOK, video.Response())
	}
	return svc.erroredResponse(c, errV)
}

func (svc *Service) getVideos(c echo.Context) error {
	u, err, ok := svc.getUserSession(c)
	if !ok {
		return err
	}
	videos, errV := svc.s.GetAll(c.Request().Context(), u.ID)
	if errV != nil {
		return svc.erroredResponse(c, errV)
	}
	resp := make([]*model.Response, 0, len(videos))
	for _, v := range videos {
		resp = append(resp, v.Response())
	}
	return c.JSON(http.StatusOK, resp)
}

func (svc *Service) watchVideo(c echo.Context) error {
	u, err, ok := svc.getUserSession(c)
	if !ok {
		return err
	}
	video, errV := svc.s.Get(c.Request().Context(), c.Param("id"), u.ID)
	if errV != nil {
		return svc.erroredResponse(c, errV)
	}

	if video.IsErrored() {
		return c.JSON(http.StatusOK, &common.Response{
			Error: "video cannot be watched",
		})
	}

	if !video.IsReady() {
		return c.JSON(http.StatusOK, &common.Response{
			Message: "video is not ready",
		})
	}

	if sessID, sessOK := svc.storeSessionAndReturnURL(c, video, svc.watchSessions); sessOK {
		return c.JSON(http.StatusAccepted, &model.WatchResponse{WatchURL: svc.getWatchURL(sessID)})
	}
	return c.JSON(http.StatusInternalServerError, &common.Response{
		Error: common.InternalError,
	})
}

func (svc *Service) deleteVideo(c echo.Context) error {
	u, err, ok := svc.getUserSession(c)
	if !ok {
		return err
	}
	err = svc.s.Delete(c.Request().Context(), c.Param("id"), u.ID)
	if err == nil {
		return c.JSON(http.StatusOK, &common.Response{
			Message: "ok",
		})
	}
	return svc.erroredResponse(c, err)
}

func (svc *Service) createVideo(c echo.Context) error {
	u, err, ok := svc.getUserSession(c)
	if !ok {
		return err
	}
	newID, errID := svc.idGen.Get()
	if errID != nil {
		svc.logger.Error("cannot generate new video id", zap.Error(errID))
		return c.JSON(http.StatusNotFound, &common.Response{
			Error: common.InternalError,
		})
	}

	newVideo := model.NewVideo(newID, u.ID)
	err = svc.s.Create(c.Request().Context(), newVideo)
	if err == nil {
		if sess, sessOK := svc.storeSessionAndReturnURL(c, newVideo, svc.uploadSessions); sessOK {
			return c.JSON(http.StatusCreated, newVideo.UploadResponse(svc.getUploadURL(sess)))
		}
		return c.JSON(http.StatusInternalServerError, &common.Response{
			Error: common.InternalError,
		})
	}
	switch {
	case errors.Is(err, model.ErrAlreadyExists):
		svc.logger.Error("video with generated id already exists")
	default:
		svc.logger.Error("cannot create video in storage", zap.Error(err))
	}
	return c.JSON(http.StatusInternalServerError, &common.Response{
		Error: common.InternalError,
	})
}

func (svc *Service) storeSessionAndReturnURL(
	c echo.Context,
	vi *model.Video,
	sessStore *sessionStore.Store,
) (*session.Session, bool) {
	sessID, errSess := svc.idGen.Get()
	if errSess != nil {
		svc.logger.Error("cannot generate session id",
			zap.String("type", sessStore.Name()),
			zap.Error(errSess))
		return nil, false
	}
	sess := &session.Session{
		ID:       sessID,
		VideoID:  vi.ID,
		Location: vi.Location, // used for watch sessions
	}
	errSess = sessStore.Set(c.Request().Context(), sess)
	if errSess != nil {
		svc.logger.Error("cannot store session",
			zap.String("type", sessStore.Name()),
			zap.Error(errSess))
		return nil, false
	}
	svc.logger.Debug("session stored",
		zap.String("type", sessStore.Name()),
		zap.String("id", sess.ID))
	return sess, true
}
