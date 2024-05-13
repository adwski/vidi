//go:build e2e

package e2e

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/adwski/vidi/internal/api/user/model"
)

var (
	testUserName = "testuser"
)

func init() {
	testUserName += strconv.Itoa(int(time.Now().Unix()))
}

func TestUserRegistration(t *testing.T) {
	//-------------------------------------------------------------------------------
	// Login with not-existent user
	//-------------------------------------------------------------------------------
	userLoginFail(t, &model.UserRequest{
		Username: testUserName,
		Password: "testpass",
	}, http.StatusUnauthorized)

	//-------------------------------------------------------------------------------
	// Register user
	//-------------------------------------------------------------------------------
	cookie := userRegister(t, &model.UserRequest{
		Username: testUserName,
		Password: "testpass",
	})
	t.Logf("user is registered, token: %v", cookie.Value)

	//-------------------------------------------------------------------------------
	// Register existing user
	//-------------------------------------------------------------------------------
	userRegisterFail(t, &model.UserRequest{
		Username: testUserName,
		Password: "testpass",
	}, http.StatusConflict)

	//-------------------------------------------------------------------------------
	// Register with invalid data
	//-------------------------------------------------------------------------------
	userRegisterFail(t, "", http.StatusBadRequest)
}

func TestUserLogin(t *testing.T) {
	//-------------------------------------------------------------------------------
	// Login with existent user
	//-------------------------------------------------------------------------------
	cookie2 := userLogin(t, &model.UserRequest{
		Username: testUserName,
		Password: "testpass",
	})
	t.Logf("user is logged in, token: %v", cookie2.Value)

	//-------------------------------------------------------------------------------
	// Login with not-existent user
	//-------------------------------------------------------------------------------
	userLoginFail(t, &model.UserRequest{
		Username: "testuser2",
		Password: "testpass2",
	}, http.StatusUnauthorized)

	//-------------------------------------------------------------------------------
	// Login with wrong password
	//-------------------------------------------------------------------------------
	userLoginFail(t, &model.UserRequest{
		Username: testUserName,
		Password: "testpass2",
	}, http.StatusUnauthorized)

	//-------------------------------------------------------------------------------
	// Login with invalid params
	//-------------------------------------------------------------------------------
	userLoginFail(t, "", http.StatusBadRequest)
}
