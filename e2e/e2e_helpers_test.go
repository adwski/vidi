//go:build e2e

package e2e

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/Eyevinn/dash-mpd/mpd"
	common "github.com/adwski/vidi/internal/api/model"
	"github.com/adwski/vidi/internal/api/user/model"
	videohttp "github.com/adwski/vidi/internal/api/video/http"
	video "github.com/adwski/vidi/internal/api/video/model"
	"github.com/adwski/vidi/internal/file"
	"github.com/adwski/vidi/internal/mp4"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	endpointUserLogin    = "http://localhost:18081/api/user/login"
	endpointUserRegister = "http://localhost:18081/api/user/register"
	endpointVideo        = "http://localhost:18082/api/video/"
	endpointWatch        = "http://localhost:18082/api/watch/"

	partSize                 = 10 * 1024 * 1024
	contentTypeVidiMediaPart = "application/x-vidi-mediapart"

	testFilePath = "../testfiles/test_seq_h264_high.mp4"
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
	require.Equal(t, "ok", body.Message)
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

func videoWatch(t *testing.T, userCookie *http.Cookie, v *videohttp.VideoResponse) []byte {
	t.Helper()

	var (
		errBody common.Response
	)
	resp, err := resty.New().R().SetHeader("Accept", "application/json").
		SetError(&errBody).
		SetCookie(userCookie).SetDebug(true).Get(endpointWatch + v.ID)
	t.Log("video watch response", resp, err)
	require.NoError(t, err)
	require.True(t, resp.IsSuccess())
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.Empty(t, errBody.Error)
	require.NotEmpty(t, resp.Body())

	return resp.Body()
}

func videoWatchURL(t *testing.T, userCookie *http.Cookie, v *videohttp.VideoResponse) string {
	t.Helper()

	var (
		errBody   common.Response
		watchBody videohttp.WatchResponse
	)
	resp, err := resty.New().R().SetHeader("Accept", "application/json").
		SetError(&errBody).
		SetCookie(userCookie).
		SetResult(&watchBody).
		SetQueryParam("mode", "url").
		Get(endpointWatch + v.ID)
	require.NoError(t, err)
	require.True(t, resp.IsSuccess())
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.Empty(t, errBody.Error)
	require.NotEmpty(t, watchBody.WatchURL)

	return watchBody.WatchURL
}

func videoWatchFail(t *testing.T, userCookie *http.Cookie, v *videohttp.VideoResponse, code int) {
	t.Helper()

	var (
		errBody common.Response
	)
	resp, err := resty.New().R().SetHeader("Accept", "application/json").
		SetError(&errBody).
		SetCookie(userCookie).
		Get(endpointWatch + v.ID)
	require.NoError(t, err)
	require.True(t, resp.IsError())
	require.Equal(t, code, resp.StatusCode())
	require.NotEmpty(t, errBody.Error)
}

func videoUploadFail(t *testing.T, url string) {
	t.Helper()

	f, errF := os.Open(testFilePath)
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

	f, errF := os.Open(testFilePath)
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

func videoGet(t *testing.T, userCookie *http.Cookie, id string) *videohttp.VideoResponse {
	t.Helper()

	var (
		videoBody videohttp.VideoResponse
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
		videoBody videohttp.VideoResponse
		errBody   common.Response
	)
	resp, err := resty.New().R().SetHeader("Accept", "application/json").
		SetError(&errBody).
		SetCookie(userCookie).
		SetResult(&videoBody).Get(endpointVideo + id)
	require.NoError(t, err)
	require.True(t, resp.IsError())
	require.Equal(t, code, resp.StatusCode())
}

func videoGetAll(t *testing.T, userCookie *http.Cookie) []*videohttp.VideoResponse {
	t.Helper()

	var (
		videoBody = make([]*videohttp.VideoResponse, 0)
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

func videoCreate(t *testing.T, userCookie *http.Cookie) *videohttp.VideoResponse {
	t.Helper()

	// get size
	size, err := file.GetSize(testFilePath)
	require.NoError(t, err)
	require.Greater(t, size, uint64(0))

	// make parts with checksums
	parts, err := file.MakePartsFromFile(testFilePath, partSize, size)
	require.NoError(t, err)
	require.Greater(t, len(parts), 0)

	var (
		respBody videohttp.VideoResponse
		reqBody  video.CreateRequest
		errBody  common.Response
	)
	reqBody.Name = "test"
	reqBody.Size = size
	for _, part := range parts {
		reqBody.Parts = append(reqBody.Parts, &video.Part{
			Checksum: part.Checksum,
			Num:      part.Num,
			Size:     uint64(part.Size),
		})
	}
	resp, err := resty.New().R().SetHeader("Accept", "application/json").
		SetError(&errBody).
		SetCookie(userCookie).
		SetBody(&reqBody).
		SetResult(&respBody).
		Post(endpointVideo)
	t.Log("got create video response", resp)
	require.NoError(t, err)
	require.True(t, resp.IsSuccess())
	require.Empty(t, errBody.Error)
	assert.NotEmpty(t, respBody.CreatedAt)
	assert.NotEmpty(t, respBody.ID)
	assert.NotEmpty(t, respBody.UploadInfo.URL)
	status, err := video.GetStatusFromName(respBody.Status)
	require.NoError(t, err)
	require.Equal(t, video.StatusCreated, status)

	return &respBody
}

func videoUpload(t *testing.T, url string) {
	t.Helper()

	// get size
	size, err := file.GetSize(testFilePath)
	require.NoError(t, err)
	require.Greater(t, size, uint64(0))

	// make parts with checksums
	parts, err := file.MakePartsFromFile(testFilePath, partSize, size)
	require.NoError(t, err)
	require.Greater(t, len(parts), 0)

	// upload parts
	f, err := os.Open(testFilePath)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	var offset uint
	for _, part := range parts {
		require.Greater(t, part.Size, uint(0))
		b, errB := io.ReadAll(io.LimitReader(f, int64(part.Size)))
		require.NoError(t, errB)
		resp, rErr := resty.New().R().
			SetBody(b).
			SetHeader("Content-Length", strconv.FormatUint(uint64(part.Size), 10)).
			SetHeader("Content-Type", contentTypeVidiMediaPart).SetDebug(true).
			Post(url + "/" + strconv.FormatUint(uint64(part.Num), 10))
		t.Log("upload part response", resp, rErr)
		require.NoError(t, rErr)
		require.True(t, resp.IsSuccess())
		offset += part.Size
	}
}

func videoCreateFail(t *testing.T) {
	t.Helper()

	var (
		videoBody videohttp.VideoResponse
	)
	resp, err := resty.New().R().SetHeader("Accept", "application/json").
		SetResult(&videoBody).Post(endpointVideo)
	require.NoError(t, err)
	require.True(t, resp.IsError())
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode())
}

func watchVideoFromMPD(t *testing.T, mpdBody []byte) {
	t.Helper()

	vMpd, err := mpd.MPDFromBytes(mpdBody)
	require.NoError(t, err)
	require.NotEmpty(t, vMpd.BaseURL)
	url := string(vMpd.BaseURL[0].Value)

	checkStaticMPD(t, vMpd)
	downloadSegments(t, url)

	downloadSegmentsFail(t, url)
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

	downloadSegmentFail(t, r, prefixURL+"/not-existent.mp4", http.StatusBadRequest)
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

	downloadSegment(t, r, prefixURL+"soun1_init.mp4", "audio/mp4")
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
