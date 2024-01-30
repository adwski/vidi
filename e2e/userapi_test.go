//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/adwski/vidi/internal/api/user/model"
)

func TestUserRegistration(t *testing.T) {
	//-------------------------------------------------------------------------------
	// Login with not-existent user
	//-------------------------------------------------------------------------------
	userLoginFail(t, &model.UserRequest{
		Username: "testuser",
		Password: "testpass",
	})

	//-------------------------------------------------------------------------------
	// Register user
	//-------------------------------------------------------------------------------
	cookie := userRegister(t, &model.UserRequest{
		Username: "testuser",
		Password: "testpass",
	})
	t.Logf("user is registered, token: %v", cookie.Value)
}

func TestUserLogin(t *testing.T) {
	//-------------------------------------------------------------------------------
	// Login with existent user
	//-------------------------------------------------------------------------------
	cookie2 := userLogin(t, &model.UserRequest{
		Username: "testuser",
		Password: "testpass",
	})
	t.Logf("user is logged in, token: %v", cookie2.Value)

	//-------------------------------------------------------------------------------
	// Login with not-existent user
	//-------------------------------------------------------------------------------
	userLoginFail(t, &model.UserRequest{
		Username: "testuser2",
		Password: "testpass2",
	})
}
