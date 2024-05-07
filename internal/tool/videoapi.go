package tool

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/adwski/vidi/internal/api/video/grpc/userside/pb"
	"github.com/dustin/go-humanize"
	"github.com/minio/sha256-simd"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"io"
	"os"
	"strconv"
	"time"
)

const (
	partSize = 10 * 1024 * 1024
)

func (t *Tool) getVideos() ([]Video, error) {
	ctx := t.getUserMDCtx()
	resp, err := t.videoapi.GetVideos(ctx, &pb.GetVideosRequest{})
	if err != nil {
		t.logger.Error("unable to get videos", zap.Error(err))
		return nil, fmt.Errorf("unable to get videos: %w", err)
	}
	var videos = make([]Video, 0, len(resp.Videos))
	for _, v := range resp.Videos {
		videos = append(videos, Video{
			ID:        v.Id,
			Name:      "",
			Status:    v.Status,
			Size:      "",
			CreatedAt: v.CreatedAt,
		})
	}
	return videos, nil
}

func (t *Tool) getQuotas() ([]QuotaParam, error) {
	ctx := t.getUserMDCtx()
	resp, err := t.videoapi.GetQuota(ctx, &pb.GetQuotaRequest{})
	if err != nil {
		t.logger.Error("unable to get quotas", zap.Error(err))
		return nil, fmt.Errorf("unable to get quotas: %w", err)
	}
	return []QuotaParam{
		{"size_quota", humanize.Bytes(resp.SizeQuota)},
		{"size_usage", humanize.Bytes(resp.SizeUsage)},
		{"size_remain", humanize.Bytes(resp.SizeQuota - resp.SizeUsage)},
		{"videos_quota", strconv.FormatUint(uint64(resp.VideosQuota), 10)},
		{"videos_usage", strconv.FormatUint(uint64(resp.VideosUsage), 10)},
		{"videos_remain", strconv.FormatUint(uint64(resp.VideosQuota-resp.VideosUsage), 10)},
	}, nil
}

func (t *Tool) prepareUpload(filePath string) error {
	size, err := getFileSize(filePath)
	if err != nil {
		return err
	}
	if err = t.checkQuotas(size); err != nil {
		return err
	}
	return t.prepareParts(filePath, size)
}

func getFileSize(filePath string) (uint64, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("unable to open file: %w", err)
	}
	defer func() { _ = f.Close() }()
	stat, err := f.Stat()
	if err != nil {
		return 0, fmt.Errorf("unable to get file stats: %w", err)
	}
	return uint64(stat.Size()), nil
}

func (t *Tool) checkQuotas(size uint64) error {
	ctx := t.getUserMDCtx()
	qResp, err := t.videoapi.GetQuota(ctx, &pb.GetQuotaRequest{})
	if err != nil {
		return fmt.Errorf("unable to get quotas: %w", err)
	}
	if qResp.VideosQuota == qResp.VideosUsage {
		return errors.New("video quota exceeded")
	}
	if qResp.SizeQuota-qResp.SizeUsage < size {
		return errors.New("size quota exceeded")
	}
	return nil
}

func (t *Tool) createVideo(name string, size uint64, uploadParts []Part) (*pb.VideoResponse, error) {
	ctx := t.getUserMDCtx()
	parts := make([]*pb.VideoPart, 0, len(uploadParts))
	for _, part := range uploadParts {
		parts = append(parts, &pb.VideoPart{
			Num:      uint32(part.Num),
			Size:     uint64(part.Size),
			Checksum: part.Checksum,
		})
	}
	cvResp, err := t.videoapi.CreateVideo(ctx, &pb.CreateVideoRequest{
		Size:  size,
		Name:  name,
		Parts: parts,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create video: %w", err)
	}
	return cvResp, nil
}

func (t *Tool) prepareParts(filePath string, size uint64) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("unable to open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	partCount := size / partSize
	if size%partSize != 0 {
		partCount++
	}
	uploadInfo := &Upload{
		Filename: filePath,
	}
	for i := 0; i < int(partCount); i++ {
		h := sha256.New()
		n, err := io.CopyN(h, f, partSize)
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("unable to calculate sha256 sum: %w", err)
			}
		}
		uploadInfo.Parts = append(uploadInfo.Parts, Part{
			Num:      uint(i),
			Size:     uint(n),
			Checksum: base64.StdEncoding.EncodeToString(h.Sum(nil)),
		})
	}
	t.logger.Debug("prepared upload info", zap.Any("info", uploadInfo))
	t.state.getCurrentUserUnsafe().CurrentUpload = uploadInfo
	return t.state.persist()
}

func (t *Tool) getTmpDir() string {
	return t.dir + "/tmp" + strconv.Itoa(int(time.Now().Unix()))
}

func (t *Tool) getUserMDCtx() context.Context {
	md := metadata.New(map[string]string{"authorization": "bearer " + t.state.getCurrentUserUnsafe().Token})
	return metadata.NewOutgoingContext(context.TODO(), md)
}
