//go:build e2e && skip

package e2e

import (
	"net/http"
	"strings"
	"testing"
	"time"

	user "github.com/adwski/vidi/internal/api/user/model"
	video "github.com/adwski/vidi/internal/api/video/model"
	"github.com/stretchr/testify/require"
)

func TestCreateFail(t *testing.T) {
	//-------------------------------------------------------------------------------
	// Create video with no cookie
	//-------------------------------------------------------------------------------
	videoCreateFail(t)
}

func TestCreateAndDeleteVideo(t *testing.T) {
	//-------------------------------------------------------------------------------
	// Login with existent user
	//-------------------------------------------------------------------------------
	cookie := userLogin(t, &user.UserRequest{
		Username: "testuser",
		Password: "testpass",
	})
	t.Logf("user logged in, token: %v", cookie.Value)

	//-------------------------------------------------------------------------------
	// Create video
	//-------------------------------------------------------------------------------
	videoResponse := videoCreate(t, cookie)
	t.Logf("video created, id: %s, upload url: %v", videoResponse.ID, videoResponse.UploadInfo.URL)

	//-------------------------------------------------------------------------------
	// Get video
	//-------------------------------------------------------------------------------
	videoResponse2 := videoGet(t, cookie, videoResponse.ID)
	t.Logf("video retrieved, id: %s", videoResponse2.ID)

	require.Equal(t, videoResponse.ID, videoResponse2.ID)
	require.Equal(t, videoResponse.Status, videoResponse2.Status)
	require.Equal(t, videoResponse.CreatedAt, videoResponse2.CreatedAt)

	//-------------------------------------------------------------------------------
	// Delete video
	//-------------------------------------------------------------------------------
	videoDelete(t, cookie, videoResponse.ID)
	t.Logf("video deleted, id: %s", videoResponse.ID)
}

func TestCreateAndUploadVideo(t *testing.T) {
	//-------------------------------------------------------------------------------
	// Login with existent user
	//-------------------------------------------------------------------------------
	cookie := userLogin(t, &user.UserRequest{
		Username: "testuser",
		Password: "testpass",
	})
	t.Logf("user logged in, token: %v", cookie.Value)

	//-------------------------------------------------------------------------------
	// Create video
	//-------------------------------------------------------------------------------
	videoResponse := videoCreate(t, cookie)
	t.Logf("video created, id: %s, upload url: %v", videoResponse.ID, videoResponse.UploadInfo.URL)

	//-------------------------------------------------------------------------------
	// Upload video
	//-------------------------------------------------------------------------------
	videoUpload(t, videoResponse.UploadInfo.URL)

	//-------------------------------------------------------------------------------
	// Wait until processed
	//-------------------------------------------------------------------------------
	deadline := time.After(10 * time.Second)
Loop:
	for {
		select {
		case <-time.After(3 * time.Second):
			videoResponse2 := videoGet(t, cookie, videoResponse.ID)
			t.Logf("video retrieved, id: %s, status: %v", videoResponse2.ID, videoResponse2.Status)
			status, err := video.GetStatusFromName(videoResponse2.Status)
			require.NoError(t, err)
			if status != video.StatusReady {
				continue Loop
			}
			break Loop

		case <-deadline:
			t.Errorf("video did not became ready")
			break Loop
		}
	}
	t.Log("video processed")
}

func TestWatchVideo(t *testing.T) {
	//-------------------------------------------------------------------------------
	// Login with existent user
	//-------------------------------------------------------------------------------
	cookie := userLogin(t, &user.UserRequest{
		Username: "testuser",
		Password: "testpass",
	})
	t.Logf("user logged in, token: %v", cookie.Value)

	//-------------------------------------------------------------------------------
	// Get videos
	//-------------------------------------------------------------------------------
	videosResponse := videoGetAll(t, cookie)
	t.Logf("videos retrieved: %d", len(videosResponse))

	//-------------------------------------------------------------------------------
	// Get watch URL
	//-------------------------------------------------------------------------------
	watchResponse := videoWatch(t, cookie, videosResponse[0])
	t.Logf("watch url retrieved: %s", watchResponse.WatchURL)

	watchVideo(t, watchResponse.WatchURL)
}

func TestFails(t *testing.T) {
	//-------------------------------------------------------------------------------
	// Login with existent user
	//-------------------------------------------------------------------------------
	cookie := userLogin(t, &user.UserRequest{
		Username: "testuser",
		Password: "testpass",
	})
	t.Logf("user logged in, token: %v", cookie.Value)

	//-------------------------------------------------------------------------------
	// Create video
	//-------------------------------------------------------------------------------
	videoResponse := videoCreate(t, cookie)
	t.Logf("video created, id: %s, upload url: %v", videoResponse.ID, videoResponse.UploadInfo.URL)

	//-------------------------------------------------------------------------------
	// Get video
	//-------------------------------------------------------------------------------
	videoGetFail(t, cookie, "not-existent", http.StatusNotFound)

	//-------------------------------------------------------------------------------
	// Get video no auth
	//-------------------------------------------------------------------------------
	videoGetFail(t, &http.Cookie{Name: "x", Value: "y"}, videoResponse.ID, http.StatusUnauthorized)

	//-------------------------------------------------------------------------------
	// Get watch URL no-ready video
	//-------------------------------------------------------------------------------
	videoWatchFail(t, cookie, videoResponse, http.StatusMethodNotAllowed)

	//-------------------------------------------------------------------------------
	// Upload video, invalid request
	//-------------------------------------------------------------------------------
	videoUploadFailGet(t, videoResponse.UploadInfo.URL)

	//-------------------------------------------------------------------------------
	// Upload video, invalid requests
	//-------------------------------------------------------------------------------
	url := videoResponse.UploadInfo.URL[:strings.LastIndex(videoResponse.UploadInfo.URL, "/")]
	videoUploadFail(t, url)
	videoUploadFail(t, url+"/qweqwe/")
}
