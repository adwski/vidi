//nolint:wrapcheck  // we use echo-style handler returns, i.e. return c.JSON(..)
package video

import (
	"context"
	"errors"
	"github.com/adwski/vidi/internal/api/video/model"
)

func (svc *Service) UpdateVideoStatus(ctx context.Context, vid string, status model.Status) error {
	if err := model.ValidateStatus(status); err != nil {
		return err // passing ErrIncorrectStatusNum as is
	}
	if err := svc.s.UpdateStatus(ctx, &model.Video{
		ID:     vid,
		Status: status,
	}); err != nil {
		return errors.Join(model.ErrStorage, err)
	}
	if status == model.StatusReady {
		if err := svc.s.DeleteUploadedParts(ctx, vid); err != nil {
			return errors.Join(model.ErrStorage, err)
		}
	}
	return nil
}

func (svc *Service) UpdateVideoStatusAndLocation(ctx context.Context, vid, location string, status model.Status) error {
	if err := model.ValidateStatus(status); err != nil {
		return err // passing ErrIncorrectStatusNum as is
	}
	if len(location) == 0 {
		return model.ErrEmptyLocation
	}
	if err := svc.s.Update(ctx, &model.Video{
		ID:       vid,
		Status:   status,
		Location: location,
	}); err != nil {
		return errors.Join(model.ErrStorage, err)
	}
	return nil
}

func (svc *Service) GetVideosByStatus(ctx context.Context, status model.Status) ([]*model.Video, error) {
	if err := model.ValidateStatus(status); err != nil {
		return nil, err // passing ErrIncorrectStatusNum as is
	}
	videos, err := svc.s.GetListByStatus(ctx, status)
	if err != nil {
		return nil, errors.Join(model.ErrStorage, err)
	}
	return videos, nil
}

func (svc *Service) NotifyPartUpload(ctx context.Context, vid string, part *model.Part) error {
	if err := svc.s.UpdatePart(ctx, vid, part); err != nil {
		return errors.Join(model.ErrStorage, err)
	}
	return nil
}
