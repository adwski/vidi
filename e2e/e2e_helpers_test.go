//go:build e2e
// +build e2e

package e2e

import (
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/Eyevinn/dash-mpd/mpd"
	common "github.com/adwski/vidi/internal/api/model"
	"github.com/adwski/vidi/internal/api/user/model"
	video "github.com/adwski/vidi/internal/api/video/model"
	"github.com/adwski/vidi/internal/mp4"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	endpointUserLogin    = "http://localhost:18081/api/user/login"
	endpointUserRegister = "http://localhost:18081/api/user/register"
	endpointVideo        = "http://localhost:18082/api/video/user/"
)

func userRegister(t *testing.T, user *model.UserRequest) *http.Cookie {
	t.Helper()

	resp, body := makeCommonRequest(t, endpointUserRegister, user)
	require.True(t, resp.IsSuccess())
	require.Empty(t, body.Error)
	require.Equal(t, "registration complete", body.Message)
	return getCookieWithToken(t, resp.Cookies())
}

func userRegisterFail(t *testing.T, user any, code int) {
	t.Helper()

	resp, body := makeCommonRequest(t, endpointUserRegister, user)
	require.True(t, resp.IsError())
	require.NotEmpty(t, body.Error)
	require.Empty(t, body.Message)
	require.Equal(t, code, resp.StatusCode())
}

func userLogin(t *testing.T, user *model.UserRequest) *http.Cookie {
	t.Helper()

	resp, body := makeCommonRequest(t, endpointUserLogin, user)
	require.True(t, resp.IsSuccess())
	require.Empty(t, body.Error)
	require.Equal(t, "login ok", body.Message)
	return getCookieWithToken(t, resp.Cookies())
}

func userLoginFail(t *testing.T, user any, code int) {
	t.Helper()

	resp, body := makeCommonRequest(t, endpointUserLogin, user)
	require.Truef(t, resp.IsError(), "user should not exist")
	require.Empty(t, body.Message)
	require.NotEmpty(t, body.Error)
	require.Equal(t, code, resp.StatusCode())
}

func makeCommonRequest(t *testing.T, url string, reqBody interface{}) (*resty.Response, *common.Response) {
	t.Helper()

	var (
		body common.Response
	)
	resp, err := resty.New().R().SetHeader("Accept", "application/json").
		SetError(&body).
		SetResult(&body).
		SetBody(reqBody).Post(url)
	require.NoError(t, err)
	return resp, &body
}

func getCookieWithToken(t *testing.T, cookies []*http.Cookie) *http.Cookie {
	t.Helper()

	var (
		userCookie *http.Cookie
	)
	for _, cookie := range cookies {
		if cookie.Name == "vidiSessID" {
			userCookie = cookie
			break
		}
	}
	require.NotNilf(t, userCookie, "cookie should exist")
	require.NotEmpty(t, userCookie.Value, "cookie should not be empty")
	return userCookie
}

func videoWatch(t *testing.T, userCookie *http.Cookie, v *video.Response) *video.WatchResponse {
	t.Helper()

	var (
		errBody   common.Response
		watchBody video.WatchResponse
	)
	resp, err := resty.New().R().SetHeader("Accept", "application/json").
		SetError(&errBody).
		SetCookie(userCookie).
		SetResult(&watchBody).Post(endpointVideo + v.ID + "/watch")
	require.NoError(t, err)
	require.True(t, resp.IsSuccess())
	require.Equal(t, http.StatusAccepted, resp.StatusCode())
	require.Empty(t, errBody.Error)
	require.NotEmpty(t, watchBody.WatchURL)

	return &watchBody
}

func videoWatchFail(t *testing.T, userCookie *http.Cookie, v *video.Response, code int) {
	t.Helper()

	var (
		errBody common.Response
	)
	resp, err := resty.New().R().SetHeader("Accept", "application/json").
		SetError(&errBody).
		SetCookie(userCookie).
		Post(endpointVideo + v.ID + "/watch")
	require.NoError(t, err)
	require.True(t, resp.IsError())
	require.Equal(t, code, resp.StatusCode())
	require.NotEmpty(t, errBody.Error)
}

func videoUpload(t *testing.T, url string) {
	t.Helper()

	f, errF := os.Open("../testFiles/test_seq_h264_high.mp4")
	require.NoError(t, errF)

	resp, err := resty.New().R().
		SetHeader("Content-Type", "video/mp4").
		SetBody(f).Post(url)
	require.NoError(t, err)
	require.True(t, resp.IsSuccess())
	require.Equal(t, http.StatusNoContent, resp.StatusCode())
}

func videoUploadFail(t *testing.T, url string) {
	t.Helper()

	f, errF := os.Open("../testFiles/test_seq_h264_high.mp4")
	require.NoError(t, errF)

	resp, err := resty.New().R().
		SetHeader("Content-Type", "video/mp4").
		SetBody(f).Post(url)
	require.NoError(t, err)
	require.True(t, resp.IsError())
	require.Equal(t, http.StatusBadRequest, resp.StatusCode())
}

func videoUploadFailGet(t *testing.T, url string) {
	t.Helper()

	f, errF := os.Open("../testFiles/test_seq_h264_high.mp4")
	require.NoError(t, errF)

	resp, err := resty.New().R().
		SetHeader("Content-Type", "video/mp4").
		SetBody(f).Get(url)
	require.NoError(t, err)
	require.True(t, resp.IsError())
	require.Equal(t, http.StatusBadRequest, resp.StatusCode())
}

func videoDelete(t *testing.T, userCookie *http.Cookie, id string) {
	t.Helper()

	var (
		body common.Response
	)
	resp, err := resty.New().R().SetHeader("Accept", "application/json").
		SetError(&body).
		SetCookie(userCookie).
		SetResult(&body).Delete(endpointVideo + id)
	require.NoError(t, err)
	require.True(t, resp.IsSuccess())
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.Empty(t, body.Error)
	require.Equal(t, "ok", body.Message)
}

func videoGet(t *testing.T, userCookie *http.Cookie, id string) *video.Response {
	t.Helper()

	var (
		videoBody video.Response
		errBody   common.Response
	)
	resp, err := resty.New().R().SetHeader("Accept", "application/json").
		SetError(&errBody).
		SetCookie(userCookie).
		SetResult(&videoBody).Get(endpointVideo + id)
	require.NoError(t, err)
	require.True(t, resp.IsSuccess())
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.Empty(t, errBody.Error)
	require.NotEmpty(t, videoBody.CreatedAt)
	require.NotEmpty(t, videoBody.ID)

	return &videoBody
}

func videoGetFail(t *testing.T, userCookie *http.Cookie, id string, code int) {
	t.Helper()

	var (
		videoBody video.Response
		errBody   common.Response
	)
	resp, err := resty.New().R().SetHeader("Accept", "application/json").
		SetError(&errBody).
		SetCookie(userCookie).
		SetResult(&videoBody).Get(endpointVideo + id)
	require.NoError(t, err)
	require.True(t, resp.IsError())
	require.Equal(t, code, resp.StatusCode())
	// require.NotEmpty(t, errBody.Error)
}

func videoGetAll(t *testing.T, userCookie *http.Cookie) []*video.Response {
	t.Helper()

	var (
		videoBody = make([]*video.Response, 0)
		errBody   common.Response
	)
	resp, err := resty.New().R().SetHeader("Accept", "application/json").
		SetError(&errBody).
		SetCookie(userCookie).
		SetResult(&videoBody).Get(endpointVideo)
	require.NoError(t, err)
	require.True(t, resp.IsSuccess())
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.Empty(t, errBody.Error)
	require.NotEmpty(t, videoBody)

	return videoBody
}

func videoCreate(t *testing.T, userCookie *http.Cookie) *video.Response {
	t.Helper()

	var (
		videoBody video.Response
		errBody   common.Response
	)
	resp, err := resty.New().R().SetHeader("Accept", "application/json").
		SetError(&errBody).
		SetCookie(userCookie).
		SetResult(&videoBody).Post(endpointVideo)
	require.NoError(t, err)
	require.True(t, resp.IsSuccess())
	require.Empty(t, errBody.Error)
	require.NotEmpty(t, videoBody.CreatedAt)
	require.NotEmpty(t, videoBody.ID)
	require.NotEmpty(t, videoBody.UploadURL)

	status, err := videoBody.GetStatus()
	require.NoError(t, err)
	require.Equal(t, video.StatusCreated, status)

	return &videoBody
}

func videoCreateFail(t *testing.T) {
	t.Helper()

	var (
		videoBody video.Response
	)
	resp, err := resty.New().R().SetHeader("Accept", "application/json").
		SetResult(&videoBody).Post(endpointVideo)
	require.NoError(t, err)
	require.True(t, resp.IsError())
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode())
}

func watchVideo(t *testing.T, url string) {
	t.Helper()

	resp, err := resty.New().R().Get(url)
	require.NoError(t, err)
	require.True(t, resp.IsSuccess())
	require.NotEmpty(t, resp.Body())

	vMpd, err := mpd.MPDFromBytes(resp.Body())
	require.NoError(t, err)

	checkStaticMPD(t, vMpd)
	downloadSegments(t, url)

	downloadSegmentsFail(t, url)
}

func downloadSegmentsFail(t *testing.T, url string) {
	t.Helper()

	prefixURL := strings.TrimSuffix(url, "/manifest.mpd")

	prefixURLNoSess := prefixURL[:strings.LastIndexByte(prefixURL, '/')]

	r := resty.New()

	downloadSegmentFail(t, r, prefixURL+"/not-existent.mp4", http.StatusNotFound)
	downloadSegmentFail(t, r, prefixURLNoSess+"/qweqweqwe/vide1_1.m4s", http.StatusNotFound)
	downloadSegmentFail(t, r, prefixURLNoSess, http.StatusBadRequest)
	downloadSegmentFail(t, r, prefixURLNoSess+"/qw/", http.StatusBadRequest)
	downloadSegmentFail(t, r, prefixURL+"/vide1_init.wrong", http.StatusBadRequest)
	downloadSegmentFailPost(t, r, prefixURL+"/vide1_init.mp4", http.StatusBadRequest)
	spew.Dump(prefixURLNoSess)
}

func downloadSegments(t *testing.T, url string) {
	t.Helper()

	prefixURL := strings.TrimSuffix(url, "manifest.mpd")

	r := resty.New()

	downloadSegment(t, r, prefixURL+"vide1_init.mp4", "video/mp4")
	downloadSegment(t, r, prefixURL+"vide1_1.m4s", "video/iso.segment")
	downloadSegment(t, r, prefixURL+"vide1_2.m4s", "video/iso.segment")
	downloadSegment(t, r, prefixURL+"vide1_3.m4s", "video/iso.segment")
	downloadSegment(t, r, prefixURL+"vide1_4.m4s", "video/iso.segment")

	downloadSegment(t, r, prefixURL+"soun1_init.mp4", "video/mp4")
	downloadSegment(t, r, prefixURL+"soun1_1.m4s", "video/iso.segment")
	downloadSegment(t, r, prefixURL+"soun1_2.m4s", "video/iso.segment")
	downloadSegment(t, r, prefixURL+"soun1_3.m4s", "video/iso.segment")
	downloadSegment(t, r, prefixURL+"soun1_4.m4s", "video/iso.segment")

	t.Log("all segments downloaded")
}

func downloadSegment(t *testing.T, r *resty.Client, url, contentType string) {
	t.Helper()

	resp, err := r.R().Get(url)
	require.NoError(t, err)
	require.True(t, resp.IsSuccess())
	require.NotEmpty(t, resp.Body())

	assert.Equal(t, contentType, resp.Header().Get("Content-Type"))
}

func downloadSegmentFail(t *testing.T, r *resty.Client, url string, code int) {
	t.Helper()

	resp, err := r.R().Get(url)
	require.NoError(t, err)
	require.True(t, resp.IsError())
	require.Equal(t, code, resp.StatusCode())
}

func downloadSegmentFailPost(t *testing.T, r *resty.Client, url string, code int) {
	t.Helper()

	resp, err := r.R().Post(url)
	require.NoError(t, err)
	require.True(t, resp.IsError())
	require.Equal(t, code, resp.StatusCode())
}

func checkStaticMPD(t *testing.T, vMpd *mpd.MPD) {
	t.Helper()

	assert.Equal(t, "urn:mpeg:dash:schema:mpd:2011", vMpd.XMLNs)
	assert.Equal(t, mpd.ListOfProfilesType("urn:mpeg:dash:profile:isoff-on-demand:2011"), vMpd.Profiles)
	assert.Equal(t, "static", *vMpd.Type)
	assert.Equal(t, 10.0, vMpd.MediaPresentationDuration.Seconds())

	require.NotEmpty(t, vMpd.Periods)
	require.NotNil(t, vMpd.Periods[0])
	require.Len(t, vMpd.Periods[0].AdaptationSets, 2)

	var (
		vide, soun *mpd.RepresentationType
	)
	for _, as := range vMpd.Periods[0].AdaptationSets {
		require.NotNil(t, as)
		require.NotNil(t, as.SegmentTemplate)

		st := as.SegmentTemplate
		assert.Equal(t, "$RepresentationID$_$Number$"+mp4.SegmentSuffix, st.Media)
		assert.Equal(t, "$RepresentationID$_"+mp4.SegmentSuffixInit, st.Initialization)
		assert.Equal(t, uint32(1), *st.MultipleSegmentBaseType.StartNumber)
		assert.NotZero(t, *st.MultipleSegmentBaseType.Duration)
		assert.NotZero(t, *st.MultipleSegmentBaseType.SegmentBaseType.Timescale)
		assert.Equalf(t, 3, int(*st.MultipleSegmentBaseType.Duration /
			*st.MultipleSegmentBaseType.SegmentBaseType.Timescale),
			"segment duration should be equal to processor.segment_duration")

		require.Len(t, as.Representations, 1)

		switch as.Representations[0].Id {
		case "vide1":
			vide = as.Representations[0]
			assert.Equal(t, "avc1.64001f", vide.RepresentationBaseType.Codecs)
			assert.Equal(t, "video/mp4", as.RepresentationBaseType.MimeType)
		case "soun1":
			soun = as.Representations[0]
			assert.Equal(t, "mp4a.40.2", soun.RepresentationBaseType.Codecs)
			assert.Equal(t, "audio/mp4", as.RepresentationBaseType.MimeType)
		}
	}
	require.NotNilf(t, vide, "video representation must be in MPD")
	require.NotNilf(t, soun, "audio representation must be in MPD")

	t.Log("mpd is ok")
}
