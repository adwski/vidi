package vidictl

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

const (
	defaultSegmentDuration = 3 * time.Second

	defaultServiceJWTExpiration = 365 * 24 * time.Hour
)

var (
	rootCmd = &cobra.Command{
		Use:   "vidi-cli",
		Short: "vidi command line tool",
	}
)

// Execute executes the root command.
func Execute() int {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		return 1
	}
	return 0
}

func init() {
	mp4Cmd.AddCommand(dumpCmd)
	mp4Cmd.AddCommand(segmentCmd)
	mp4Cmd.PersistentFlags().StringP("file", "f", "input.mp4", "input file")
	mp4Cmd.PersistentFlags().StringP("outdir", "o", "./output", "output dir")
	mp4Cmd.PersistentFlags().DurationP("segduration", "s", defaultSegmentDuration, "segment duration")

	apiCmd.AddCommand(createSvcTokenCmd)
	apiCmd.PersistentFlags().StringP("svcname", "n", "service", "service name")
	apiCmd.PersistentFlags().StringP("jwtsecret", "s", "changeMe", "jwt secret")
	apiCmd.PersistentFlags().DurationP("expiration", "e", defaultServiceJWTExpiration, "token expiration")

	rootCmd.AddCommand(mp4Cmd)
	rootCmd.AddCommand(apiCmd)
}
