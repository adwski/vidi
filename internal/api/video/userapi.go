//nolint:wrapcheck  // we use echo-style handler returns, i.e. return c.JSON(..)
package video

import (
	"context"
	"errors"
	user "github.com/adwski/vidi/internal/api/user/model"
	"github.com/adwski/vidi/internal/api/video/model"
	"github.com/adwski/vidi/internal/session"
	sessionStore "github.com/adwski/vidi/internal/session/store"
)

const (
	videoCreateRetries = 3
)

func (svc *Service) GetQuotas(ctx context.Context, usr *user.User) (*model.UserStats, error) {
	usage, err := svc.s.Usage(ctx, usr.ID)
	if err != nil {
		return nil, errors.Join(model.ErrStorage, err)
	}
	return &model.UserStats{
		VideosQuota: svc.Quotas.VideosPerUser,
		VideosUsage: usage.Videos,
		SizeQuota:   svc.Quotas.MaxTotalSize,
		SizeUsage:   usage.Size,
	}, nil
}

func (svc *Service) GetVideo(ctx context.Context, usr *user.User, vid string) (*model.Video, error) {
	// TODO handle upload URL for upload resuming
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
	sessID, err := svc.storeSessionAndReturnURL(ctx, video, svc.watchSessions)
	if err != nil {
		return "", err
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

func (svc *Service) CreateVideo(ctx context.Context, usr *user.User) (*model.Video, error) {
	var (
		vid string
		err error
	)
	newVideo := model.NewVideoNoID(usr.ID)

	for i := 1; ; i++ {
		vid, err = svc.idGen.Get()
		if err != nil {
			return nil, errors.Join(errors.New("cannot generate video id"), err)
		}
		newVideo.ID = vid
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

	sessID, err := svc.storeSessionAndReturnURL(ctx, newVideo, svc.uploadSessions)
	if err != nil {
		return nil, err
	}
	newVideo.UploadURL = svc.getUploadURL(sessID)
	return newVideo, nil
}

func (svc *Service) storeSessionAndReturnURL(
	ctx context.Context,
	vi *model.Video,
	sessStore *sessionStore.Store,
) (*session.Session, error) {
	sessID, err := svc.idGen.Get()
	if err != nil {
		return nil, errors.Join(model.ErrInternal, err)
	}
	sess := &session.Session{
		ID:       sessID,
		VideoID:  vi.ID,
		Location: vi.Location, // used for watch sessions
	}
	if err = sessStore.Set(ctx, sess); err != nil {
		return nil, errors.Join(model.ErrSessionStorage, err)
	}
	return sess, nil
}
