package cli

import (
	"io"
	"time"

	"github.com/adwski/vidi/internal/api/user/auth"
	"github.com/adwski/vidi/internal/logging"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "vidi api command group",
}

var createSvcTokenCmd = &cobra.Command{
	Use:   "create-service-token",
	Short: "create token suitable for videoapi service calls",
	Run: func(cmd *cobra.Command, args []string) {
		name := cmd.Flag("svcname").Value.String()
		secret := cmd.Flag("jwtsecret").Value.String()
		expiration := cast.ToDuration(cmd.Flag("expiration").Value.String())
		createServiceToken(cmd.OutOrStdout(), name, secret, expiration)
	},
}

func createServiceToken(w io.Writer, name, secret string, expiration time.Duration) {
	logger := logging.GetZapLoggerWriter(w)

	au, err := auth.NewAuth(&auth.Config{
		Secret:     secret,
		Expiration: expiration,
	})
	if err != nil {
		logger.Error("cannot init authenticator", zap.Error(err))
		return
	}

	token, errT := au.NewTokenForService(name)
	if errT != nil {
		logger.Error("cannot create token", zap.Error(errT))
		return
	}
	if _, err = w.Write([]byte(token)); err != nil {
		logger.Error("cannot write token", zap.Error(err))
	}
}
