package server

import (
	"errors"
	common "github.com/adwski/vidi/internal/api/model"
	httpmodel "github.com/adwski/vidi/internal/api/video/http"
	"github.com/adwski/vidi/internal/api/video/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
)

func (srv *Server) getQuota(c echo.Context) error {
	usr, err, ok := srv.getUser(c)
	if !ok {
		return err
	}
	quota, err := srv.videoSvc.GetQuotas(c.Request().Context(), usr)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, common.ResponseInternalError)
	}
	return c.JSON(http.StatusOK, quota)
}

func (srv *Server) getVideo(c echo.Context) error {
	usr, err, ok := srv.getUser(c)
	if !ok {
		return err
	}
	vide, err := srv.videoSvc.GetVideo(c.Request().Context(), usr,
		c.Param("id"), c.QueryParam("upload") == "resume")
	if err == nil {
		return c.JSON(http.StatusOK, httpmodel.NewVideoResponse(vide))
	}
	return srv.erroredResponse(c, err)
}

func (srv *Server) getVideos(c echo.Context) error {
	usr, err, ok := srv.getUser(c)
	if !ok {
		return err
	}
	videos, err := srv.videoSvc.GetVideos(c.Request().Context(), usr)
	if err != nil {
		return srv.erroredResponse(c, err)
	}
	resp := make([]*httpmodel.VideoResponse, 0, len(videos))
	for _, v := range videos {
		resp = append(resp, httpmodel.NewVideoResponse(v))
	}
	return c.JSON(http.StatusOK, resp)
}

func (srv *Server) watchVideo(c echo.Context) error {
	usr, err, ok := srv.getUser(c)
	if !ok {
		return err
	}

	generateURL := c.QueryParam("mode") == "url"
	resp, err := srv.videoSvc.WatchVideo(c.Request().Context(), usr, c.Param("id"), generateURL)
	switch {
	case err == nil:
		if generateURL {
			return c.JSON(http.StatusOK, httpmodel.WatchResponse{WatchURL: string(resp)})
		} else {
			return c.XMLBlob(http.StatusOK, resp)
		}
	case errors.Is(err, model.ErrNotFound):
		return c.JSON(http.StatusNotFound, &common.Response{
			Error: err.Error(),
		})
	case errors.Is(err, model.ErrNotReady),
		errors.Is(err, model.ErrState):
		return c.JSON(http.StatusMethodNotAllowed, &common.Response{
			Error: err.Error(),
		})
	default:
		srv.logger.Error("watchVideo failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, common.ResponseInternalError)
	}
}

func (srv *Server) deleteVideo(c echo.Context) error {
	usr, err, ok := srv.getUser(c)
	if !ok {
		return err
	}
	err = srv.videoSvc.DeleteVideo(c.Request().Context(), usr, c.Param("id"))
	if err != nil {
		return srv.erroredResponse(c, err)
	}
	return c.JSON(http.StatusOK, common.ResponseOK)
}

func (srv *Server) createVideo(c echo.Context) error {
	usr, err, ok := srv.getUser(c)
	if !ok {
		return err
	}
	var req model.CreateRequest
	if err = c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, common.ResponseIncorrectParams)
	}
	vide, err := srv.videoSvc.CreateVideo(c.Request().Context(), usr, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, common.ResponseInternalError)
	}
	return c.JSON(http.StatusCreated, httpmodel.NewVideoResponse(vide))
}

/*
func (srv *Server) updateVideoStatus(c echo.Context) error {
	if err := srv.serviceAuth(c); err != nil {
		return err
	}
	if err := srv.videoSvc.UpdateVideoStatus(
		c.Request().Context(),
		c.Param("id"),
		c.Param("status"),
	); err != nil {
		if errors.Is(err, model.ErrIncorrectStatusName) {
			return c.JSON(http.StatusBadRequest, common.ResponseIncorrectStatus)
		}
		return c.JSON(http.StatusInternalServerError, common.ResponseInternalError)
	}
	return c.JSON(http.StatusOK, common.ResponseOK)
}

func (srv *Server) updateVideo(c echo.Context) error {
	if err := srv.serviceAuth(c); err != nil {
		return err
	}
	if err := srv.videoSvc.UpdateVideoStatusAndLocation(
		c.Request().Context(),
		c.Param("id"),
		c.Param("status"),
		c.Param("location"),
	); err != nil {
		switch {
		case errors.Is(err, model.ErrIncorrectStatusName), errors.Is(err, model.ErrEmptyLocation):
			return c.JSON(http.StatusBadRequest, err.Error())
		default:
			return c.JSON(http.StatusInternalServerError, common.ResponseInternalError)
		}
	}
	return c.JSON(http.StatusOK, common.ResponseOK)
}

func (srv *Server) searchVideos(c echo.Context) error {
	if err := srv.serviceAuth(c); err != nil {
		return err
	}
	var req httpmodel.ListRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, common.ResponseIncorrectParams)
	}
	videos, err := srv.videoSvc.GetVideosByStatus(c.Request().Context(), req.Status)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrIncorrectStatusName), errors.Is(err, model.ErrEmptyLocation):
			return c.JSON(http.StatusBadRequest, err.Error())
		default:
			return c.JSON(http.StatusInternalServerError, common.ResponseInternalError)
		}
	}
	return c.JSON(http.StatusOK, videos)
}
*/
