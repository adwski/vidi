//nolint:dupl // similar test cases
package video

import (
	"context"
	"errors"
	"fmt"
	"testing"

	usermodel "github.com/adwski/vidi/internal/api/user/model"
	"github.com/adwski/vidi/internal/api/video/model"
	"github.com/adwski/vidi/internal/mp4/meta"
	"github.com/adwski/vidi/internal/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestService_GetQuotas(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	ctx := context.Background()
	s := NewMockStore(t)
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
		Quotas: Quotas{
			VideosPerUser: 10,
			MaxTotalSize:  20,
		},
	})
	u := usermodel.User{ID: "test"}
	s.EXPECT().Usage(ctx, u.ID).Return(&model.UserUsage{
		Videos: 7,
		Size:   13,
	}, nil)
	quotaUsage, err := svc.GetQuotas(ctx, &u)
	require.NoError(t, err)
	require.Equal(t, uint(10), quotaUsage.VideosQuota)
	require.Equal(t, uint(7), quotaUsage.VideosUsage)
	require.Equal(t, uint64(20), quotaUsage.SizeQuota)
	require.Equal(t, uint64(13), quotaUsage.SizeUsage)
}

func TestService_GetQuotasDBErr(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	ctx := context.Background()
	s := NewMockStore(t)
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})
	u := usermodel.User{ID: "test"}
	s.EXPECT().Usage(ctx, u.ID).Return(nil, model.ErrStorage)
	quotaUsage, err := svc.GetQuotas(ctx, &u)
	require.Error(t, err)
	require.Nil(t, quotaUsage)
}

func TestService_GetVideo(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	ctx := context.Background()
	s := NewMockStore(t)
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})
	vid := "testvid"
	u := &usermodel.User{ID: "test"}
	v := &model.Video{
		ID: "testvid",
	}
	s.EXPECT().Get(ctx, vid, u.ID).Return(v, nil)

	vi, err := svc.GetVideo(ctx, u, vid, false)
	require.NoError(t, err)
	require.Equal(t, v.ID, vi.ID)
}

func TestService_GetVideoDBErr(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	ctx := context.Background()
	s := NewMockStore(t)
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})
	vid := "testvid"
	u := &usermodel.User{ID: "test"}
	s.EXPECT().Get(ctx, vid, u.ID).Return(nil, model.ErrStorage)

	vi, err := svc.GetVideo(ctx, u, vid, false)
	require.ErrorIs(t, err, model.ErrStorage)
	require.Nil(t, vi)
}

func TestService_GetVideoResumeNotResumable(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	ctx := context.Background()
	s := NewMockStore(t)
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})
	vid := "testvid"
	u := &usermodel.User{ID: "test"}
	v := &model.Video{
		ID:     "testvid",
		Status: model.StatusReady,
	}
	s.EXPECT().Get(ctx, vid, u.ID).Return(v, nil)

	vi, err := svc.GetVideo(ctx, u, vid, true)
	require.ErrorIs(t, err, model.ErrNotResumable)
	require.Nil(t, vi)
}

func TestService_GetVideoResume(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	ctx := context.Background()
	s := NewMockStore(t)
	ss := NewMockSessionStore(t)
	svc := NewService(&ServiceConfig{
		Logger:             logger,
		Store:              s,
		UploadSessionStore: ss,
		UploadURLPrefix:    "http://test",
	})
	vid := "testvid"
	u := &usermodel.User{ID: "test"}
	v := &model.Video{
		ID:         "testvid",
		Location:   "testloc",
		Status:     model.StatusCreated,
		UploadInfo: &model.UploadInfo{},
	}
	s.EXPECT().Get(ctx, vid, u.ID).Return(v, nil)
	ss.EXPECT().Set(ctx, mock.Anything).Run(func(_ context.Context, s *session.Session) {
		assert.Equal(t, v.Location, s.ID)
		assert.Equal(t, v.ID, s.VideoID)
	}).Return(nil)

	vi, err := svc.GetVideo(ctx, u, vid, true)
	require.NoError(t, err)
	assert.Equal(t, v.ID, vi.ID)
	assert.Equal(t, "http://test/"+v.Location, vi.UploadInfo.URL)
}

func TestService_GetVideoResumeSessErr(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	ctx := context.Background()
	s := NewMockStore(t)
	ss := NewMockSessionStore(t)
	svc := NewService(&ServiceConfig{
		Logger:             logger,
		Store:              s,
		UploadSessionStore: ss,
		UploadURLPrefix:    "http://test",
	})
	vid := "testvid"
	u := &usermodel.User{ID: "test"}
	v := &model.Video{
		ID:         "testvid",
		Location:   "testloc",
		Status:     model.StatusCreated,
		UploadInfo: &model.UploadInfo{},
	}
	s.EXPECT().Get(ctx, vid, u.ID).Return(v, nil)
	ss.EXPECT().Set(ctx, mock.Anything).Return(errors.New("test"))

	vi, err := svc.GetVideo(ctx, u, vid, true)
	require.ErrorIs(t, err, model.ErrSessionStorage)
	assert.Nil(t, vi)
}

func TestService_GetVideos(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	ctx := context.Background()
	s := NewMockStore(t)
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})
	u := &usermodel.User{ID: "test"}
	vi := []*model.Video{{ID: "testvid"}}
	s.EXPECT().GetAll(ctx, u.ID).Return(vi, nil)

	videos, err := svc.GetVideos(ctx, u)
	require.NoError(t, err)
	require.NotEmpty(t, videos)
	require.Equal(t, vi[0].ID, videos[0].ID)
}

func TestService_GetVideosDBError(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	ctx := context.Background()
	s := NewMockStore(t)
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})
	u := &usermodel.User{ID: "test"}
	s.EXPECT().GetAll(ctx, u.ID).Return(nil, errors.New("err"))

	videos, err := svc.GetVideos(ctx, u)
	require.ErrorIs(t, err, model.ErrStorage)
	require.Nil(t, videos)
}

func TestService_WatchVideoErrStatus(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	ctx := context.Background()
	s := NewMockStore(t)
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})

	v := &model.Video{
		ID:     "testvid",
		Status: model.StatusError,
	}
	u := &usermodel.User{ID: "test"}
	s.EXPECT().Get(ctx, v.ID, u.ID).Return(v, nil)

	b, err := svc.WatchVideo(ctx, u, v.ID, false)
	require.ErrorIs(t, err, model.ErrState)
	require.Nil(t, b)
}

func TestService_WatchVideoStatusNotReady(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	ctx := context.Background()
	s := NewMockStore(t)
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})

	v := &model.Video{
		ID:     "testvid",
		Status: model.StatusUploaded,
	}
	u := &usermodel.User{ID: "test"}
	s.EXPECT().Get(ctx, v.ID, u.ID).Return(v, nil)

	b, err := svc.WatchVideo(ctx, u, v.ID, false)
	require.ErrorIs(t, err, model.ErrNotReady)
	require.Nil(t, b)
}

func TestService_WatchVideoGenURL(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	ctx := context.Background()
	s := NewMockStore(t)
	ss := NewMockSessionStore(t)
	svc := NewService(&ServiceConfig{
		Logger:            logger,
		Store:             s,
		WatchSessionStore: ss,
		WatchURLPrefix:    "http://test",
	})

	var sessID string
	v := &model.Video{
		ID:       "testvid",
		Location: "testloc",
		Status:   model.StatusReady,
	}
	u := &usermodel.User{ID: "test"}
	s.EXPECT().Get(ctx, v.ID, u.ID).Return(v, nil)
	ss.EXPECT().Set(ctx, mock.Anything).Run(func(_ context.Context, sess *session.Session) {
		sessID = sess.ID
		assert.Equal(t, v.Location, sess.Location)
		assert.Equal(t, v.ID, sess.VideoID)
	}).Return(nil)

	b, err := svc.WatchVideo(ctx, u, v.ID, true)
	require.NoError(t, err)
	require.Equal(t, "http://test/"+sessID+"/manifest.mpd", string(b))
}

func TestService_WatchVideoStaticMPD(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	ctx := context.Background()
	s := NewMockStore(t)
	ss := NewMockSessionStore(t)
	svc := NewService(&ServiceConfig{
		Logger:            logger,
		Store:             s,
		WatchSessionStore: ss,
		WatchURLPrefix:    "http://test",
	})

	var (
		sessID string
		//nolint:lll // test data
		mpdTempl = `<MPD xmlns="urn:mpeg:dash:schema:mpd:2011" profiles="urn:mpeg:dash:profile:isoff-on-demand:2011" type="static" mediaPresentationDuration="PT100000�S">
  <BaseURL>%s</BaseURL>
  <Period id="p0">
    <AdaptationSet mimeType="video/mp4">
      <SegmentTemplate media="$RepresentationID$_$Number$.m4s" initialization="$RepresentationID$_init" duration="100000" startNumber="0" timescale="30000"></SegmentTemplate>
      <Representation id="vide" bandwidth="0" audioSamplingRate="10000" codecs="prof1"></Representation>
    </AdaptationSet>
  </Period>
</MPD>`
	)
	v := &model.Video{
		ID:       "testvid",
		Location: "testloc",
		Status:   model.StatusReady,
		PlaybackMeta: &meta.Meta{
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
		},
	}
	u := &usermodel.User{ID: "test"}
	s.EXPECT().Get(ctx, v.ID, u.ID).Return(v, nil)
	ss.EXPECT().Set(ctx, mock.Anything).Run(func(_ context.Context, sess *session.Session) {
		sessID = sess.ID
		assert.Equal(t, v.Location, sess.Location)
		assert.Equal(t, v.ID, sess.VideoID)
	}).Return(nil)

	b, err := svc.WatchVideo(ctx, u, v.ID, false)
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(mpdTempl, "http://test/"+sessID+"/"), string(b))
}

func TestService_WatchVideoSessErr(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	ctx := context.Background()
	s := NewMockStore(t)
	ss := NewMockSessionStore(t)
	svc := NewService(&ServiceConfig{
		Logger:            logger,
		Store:             s,
		WatchSessionStore: ss,
		WatchURLPrefix:    "http://test",
	})

	v := &model.Video{
		ID:       "testvid",
		Location: "testloc",
		Status:   model.StatusReady,
	}
	u := &usermodel.User{ID: "test"}
	s.EXPECT().Get(ctx, v.ID, u.ID).Return(v, nil)
	ss.EXPECT().Set(ctx, mock.Anything).Run(func(_ context.Context, sess *session.Session) {
		assert.Equal(t, v.Location, sess.Location)
		assert.Equal(t, v.ID, sess.VideoID)
	}).Return(errors.New("test"))

	b, err := svc.WatchVideo(ctx, u, v.ID, false)
	require.ErrorIs(t, err, model.ErrSessionStorage)
	assert.Nil(t, b)
}

func TestService_WatchVideoDBError(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	ctx := context.Background()
	s := NewMockStore(t)
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})

	vid := "test"
	u := &usermodel.User{ID: "test"}
	s.EXPECT().Get(ctx, vid, u.ID).Return(nil, errors.New("err"))

	v, err := svc.WatchVideo(ctx, u, vid, false)
	require.ErrorIs(t, err, model.ErrStorage)
	require.Nil(t, v)
}

func TestService_DeleteVideoDBError(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	ctx := context.Background()
	s := NewMockStore(t)
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})

	vid := "test"
	userID := "test"
	s.EXPECT().Delete(ctx, vid, userID).Return(errors.New("err"))
	u := usermodel.User{ID: userID}

	err = svc.DeleteVideo(ctx, &u, vid)
	require.ErrorIs(t, err, model.ErrStorage)
}
func TestService_DeleteVideo(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	ctx := context.Background()
	s := NewMockStore(t)
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})

	vid := "test"
	userID := "test"
	s.EXPECT().Delete(ctx, vid, userID).Return(nil)
	u := usermodel.User{ID: userID}
	err = svc.DeleteVideo(ctx, &u, vid)
	require.NoError(t, err)
}

func TestService_CreateVideoDBError(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	ctx := context.Background()
	s := NewMockStore(t)
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})

	u := &usermodel.User{ID: "test"}
	cr := &model.CreateRequest{
		Name: "test",
		Size: 123,
		Parts: []*model.Part{{
			Num:      0,
			Size:     123,
			Status:   0,
			Checksum: "checksum",
		}},
	}
	s.EXPECT().Create(ctx, mock.Anything).Run(func(_ context.Context, v *model.Video) {
		assert.Equal(t, cr.Name, v.Name)
		assert.Equal(t, cr.Size, v.Size)
		require.NotNil(t, v.UploadInfo)
		assert.Equal(t, len(cr.Parts), len(v.UploadInfo.Parts))
	}).Return(errors.New("err"))
	v, err := svc.CreateVideo(ctx, u, cr)
	require.Nil(t, v)
	require.ErrorIs(t, err, model.ErrStorage)
}

func TestService_CreateVideoAlreadyExists(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	ctx := context.Background()
	s := NewMockStore(t)
	svc := NewService(&ServiceConfig{
		Logger: logger,
		Store:  s,
	})

	u := &usermodel.User{ID: "test"}
	cr := &model.CreateRequest{
		Name: "test",
		Size: 123,
		Parts: []*model.Part{{
			Num:      0,
			Size:     123,
			Status:   0,
			Checksum: "checksum",
		}},
	}
	s.EXPECT().Create(ctx, mock.Anything).Run(func(_ context.Context, v *model.Video) {
		assert.Equal(t, cr.Name, v.Name)
		assert.Equal(t, cr.Size, v.Size)
		require.NotNil(t, v.UploadInfo)
		assert.Equal(t, len(cr.Parts), len(v.UploadInfo.Parts))
	}).Return(model.ErrAlreadyExists).Times(videoCreateRetries)
	v, err := svc.CreateVideo(ctx, u, cr)
	require.Nil(t, v)
	require.ErrorIs(t, err, model.ErrStorage)
	require.ErrorIs(t, err, model.ErrGivenUp)
}

func TestService_CreateVideoAlreadyExistsSuccess(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	ctx := context.Background()
	s := NewMockStore(t)
	ss := NewMockSessionStore(t)
	svc := NewService(&ServiceConfig{
		Logger:             logger,
		Store:              s,
		UploadSessionStore: ss,
		UploadURLPrefix:    "http://test",
	})

	u := &usermodel.User{ID: "test"}
	cr := &model.CreateRequest{
		Name: "test",
		Size: 123,
		Parts: []*model.Part{{
			Num:      0,
			Size:     123,
			Status:   0,
			Checksum: "checksum",
		}},
	}
	var (
		sessID string
		vid    string
		loc    string
	)
	s.EXPECT().Create(ctx, mock.Anything).Run(func(_ context.Context, v *model.Video) {
		assert.Equal(t, cr.Name, v.Name)
		assert.Equal(t, cr.Size, v.Size)
		require.NotNil(t, v.UploadInfo)
		assert.Equal(t, len(cr.Parts), len(v.UploadInfo.Parts))
	}).Return(model.ErrAlreadyExists).Times(videoCreateRetries - 1)
	s.EXPECT().Create(ctx, mock.Anything).Run(func(_ context.Context, v *model.Video) {
		loc = v.Location
		vid = v.ID
		assert.Equal(t, cr.Name, v.Name)
		assert.Equal(t, cr.Size, v.Size)
		require.NotNil(t, v.UploadInfo)
		assert.Equal(t, len(cr.Parts), len(v.UploadInfo.Parts))
	}).Return(nil)
	ss.EXPECT().Set(ctx, mock.Anything).Run(func(_ context.Context, sess *session.Session) {
		sessID = sess.ID
		assert.Equal(t, sessID, loc)
		assert.Equal(t, sess.VideoID, vid)
		assert.Equal(t, uint64(defaultPartSize), sess.PartSize)
	}).Return(nil)
	v, err := svc.CreateVideo(ctx, u, cr)
	require.NoError(t, err)
	require.NotNil(t, v)
	assert.Equal(t, cr.Name, v.Name)
	assert.Equal(t, cr.Size, v.Size)
	require.NotNil(t, v.UploadInfo)
	assert.Equal(t, "http://test/"+sessID, v.UploadInfo.URL)
}

func TestService_CreateVideoSessStorageErr(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	ctx := context.Background()
	s := NewMockStore(t)
	ss := NewMockSessionStore(t)
	svc := NewService(&ServiceConfig{
		Logger:             logger,
		Store:              s,
		UploadSessionStore: ss,
		UploadURLPrefix:    "http://test",
	})

	u := &usermodel.User{ID: "test"}
	cr := &model.CreateRequest{
		Name: "test",
		Size: 123,
		Parts: []*model.Part{{
			Num:      0,
			Size:     123,
			Status:   0,
			Checksum: "checksum",
		}},
	}
	var (
		sessID string
		vid    string
		loc    string
	)
	s.EXPECT().Create(ctx, mock.Anything).Run(func(_ context.Context, v *model.Video) {
		loc = v.Location
		vid = v.ID
		assert.Equal(t, cr.Name, v.Name)
		assert.Equal(t, cr.Size, v.Size)
		require.NotNil(t, v.UploadInfo)
		assert.Equal(t, len(cr.Parts), len(v.UploadInfo.Parts))
	}).Return(nil)
	ss.EXPECT().Set(ctx, mock.Anything).Run(func(_ context.Context, sess *session.Session) {
		sessID = sess.ID
		assert.Equal(t, sessID, loc)
		assert.Equal(t, sess.VideoID, vid)
		assert.Equal(t, uint64(defaultPartSize), sess.PartSize)
	}).Return(errors.New("test"))
	v, err := svc.CreateVideo(ctx, u, cr)
	require.ErrorIs(t, err, model.ErrSessionStorage)
	require.Nil(t, v)
}
