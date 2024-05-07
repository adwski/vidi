//nolint:wrapcheck  // we use echo-style handler returns, i.e. return c.JSON(..)
package video

import (
	"context"
	"errors"
	user "github.com/adwski/vidi/internal/api/user/model"
	"github.com/adwski/vidi/internal/api/video/model"
	"github.com/adwski/vidi/internal/session"
)

const (
	defaultPartSize    = 10 * 1024 * 1024
	videoCreateRetries = 3
)

func (svc *Service) GetQuotas(ctx context.Context, usr *user.User) (*model.UserStats, error) {
	usage, err := svc.s.Usage(ctx, usr.ID)
	if err != nil {
		return nil, errors.Join(model.ErrStorage, err)
	}
	return &model.UserStats{
		VideosQuota: svc.quotas.VideosPerUser,
		VideosUsage: usage.Videos,
		SizeQuota:   svc.quotas.MaxTotalSize,
		SizeUsage:   usage.Size,
	}, nil
}

func (svc *Service) GetVideo(ctx context.Context, usr *user.User, vid string) (*model.Video, error) {
	video, err := svc.s.Get(ctx, vid, usr.ID)
	if err != nil {
		return nil, errors.Join(model.ErrStorage, err)
	}
	return video, nil
}

func (svc *Service) GetVideos(ctx context.Context, usr *user.User) ([]*model.Video, error) {
	videos, err := svc.s.GetAll(ctx, usr.ID)
	if err != nil {
		return nil, errors.Join(model.ErrStorage, err)
	}
	return videos, nil
}

func (svc *Service) WatchVideo(ctx context.Context, usr *user.User, vid string) (string, error) {
	video, err := svc.s.Get(ctx, vid, usr.ID)
	if err != nil {
		return "", errors.Join(model.ErrStorage, err)
	}
	if video.IsErrored() {
		return "", model.ErrState
	}
	if !video.IsReady() {
		return "", model.ErrNotReady
	}
	var sessID string
	sessID, err = svc.idGen.Get()
	if err != nil {
		return "", errors.Join(errors.New("cannot generate watch session id"), err)
	}
	sess := &session.Session{
		ID:       sessID,
		VideoID:  video.ID,
		Location: video.Location,
	}
	if err = svc.uploadSessions.Set(ctx, sess); err != nil {
		return "", errors.Join(model.ErrSessionStorage, err)
	}
	return svc.getWatchURL(sessID), nil
}

func (svc *Service) DeleteVideo(ctx context.Context, usr *user.User, vid string) error {
	err := svc.s.Delete(ctx, vid, usr.ID)
	if err != nil {
		return errors.Join(model.ErrStorage, err)
	}
	return nil
}

func (svc *Service) CreateVideo(ctx context.Context, usr *user.User, req *model.CreateRequest) (*model.Video, error) {
	var (
		err error
	)
	newVideo := model.NewVideoNoID(usr.ID, req.Name, req.Size)
	newVideo.UploadInfo = &model.UploadInfo{
		Parts: req.Parts,
	}

	for i := 1; ; i++ {
		newVideo.ID, err = svc.idGen.Get()
		if err != nil {
			return nil, errors.Join(errors.New("cannot generate video id"), err)
		}
		newVideo.Location, err = svc.idGen.Get()
		if err != nil {
			return nil, errors.Join(errors.New("cannot generate location id"), err)
		}
		if err = svc.s.Create(ctx, newVideo); err == nil {
			break
		}
		if errors.Is(err, model.ErrAlreadyExists) {
			if i < videoCreateRetries {
				continue
			}
			err = errors.Join(model.ErrGivenUp, err)
		}
		break
	}
	if err != nil {
		return nil, errors.Join(model.ErrStorage, err)
	}

	sess := &session.Session{
		ID:       newVideo.Location,
		VideoID:  newVideo.ID,
		PartSize: defaultPartSize,
	}
	if err = svc.uploadSessions.Set(ctx, sess); err != nil {
		return nil, errors.Join(model.ErrSessionStorage, err)
	}
	newVideo.UploadInfo.URL = svc.getUploadURL(sess.ID)
	return newVideo, nil
}
