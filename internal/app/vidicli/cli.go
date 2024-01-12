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
	rootCmd.PersistentFlags().StringP("file", "f", "", "input file")
	rootCmd.PersistentFlags().StringP("outdir", "o", "./output", "output dir")
	rootCmd.PersistentFlags().DurationP("segduration", "s", defaultSegmentDuration, "segment duration")

	rootCmd.AddCommand(segmentCmd)
	rootCmd.AddCommand(dumpCmd)
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
		segduration := cast.ToDuration(cmd.Flag("segduration").Value.String())
		mp4.Dump(fileName, segduration)
	},
}

func segmentFile(fileName, outdir string, segDuration time.Duration) {
	var (
		logger     = logging.GetZapLoggerDefaultLevel()
		mediaStore = file.NewStore("", outdir)

		proc = processor.New(&processor.Config{
			Logger:          logger,
			Store:           mediaStore,
			SegmentDuration: segDuration,
		})
	)
	f, _ := os.Open(fileName)
	defer func() { _ = f.Close() }()
	err := proc.ProcessFileFromReader(context.Background(), f, "")
	if err != nil {
		logger.Error("error processing file", zap.Error(err))
	}
	logger.Info("processing is done", zap.String("output", outdir))
}
