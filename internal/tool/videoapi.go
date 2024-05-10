package tool

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/adwski/vidi/internal/api/video/grpc/userside/pb"
	"github.com/adwski/vidi/internal/api/video/model"
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
			Name:      v.Name,
			Status:    model.Status(v.Status).String(),
			Size:      humanize.Bytes(v.Size),
			CreatedAt: time.UnixMilli(v.CreatedAt).Format(time.RFC3339),
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

func (t *Tool) deleteVideo(vid string) error {
	ctx := t.getUserMDCtx()
	_, err := t.videoapi.DeleteVideo(ctx, &pb.DeleteRequest{Id: vid})
	if err != nil {
		return fmt.Errorf("unable to delete video: %w", err)
	}
	return nil
}

func (t *Tool) resumeUploadFileNotify(upload *Upload) {
	partsToUpload, uploadURL, size, err := t.checkUploadPartsState(upload)
	if err != nil {
		t.fb <- err
		return
	}
	t.fb <- uploadInfo{
		name:     upload.Name,
		filePath: upload.Filename,
	}
	if len(partsToUpload) == 0 {
		// nothing to upload
		// This case should not happen
		t.fb <- uploadCompleted{wasCompletedBefore: true}
		return
	}

	f, err := os.Open(upload.Filename)
	if err != nil {
		t.fb <- fmt.Errorf("unable to open file: %w", err)
		return
	}
	defer func() { _ = f.Close() }()
	stat, err := f.Stat()
	if err != nil {
		t.fb <- fmt.Errorf("unable to get file stats: %w", err)
		return
	}
	if uint64(stat.Size()) != size {
		t.fb <- fmt.Errorf("file have changed during upload: %d != %d", stat.Size(), size)
		return
	}

	var offset = size - uint64(partSize*len(partsToUpload))
	t.fb <- uploadProgress{completed: offset, total: size}

	for n, part := range partsToUpload {
		if _, err = f.Seek(int64(n)*int64(partSize), io.SeekStart); err != nil {
			t.fb <- fmt.Errorf("unable to seek file part %d: %w", n, err)
		}
		resp, rErr := t.httpC.NewRequest().
			SetBody(io.LimitReader(f, int64(part.Size))).
			SetHeader("Content-Type", "application/x-vidi-mediapart").
			Post(uploadURL + "/" + strconv.FormatUint(uint64(part.Num), 10))
		if rErr != nil {
			t.fb <- fmt.Errorf("unable to upload part %d: %w", n, rErr)
			return
		}
		if resp.IsError() {
			t.fb <- fmt.Errorf("server responded with error status: %s", resp.Status())
			return
		}
		offset += part.Size
		t.fb <- uploadProgress{completed: offset, total: size}
	}
	t.fb <- uploadCompleted{}
}

func (t *Tool) checkUploadPartsState(upload *Upload) (map[uint32]*pb.VideoPart, string, uint64, error) {
	size := uint64(0)
	ctx := t.getUserMDCtx()
	resp, err := t.videoapi.GetVideo(ctx, &pb.VideoRequest{
		Id:           upload.ID,
		ResumeUpload: true,
	})
	if err != nil {
		return nil, "", 0, fmt.Errorf("unable to get video info: %w", err)
	}
	if len(resp.UploadParts) == 0 && model.Status(resp.Status) == model.StatusCreated {
		return nil, "", 0, errors.New("video does not have upload parts")
	}
	var (
		partsToUpload    = make(map[uint32]*pb.VideoPart)
		partsHaveLocally = make(map[uint]Part)
	)
	for _, p := range upload.Parts {
		partsHaveLocally[p.Num] = p
		size += uint64(p.Size)
	}
	for _, p := range resp.UploadParts {
		if p.Status == model.PartStatusOK {
			continue
		}
		pl, ok := partsHaveLocally[uint(p.Num)]
		if !ok {
			return nil, "", 0, fmt.Errorf("part is missing locally: %d", p.Num)
		}
		if pl.Checksum != p.Checksum {
			return nil, "", 0, fmt.Errorf("part checksum doesn't match, have: %s, remote: %s", pl.Checksum, p.Checksum)
		}
		partsToUpload[p.Num] = p
	}
	return partsToUpload, resp.UploadUrl, size, nil
}

func (t *Tool) uploadFileNotify(name, filePath string) {
	size, err := t.prepareUpload(filePath)
	if err != nil {
		t.fb <- err
		return
	}

	// get ref to current upload
	upload := t.state.activeUserUnsafe().CurrentUpload

	// create video in videoapi
	cv, err := t.createVideo(name, size, upload.Parts)
	if err != nil {
		t.fb <- err
		return
	}

	// save videoapi video id
	upload.ID = cv.Id
	if err = t.state.persist(); err != nil {
		t.fb <- err
		return
	}

	// upload parts
	f, err := os.Open(filePath)
	if err != nil {
		t.fb <- fmt.Errorf("unable to open file: %w", err)
		return
	}
	defer func() { _ = f.Close() }()
	stat, err := f.Stat()
	if err != nil {
		t.fb <- fmt.Errorf("unable to get file stats: %w", err)
		return
	}
	if uint64(stat.Size()) != size {
		t.fb <- fmt.Errorf("file have changed during upload: %d != %d", stat.Size(), size)
		return
	}

	var offset uint
	for _, part := range upload.Parts {
		resp, rErr := t.httpC.NewRequest().
			SetBody(io.LimitReader(f, int64(part.Size))).
			SetHeader("Content-Type", "application/x-vidi-mediapart").
			Post(cv.UploadUrl + "/" + strconv.FormatUint(uint64(part.Num), 10))
		if rErr != nil {
			t.fb <- fmt.Errorf("unable to upload part: %w", rErr)
			return
		}
		if resp.IsError() {
			t.fb <- fmt.Errorf("server responded with error status: %s", resp.Status())
			return
		}
		offset += part.Size
		t.fb <- uploadProgress{completed: uint64(offset), total: size}
	}
	t.fb <- uploadCompleted{}
}

func (t *Tool) prepareUpload(filePath string) (uint64, error) {
	size, err := getFileSize(filePath)
	if err != nil {
		return 0, err
	}
	if err = t.checkQuotas(size); err != nil {
		return 0, err
	}
	return size, t.prepareParts(filePath, size)
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
	t.logger.Debug("parts", zap.Any("parts", parts), zap.Any("cvResp", cvResp))
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
	for i := uint64(0); i < partCount; i++ {
		h := sha256.New()
		n, err := io.CopyN(h, f, partSize)
		if err != nil {
			if err != io.EOF || i != partCount-1 {
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
	t.state.activeUserUnsafe().CurrentUpload = uploadInfo
	return t.state.persist()
}

func (t *Tool) getTmpDir() string {
	return t.dir + "/tmp" + strconv.Itoa(int(time.Now().Unix()))
}

func (t *Tool) getUserMDCtx() context.Context {
	md := metadata.Pairs("authorization", "bearer "+t.state.activeUserUnsafe().Token)
	return metadata.NewOutgoingContext(context.TODO(), md)
}
