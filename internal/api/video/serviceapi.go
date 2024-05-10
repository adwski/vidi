//nolint:wrapcheck  // we use echo-style handler returns, i.e. return c.JSON(..)
package video

import (
	"context"
	"errors"
	"github.com/adwski/vidi/internal/api/video/model"
	"github.com/adwski/vidi/internal/mp4/meta"
	"github.com/vmihailenco/msgpack/v5"
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

func (svc *Service) UpdateVideoStatusAndMeta(ctx context.Context, vid string, status model.Status, pbMeta []byte) error {
	if err := model.ValidateStatus(status); err != nil {
		return err // passing ErrIncorrectStatusNum as is
	}
	if len(pbMeta) == 0 {
		return model.ErrEmptyPlaybackMeta
	}
	var playbackMeta meta.Meta
	if err := msgpack.Unmarshal(pbMeta, &playbackMeta); err != nil {
		return errors.Join(model.ErrInvalidPlaybackMeta, err)
	}
	if err := svc.s.Update(ctx, &model.Video{
		ID:           vid,
		Status:       status,
		PlaybackMeta: &playbackMeta,
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
