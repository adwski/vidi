package video

import (
	"context"
	"errors"
	"testing"

	usermodel "github.com/adwski/vidi/internal/api/user/model"
	"github.com/adwski/vidi/internal/api/video/model"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestService_getVideosDBError(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	s := NewMockStore(t)
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})

	s.EXPECT().GetAll(mock.Anything, "qweqweqwe").Return(nil, errors.New("err"))
	u := usermodel.User{ID: "test"}

	videos, err := svc.GetVideos(context.TODO(), &u)
	require.ErrorIs(t, err, model.ErrStorage)
	require.Nil(t, videos)
}

func TestService_watchVideoDBError(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	s := NewMockStore(t)
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})

	vid := "test"
	userID := "test"
	s.EXPECT().Get(mock.Anything, vid, userID).Return(nil, errors.New("err"))
	u := usermodel.User{ID: userID}

	v, err := svc.WatchVideo(context.TODO(), &u, vid, false)
	require.ErrorIs(t, err, model.ErrStorage)
	require.Nil(t, v)
}

func TestService_deleteVideoDBError(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	s := NewMockStore(t)
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})

	vid := "test"
	userID := "test"
	s.EXPECT().Delete(mock.Anything, vid, userID).Return(errors.New("err"))
	u := usermodel.User{ID: userID}

	err = svc.DeleteVideo(context.TODO(), &u, vid)
	require.ErrorIs(t, err, model.ErrStorage)
}

func TestService_createVideoDBError(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	s := NewMockStore(t)
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})

	u := usermodel.User{ID: "test"}
	s.EXPECT().Create(mock.Anything, mock.Anything).Return(errors.New("err"))
	v, err := svc.CreateVideo(context.TODO(), &u, &model.CreateRequest{
		Name: "test",
		Size: 123,
		Parts: []*model.Part{{
			Num:      0,
			Size:     123,
			Status:   0,
			Checksum: "checksum",
		}},
	})
	require.Nil(t, v)
	require.ErrorIs(t, err, model.ErrStorage)
}
