package cli

import (
	"context"
	"io"
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
		segmentFile(cmd.OutOrStdout(), fileName, outdir, segduration)
	},
}

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "dump mp4 file",
	Run: func(cmd *cobra.Command, args []string) {
		fileName := cmd.Flag("file").Value.String()
		segDuration := cast.ToDuration(cmd.Flag("segduration").Value.String())
		mp4.Dump(cmd.OutOrStdout(), fileName, segDuration)
	},
}

func segmentFile(w io.Writer, fileName, outdir string, segDuration time.Duration) {
	var (
		logger     = logging.GetZapLoggerWriter(w)
		mediaStore = file.NewStore("", outdir)
		proc, _    = processor.New(&processor.Config{ // will never produce error in local mode
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
	_, err = proc.ProcessFileFromReader(context.Background(), f, "")
	if err != nil {
		logger.Error("error processing file", zap.Error(err))
		return
	}
	logger.Info("processing is done", zap.String("output", outdir))
}
