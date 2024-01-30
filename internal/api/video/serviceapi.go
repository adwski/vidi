//nolint:wrapcheck  // we use echo-style handler returns, i.e. return c.JSON(..)
package video

import (
	"errors"
	"fmt"
	"net/http"

	common "github.com/adwski/vidi/internal/api/model"
	"github.com/adwski/vidi/internal/api/user/auth"
	"github.com/adwski/vidi/internal/api/video/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func (svc *Service) getServiceSession(c echo.Context) error {
	claims, err := auth.GetClaimFromContext(c)
	if err != nil {
		svc.logger.Error("service auth fail", zap.Any("claims", claims))
		return c.JSON(http.StatusUnauthorized, &common.Response{
			Error: "unauthorized",
		})
	}
	logf := svc.logger.With(zap.String("id", claims.UserID),
		zap.String("name", claims.Name),
		zap.String("role", claims.Role))

	if claims.IsService() {
		svc.logger.Debug("service auth ok")
		return nil
	}
	logf.Error("service auth incorrect role")
	return c.JSON(http.StatusUnauthorized, &common.Response{
		Error: "unauthorized",
	})
}

func (svc *Service) updateVideoLocation(c echo.Context) error {
	if err := svc.getServiceSession(c); err != nil {
		return err
	}
	id := c.Param("id")
	location := c.Param("location")
	if len(location) == 0 {
		return c.JSON(http.StatusBadRequest, &common.Response{
			Error: "location cannot be empty",
		})
	}
	err := svc.s.UpdateLocation(c.Request().Context(), &model.Video{
		ID:       id,
		Location: location,
	})
	return svc.commonResponse(c, err)
}

func (svc *Service) updateVideoStatus(c echo.Context) error {
	if err := svc.getServiceSession(c); err != nil {
		return err
	}
	id := c.Param("id")
	status, err := model.GetStatusFromName(c.Param("status"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, &common.Response{
			Error: err.Error(),
		})
	}
	err = svc.s.UpdateStatus(c.Request().Context(), &model.Video{
		ID:     id,
		Status: status,
	})
	return svc.commonResponse(c, err)
}

func (svc *Service) updateVideo(c echo.Context) error {
	if err := svc.getServiceSession(c); err != nil {
		return err
	}
	vi, err := getVideoFromUpdateRequest(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &common.Response{
			Error: err.Error(),
		})
	}
	err = svc.s.Update(c.Request().Context(), vi)
	return svc.commonResponse(c, err)
}

func getVideoFromUpdateRequest(c echo.Context) (*model.Video, error) {
	var req model.UpdateRequest
	if err := c.Bind(&req); err != nil {
		return nil, errors.New("invalid params")
	}
	status, err := model.GetStatusFromName(req.Status)
	if err != nil {
		return nil, fmt.Errorf("invalid status: %w", err)
	}
	if len(req.Location) == 0 {
		return nil, errors.New("empty location")
	}
	return &model.Video{
		ID:       c.Param("id"),
		Location: req.Location,
		Status:   status,
	}, nil
}

func (svc *Service) listVideos(c echo.Context) error {
	if err := svc.getServiceSession(c); err != nil {
		return err
	}
	var req model.ListRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, &common.Response{
			Error: "incorrect params",
		})
	}
	videos, err := svc.s.GetListByStatus(c.Request().Context(), req.Status)
	if err != nil {
		return svc.erroredResponse(c, err)
	}
	return c.JSON(http.StatusOK, videos)
}
