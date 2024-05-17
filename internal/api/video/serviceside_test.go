package video

import (
	"context"
	"errors"
	"testing"

	"github.com/adwski/vidi/internal/api/video/model"
	"github.com/adwski/vidi/internal/mp4/meta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
)

func TestService_UpdateVideoStatus(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	vid := "test"
	status := model.StatusUploaded
	ctx := context.Background()
	s := NewMockStore(t)
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})
	s.EXPECT().UpdateStatus(ctx, mock.Anything).Run(func(_ context.Context, v *model.Video) {
		assert.Equal(t, vid, v.ID)
		assert.Equal(t, status, v.Status)
	}).Return(nil)

	err = svc.UpdateVideoStatus(ctx, vid, status)
	require.NoError(t, err)
}

func TestService_UpdateVideoStatusReady(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	vid := "test"
	status := model.StatusReady
	ctx := context.Background()
	s := NewMockStore(t)
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})
	s.EXPECT().UpdateStatus(ctx, mock.Anything).Run(func(_ context.Context, v *model.Video) {
		assert.Equal(t, vid, v.ID)
		assert.Equal(t, status, v.Status)
	}).Return(nil)
	s.EXPECT().DeleteUploadedParts(ctx, vid).Return(nil)

	err = svc.UpdateVideoStatus(ctx, vid, status)
	require.NoError(t, err)
}

func TestService_UpdateVideoStatusReadyDBErr(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	vid := "test"
	status := model.StatusReady
	ctx := context.Background()
	s := NewMockStore(t)
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})
	s.EXPECT().UpdateStatus(ctx, mock.Anything).Run(func(_ context.Context, v *model.Video) {
		assert.Equal(t, vid, v.ID)
		assert.Equal(t, status, v.Status)
	}).Return(nil)
	s.EXPECT().DeleteUploadedParts(ctx, vid).Return(errors.New("test"))

	err = svc.UpdateVideoStatus(ctx, vid, status)
	require.ErrorIs(t, err, model.ErrStorage)
}

func TestService_UpdateVideoStatusIncorrect(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	vid := "test"
	status := model.Status(123123123)
	ctx := context.Background()
	svc := NewService(&ServiceConfig{
		Logger: logger,
	})

	err = svc.UpdateVideoStatus(ctx, vid, status)
	require.ErrorIs(t, err, model.ErrIncorrectStatusNum)
}

func TestService_UpdateVideoStatusDBErr(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	vid := "test"
	status := model.StatusUploaded
	ctx := context.Background()
	s := NewMockStore(t)
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})
	s.EXPECT().UpdateStatus(ctx, mock.Anything).Run(func(_ context.Context, v *model.Video) {
		assert.Equal(t, vid, v.ID)
		assert.Equal(t, status, v.Status)
	}).Return(errors.New("test"))

	err = svc.UpdateVideoStatus(ctx, vid, status)
	require.ErrorIs(t, err, model.ErrStorage)
}

func TestService_UpdateVideoStatusAndMetaIncorrectStatus(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	vid := "test"
	status := model.Status(123123123)
	ctx := context.Background()
	svc := NewService(&ServiceConfig{
		Logger: logger,
	})

	err = svc.UpdateVideoStatusAndMeta(ctx, vid, status, []byte{})
	require.ErrorIs(t, err, model.ErrIncorrectStatusNum)
}

func TestService_UpdateVideoStatusAndMeta(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	vid := "test"
	status := model.StatusReady
	m := &meta.Meta{
		Tracks: []meta.Track{{
			Codec: &meta.Codec{
				Profile:    "prof1",
				SampleRate: 10000,
			},
			Segment: &meta.SegmentConfig{
				Init:        "init",
				StartNumber: 0,
				Duration:    100000,
				Timescale:   30000,
			},
			Name:     "vide",
			MimeType: "video/mp4",
		}},
		Duration: 100000,
	}
	b, errM := msgpack.Marshal(m)
	require.NoError(t, errM)

	ctx := context.Background()
	s := NewMockStore(t)
	s.EXPECT().Update(ctx, mock.Anything).Run(func(_ context.Context, v *model.Video) {
		assert.Equal(t, vid, v.ID)
		assert.Equal(t, status, v.Status)
		assert.Equal(t, m, v.PlaybackMeta)
	}).Return(nil)
	s.EXPECT().DeleteUploadedParts(ctx, vid).Return(nil)

	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})

	err = svc.UpdateVideoStatusAndMeta(ctx, vid, status, b)
	require.NoError(t, err)
}

func TestService_UpdateVideoStatusAndMetaDBErrParts(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	vid := "test"
	status := model.StatusReady
	m := &meta.Meta{
		Tracks: []meta.Track{{
			Codec: &meta.Codec{
				Profile:    "prof1",
				SampleRate: 10000,
			},
			Segment: &meta.SegmentConfig{
				Init:        "init",
				StartNumber: 0,
				Duration:    100000,
				Timescale:   30000,
			},
			Name:     "vide",
			MimeType: "video/mp4",
		}},
		Duration: 100000,
	}
	b, errM := msgpack.Marshal(m)
	require.NoError(t, errM)

	ctx := context.Background()
	s := NewMockStore(t)
	s.EXPECT().Update(ctx, mock.Anything).Run(func(_ context.Context, v *model.Video) {
		assert.Equal(t, vid, v.ID)
		assert.Equal(t, status, v.Status)
		assert.Equal(t, m, v.PlaybackMeta)
	}).Return(nil)
	s.EXPECT().DeleteUploadedParts(ctx, vid).Return(errors.New("test"))

	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})

	err = svc.UpdateVideoStatusAndMeta(ctx, vid, status, b)
	require.ErrorIs(t, err, model.ErrStorage)
}

func TestService_UpdateVideoStatusAndMetaDBErr(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	vid := "test"
	status := model.StatusReady
	m := &meta.Meta{
		Tracks: []meta.Track{{
			Codec: &meta.Codec{
				Profile:    "prof1",
				SampleRate: 10000,
			},
			Segment: &meta.SegmentConfig{
				Init:        "init",
				StartNumber: 0,
				Duration:    100000,
				Timescale:   30000,
			},
			Name:     "vide",
			MimeType: "video/mp4",
		}},
		Duration: 100000,
	}
	b, errM := msgpack.Marshal(m)
	require.NoError(t, errM)

	ctx := context.Background()
	s := NewMockStore(t)
	s.EXPECT().Update(ctx, mock.Anything).Run(func(_ context.Context, v *model.Video) {
		assert.Equal(t, vid, v.ID)
		assert.Equal(t, status, v.Status)
		assert.Equal(t, m, v.PlaybackMeta)
	}).Return(errors.New("test"))

	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})

	err = svc.UpdateVideoStatusAndMeta(ctx, vid, status, b)
	require.ErrorIs(t, err, model.ErrStorage)
}

func TestService_UpdateVideoStatusAndMetaIncorrectMeta(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	vid := "test"
	status := model.StatusReady

	ctx := context.Background()

	svc := NewService(&ServiceConfig{
		Logger: logger,
	})

	err = svc.UpdateVideoStatusAndMeta(ctx, vid, status, []byte("qweqweqwe"))
	require.ErrorIs(t, err, model.ErrInvalidPlaybackMeta)
}

func TestService_UpdateVideoStatusAndMetaEmptyMeta(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	vid := "test"
	status := model.StatusReady

	ctx := context.Background()

	svc := NewService(&ServiceConfig{
		Logger: logger,
	})

	err = svc.UpdateVideoStatusAndMeta(ctx, vid, status, nil)
	require.ErrorIs(t, err, model.ErrEmptyPlaybackMeta)
}

func TestService_GetVideosByStatusIncorrectStatus(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	status := model.Status(123123123)
	ctx := context.Background()
	svc := NewService(&ServiceConfig{
		Logger: logger,
	})

	vi, err := svc.GetVideosByStatus(ctx, status)
	require.ErrorIs(t, err, model.ErrIncorrectStatusNum)
	require.Nil(t, vi)
}

func TestService_GetVideosByStatus(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	status := model.StatusUploaded
	ctx := context.Background()
	s := NewMockStore(t)
	s.EXPECT().GetListByStatus(ctx, status).Return([]*model.Video{{ID: "vid"}}, nil)
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})

	vi, err := svc.GetVideosByStatus(ctx, status)
	require.NoError(t, err)
	require.NotNil(t, vi)
	assert.Equal(t, "vid", vi[0].ID)
}

func TestService_GetVideosByStatusDBErr(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	status := model.StatusUploaded
	ctx := context.Background()
	s := NewMockStore(t)
	s.EXPECT().GetListByStatus(ctx, status).Return(nil, errors.New("test"))
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})

	vi, err := svc.GetVideosByStatus(ctx, status)
	require.ErrorIs(t, err, model.ErrStorage)
	require.Nil(t, vi)
}

func TestService_NotifyPartUpload(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	ctx := context.Background()
	vid := "test"
	part := &model.Part{
		Checksum: "qwe123",
		Num:      0,
		Size:     123123,
		Status:   model.PartStatusOK,
	}
	s := NewMockStore(t)
	s.EXPECT().UpdatePart(ctx, vid, part).Return(nil)
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})

	err = svc.NotifyPartUpload(ctx, vid, part)
	require.NoError(t, err)
}

func TestService_NotifyPartUploadDBErr(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	ctx := context.Background()
	vid := "test"
	part := &model.Part{
		Checksum: "qwe123",
		Num:      0,
		Size:     123123,
		Status:   model.PartStatusOK,
	}
	s := NewMockStore(t)
	s.EXPECT().UpdatePart(ctx, vid, part).Return(errors.New("test"))
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})

	err = svc.NotifyPartUpload(ctx, vid, part)
	require.ErrorIs(t, err, model.ErrStorage)
}
