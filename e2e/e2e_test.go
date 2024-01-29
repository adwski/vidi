//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	common "github.com/adwski/vidi/internal/api/model"
	"github.com/adwski/vidi/internal/api/user/model"
	"github.com/adwski/vidi/internal/app/user"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	var (
		done        = make(chan struct{})
		ctx, cancel = context.WithCancel(context.Background())
	)

	go func() {
		user.NewApp().RunWithContextAndConfig(ctx, "userapi.yaml")
		done <- struct{}{}
	}()

	time.Sleep(time.Second)

	code := m.Run()
	cancel()
	<-done
	defer func() {
		os.Exit(code)
	}()
}

func TestE2EUserRegistration(t *testing.T) {
	//-------------------------------------------------------------------------------
	// Login with not-existent user
	//-------------------------------------------------------------------------------
	userLoginFail(t, &model.UserRequest{
		Username: "qweqweqwe",
		Password: "asdasdasd",
	})

	//-------------------------------------------------------------------------------
	// Register user
	//-------------------------------------------------------------------------------
	cookie := userRegister(t, &model.UserRequest{
		Username: "qweqweqwe",
		Password: "asdasdasd",
	})
	t.Logf("user is registered, token: %v", cookie.Value)

	//-------------------------------------------------------------------------------
	// Login with existent user
	//-------------------------------------------------------------------------------
	cookie2 := userLogin(t, &model.UserRequest{
		Username: "qweqweqwe",
		Password: "asdasdasd",
	})
	t.Logf("user is logged in, token: %v", cookie2.Value)

	//-------------------------------------------------------------------------------
	// Login with not-existent user
	//-------------------------------------------------------------------------------
	userLoginFail(t, &model.UserRequest{
		Username: "qweqweqwe1",
		Password: "asdasdasd2",
	})
}

func userRegister(t *testing.T, user *model.UserRequest) *http.Cookie {
	t.Helper()

	resp, body := makeCommonRequest(t, "http://localhost:18081/api/user/register", user)
	require.True(t, resp.IsSuccess())
	require.Empty(t, body.Error)
	require.Equal(t, "registration complete", body.Message)
	return getCookieWithToken(t, resp.Cookies())
}

func userLogin(t *testing.T, user *model.UserRequest) *http.Cookie {
	t.Helper()

	resp, body := makeCommonRequest(t, "http://localhost:18081/api/user/login", user)
	require.True(t, resp.IsSuccess())
	require.Empty(t, body.Error)
	require.Equal(t, "login ok", body.Message)
	return getCookieWithToken(t, resp.Cookies())
}

func userLoginFail(t *testing.T, user *model.UserRequest) {
	t.Helper()

	resp, body := makeCommonRequest(t, "http://localhost:18081/api/user/login", user)
	require.Truef(t, resp.IsError(), "user should not exist")
	require.Empty(t, body.Message)
	require.NotEmpty(t, body.Error)
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
