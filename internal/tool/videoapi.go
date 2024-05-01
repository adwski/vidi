package tool

import (
	"context"
	"fmt"
	"github.com/adwski/vidi/internal/api/video/grpc/user/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

func (t *Tool) getVideos() ([]Video, error) {
	md := metadata.New(map[string]string{"bearer": t.state.getCurrentUserUnsafe().Token})
	ctx := metadata.NewOutgoingContext(context.TODO(), md)
	resp, err := t.videoapi.GetVideos(ctx, &pb.GetVideosRequest{})
	if err != nil {
		t.logger.Error("unable to get videos", zap.Error(err))
		return nil, fmt.Errorf("unable to get videos: %w", err)
	}
	var videos []Video
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
