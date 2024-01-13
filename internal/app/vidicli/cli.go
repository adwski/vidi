package vidicli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/adwski/vidi/internal/logging"
	"github.com/adwski/vidi/internal/media/processor"
	"github.com/adwski/vidi/internal/media/store/file"
	"github.com/adwski/vidi/internal/mp4"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const (
	defaultSegmentDuration = 3 * time.Second
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

	rootCmd.AddCommand(mp4Cmd)
}

var mp4Cmd = &cobra.Command{
	Use:   "mp4",
	Short: "mp4 isobmff command group",
}

var segmentCmd = &cobra.Command{
	Use:   "segment",
	Short: "segment mp4 file",
	Run: func(cmd *cobra.Command, args []string) {
		fileName := cmd.Flag("file").Value.String()
		outdir := cmd.Flag("outdir").Value.String()
		segduration := cast.ToDuration(cmd.Flag("segduration").Value.String())
		segmentFile(fileName, outdir, segduration)
	},
}

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "dump mp4 file",
	Run: func(cmd *cobra.Command, args []string) {
		fileName := cmd.Flag("file").Value.String()
		segDuration := cast.ToDuration(cmd.Flag("segduration").Value.String())
		mp4.Dump(fileName, segDuration)
	},
}

func segmentFile(fileName, outdir string, segDuration time.Duration) {
	var (
		logger     = logging.GetZapLoggerConsole()
		mediaStore = file.NewStore("", outdir)
		proc       = processor.New(&processor.Config{
			Logger:          logger,
			Store:           mediaStore,
			SegmentDuration: segDuration,
		})
	)
	f, err := os.Open(fileName)
	if err != nil {
		logger.Error("cannot open file", zap.Error(err))
		return
	}
	defer func() { _ = f.Close() }()
	err = proc.ProcessFileFromReader(context.Background(), f, "")
	if err != nil {
		logger.Error("error processing file", zap.Error(err))
		return
	}
	logger.Info("processing is done", zap.String("output", outdir))
}
