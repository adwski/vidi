package serviceside

import (
	"context"
	"errors"
	"fmt"
	"github.com/adwski/vidi/internal/api/user/auth"
	"github.com/adwski/vidi/internal/api/video"
	g "github.com/adwski/vidi/internal/api/video/grpc"
	"github.com/adwski/vidi/internal/api/video/grpc/serviceside/pb"
	"github.com/adwski/vidi/internal/api/video/model"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	*pb.UnimplementedServicesideapiServer
	*g.Server
	logger   *zap.Logger
	videoSvc *video.Service
}

func NewServer(cfg *g.Config, videoSvc *video.Service) (*Server, error) {
	cfg.Logger = cfg.Logger.With(zap.String("component", "serviceside-srv"))
	var (
		err error
		srv = &Server{
			logger:   cfg.Logger,
			videoSvc: videoSvc,
		}
	)
	if srv.Server, err = g.NewServer(cfg, func(s grpc.ServiceRegistrar) {
		pb.RegisterServicesideapiServer(s, srv)
	}); err != nil {
		return nil, fmt.Errorf("cannot create grpc server: %w", err)
	}
	return srv, nil
}

func (srv *Server) GetVideosByStatus(ctx context.Context, req *pb.GetByStatusRequest) (*pb.VideoListResponse, error) {
	if err := checkServiceClaims(ctx); err != nil {
		return nil, err
	}
	videos, err := srv.videoSvc.GetVideosByStatus(ctx, model.Status(req.Status))
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	var resp pb.VideoListResponse
	resp.Videos = make([]*pb.Video, 0, len(videos))
	for _, v := range videos {
		resp.Videos = append(resp.Videos, &pb.Video{
			Id:        v.ID,
			Status:    int32(v.Status),
			CreatedAt: uint64(v.CreatedAt.Unix()),
			Location:  v.Location,
		})
	}
	return &resp, nil
}

func (srv *Server) UpdateVideo(ctx context.Context, req *pb.UpdateVideoRequest) (*pb.UpdateVideoResponse, error) {
	if err := checkServiceClaims(ctx); err != nil {
		return nil, err
	}
	err := srv.videoSvc.UpdateVideoStatusAndLocation(ctx, req.Id, req.Location, model.Status(req.Status))
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &pb.UpdateVideoResponse{}, nil
}

func (srv *Server) UpdateVideoStatus(ctx context.Context, req *pb.UpdateVideoStatusRequest) (*pb.UpdateVideoStatusResponse, error) {
	if err := checkServiceClaims(ctx); err != nil {
		return nil, err
	}
	err := srv.videoSvc.UpdateVideoStatus(ctx, req.Id, model.Status(req.Status))
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &pb.UpdateVideoStatusResponse{}, nil
}

func (srv *Server) NotifyPartUpload(ctx context.Context, req *pb.NotifyPartUploadRequest) (*pb.NotifyPartUploadResponse, error) {
	if err := checkServiceClaims(ctx); err != nil {
		return nil, err
	}
	part := &model.Part{
		Num:      uint(req.Num),
		Checksum: req.Checksum,
	}
	err := srv.videoSvc.NotifyPartUpload(ctx, req.VideoId, part)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.NotifyPartUploadResponse{}, nil
}

func checkServiceClaims(ctx context.Context) error {
	claims, ok := auth.GetClaimsFromContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "cannot get claims")
	}
	if !claims.IsService() {
		return status.Error(codes.Unauthenticated, "not a service account")
	}
	return nil
}
