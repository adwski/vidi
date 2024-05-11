//nolint:wrapcheck  // return grpc statuses as errors
package userside

import (
	"context"
	"errors"
	"fmt"

	"github.com/adwski/vidi/internal/api/user/auth"
	user "github.com/adwski/vidi/internal/api/user/model"
	"github.com/adwski/vidi/internal/api/video"
	g "github.com/adwski/vidi/internal/api/video/grpc"
	"github.com/adwski/vidi/internal/api/video/grpc/userside/pb"
	"github.com/adwski/vidi/internal/api/video/model"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	*pb.UnimplementedUsersideapiServer
	*g.Server
	logger   *zap.Logger
	videoSvc *video.Service
}

func NewServer(cfg *g.Config, videoSvc *video.Service) (*Server, error) {
	cfg.Logger = cfg.Logger.With(zap.String("component", "userside-srv"))
	var (
		err error
		srv = &Server{
			logger:   cfg.Logger,
			videoSvc: videoSvc,
		}
	)
	if srv.Server, err = g.NewServer(cfg, func(s grpc.ServiceRegistrar) {
		pb.RegisterUsersideapiServer(s, srv)
	}); err != nil {
		return nil, fmt.Errorf("cannot create grpc server: %w", err)
	}
	return srv, nil
}

func (srv *Server) WatchVideo(ctx context.Context, req *pb.WatchRequest) (*pb.WatchVideoResponse, error) {
	usr, err := getUser(ctx)
	if err != nil {
		return nil, err
	}
	url, err := srv.videoSvc.WatchVideo(ctx, usr, req.Id, true)
	if err == nil {
		return &pb.WatchVideoResponse{Url: string(url)}, nil
	}
	switch {
	case errors.Is(err, model.ErrNotFound):
		return nil, status.Error(codes.NotFound, "video is not found")
	case errors.Is(err, model.ErrNotReady), errors.Is(err, model.ErrState):
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	default:
		srv.logger.Error("WatchVideo failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "cannot get video")
	}
}

func (srv *Server) GetQuota(ctx context.Context, _ *pb.GetQuotaRequest) (*pb.QuotaResponse, error) {
	usr, err := getUser(ctx)
	if err != nil {
		return nil, err
	}
	stats, err := srv.videoSvc.GetQuotas(ctx, usr)
	if err != nil {
		srv.logger.Error("GetQuotas failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "cannot get user quota")
	}
	return &pb.QuotaResponse{
		SizeQuota:   stats.SizeQuota,
		SizeUsage:   stats.SizeUsage,
		VideosQuota: uint32(stats.VideosQuota),
		VideosUsage: uint32(stats.VideosUsage),
	}, nil
}

// CreateVideo handles video create & upload request.
func (srv *Server) CreateVideo(ctx context.Context, req *pb.CreateVideoRequest) (*pb.VideoResponse, error) {
	usr, err := getUser(ctx)
	if err != nil {
		return nil, err
	}
	var r = &model.CreateRequest{
		Name:  req.Name,
		Size:  req.Size,
		Parts: make([]*model.Part, 0, len(req.Parts)),
	}
	for _, p := range req.Parts {
		r.Parts = append(r.Parts, &model.Part{
			Num:      uint(p.Num),
			Size:     p.Size,
			Checksum: p.Checksum,
		})
	}
	vide, err := srv.videoSvc.CreateVideo(ctx, usr, r)
	if err != nil {
		srv.logger.Error("CreateVideo failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "cannot create video")
	}
	return videoResponse(vide), nil
}
func (srv *Server) GetVideo(ctx context.Context, req *pb.VideoRequest) (*pb.VideoResponse, error) {
	usr, err := getUser(ctx)
	if err != nil {
		return nil, err
	}
	vide, err := srv.videoSvc.GetVideo(ctx, usr, req.Id, req.ResumeUpload)
	switch {
	case errors.Is(err, model.ErrNotFound):
		return nil, status.Error(codes.NotFound, "video is not found")
	case errors.Is(err, model.ErrNotResumable):
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	case err != nil:
		srv.logger.Error("GetVideo failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "cannot get video")
	}
	return videoResponse(vide), nil
}
func (srv *Server) GetVideos(ctx context.Context, _ *pb.GetVideosRequest) (*pb.VideosResponse, error) {
	usr, err := getUser(ctx)
	if err != nil {
		return nil, err
	}
	videos, err := srv.videoSvc.GetVideos(ctx, usr)
	switch {
	case errors.Is(err, model.ErrNotFound):
		return nil, status.Error(codes.NotFound, "no videos")
	case err != nil:
		srv.logger.Error("GetVideos failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "cannot get videos")
	}
	var resp pb.VideosResponse
	for _, v := range videos {
		resp.Videos = append(resp.Videos, videoResponse(v))
	}
	return &resp, nil
}
func (srv *Server) DeleteVideo(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteVideoResponse, error) {
	usr, err := getUser(ctx)
	if err != nil {
		return nil, err
	}
	err = srv.videoSvc.DeleteVideo(ctx, usr, req.Id)
	switch {
	case errors.Is(err, model.ErrNotFound):
		return nil, status.Error(codes.NotFound, "video is not found")
	case err != nil:
		srv.logger.Error("DeleteVideo failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "cannot delete video")
	}
	return &pb.DeleteVideoResponse{}, nil
}

func videoResponse(v *model.Video) *pb.VideoResponse {
	r := &pb.VideoResponse{
		Id:        v.ID,
		Status:    int32(v.Status),
		CreatedAt: v.CreatedAt.UnixMilli(),
		Name:      v.Name,
		Size:      v.Size,
	}
	if v.UploadInfo == nil {
		return r
	}
	r.UploadUrl = v.UploadInfo.URL
	r.UploadParts = make([]*pb.VideoPart, 0, len(v.UploadInfo.Parts))
	for _, p := range v.UploadInfo.Parts {
		r.UploadParts = append(r.UploadParts, &pb.VideoPart{
			Num:      uint32(p.Num),
			Status:   int32(p.Status),
			Size:     p.Size,
			Checksum: p.Checksum,
		})
	}
	return r
}

func getUser(ctx context.Context) (*user.User, error) {
	claims, ok := auth.GetClaimsFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "cannot get claims")
	}
	return &user.User{
		ID:   claims.UserID,
		Name: claims.Name,
	}, nil
}
