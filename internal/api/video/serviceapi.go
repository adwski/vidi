//nolint:wrapcheck  // we use echo-style handler returns, i.e. return c.JSON(..)
package video

import (
	"context"
	"errors"
	"github.com/adwski/vidi/internal/api/video/model"
)

func (svc *Service) UpdateVideoStatus(ctx context.Context, vid, statusName string) error {
	status, err := model.GetStatusFromName(statusName)
	if err != nil {
		return err // passing ErrIncorrectStatusName as is
	}
	if err = svc.s.UpdateStatus(ctx, &model.Video{
		ID:     vid,
		Status: status,
	}); err != nil {
		return errors.Join(model.ErrStorage, err)
	}
	return nil
}

func (svc *Service) UpdateVideoStatusAndLocation(ctx context.Context, vid, statusName, location string) error {
	status, err := model.GetStatusFromName(statusName)
	if err != nil {
		return err // passing ErrIncorrectStatusName as is
	}
	if len(location) == 0 {
		return model.ErrEmptyLocation
	}
	if err = svc.s.Update(ctx, &model.Video{
		ID:       vid,
		Status:   status,
		Location: location,
	}); err != nil {
		return errors.Join(model.ErrStorage, err)
	}
	return nil
}

func (svc *Service) GetVideosByStatus(ctx context.Context, statusName string) ([]*model.Video, error) {
	status, err := model.GetStatusFromName(statusName)
	if err != nil {
		return nil, err // passing ErrIncorrectStatusName as is
	}
	videos, err := svc.s.GetListByStatus(ctx, status)
	if err != nil {
		return nil, errors.Join(model.ErrStorage, err)
	}
	return videos, nil
}
