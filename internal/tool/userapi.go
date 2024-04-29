package tool

import (
	"fmt"
	"go.uber.org/zap"
)

func (t *Tool) processNewUser(uc userControl) (err error) {
	var token string
	switch uc.option {
	case newUserOptionLogin:
		token, err = t.userapi.Login(uc.username, uc.password)
	case newUserOptionRegister:
		token, err = t.userapi.Register(uc.username, uc.password)
	default:
		return fmt.Errorf("invalid option: %d", uc.option)
	}
	if err != nil {
		return err
	}
	t.state.Users = append(t.state.Users, User{
		Name:  uc.username,
		Token: token,
	})
	t.state.CurrentUser = len(t.state.Users) - 1
	if err = t.state.persist(); err != nil {
		return err
	}
	t.logger.Debug("user added",
		zap.String("username", uc.username),
		zap.Int("id", t.state.CurrentUser))
	return
}

func (t *Tool) processLoginExistingUser(username, password string) error {
	token, err := t.userapi.Login(username, password)
	if err != nil {
		return err
	}
	t.state.Users[t.state.CurrentUser].Token = token
	if err = t.state.persist(); err != nil {
		return err
	}
	t.logger.Debug("user added",
		zap.String("username", username),
		zap.Int("id", t.state.CurrentUser))
	return nil
}
