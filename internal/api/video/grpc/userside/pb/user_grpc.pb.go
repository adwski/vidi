// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.17.3
// source: internal/api/video/grpc/protobuf/user.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Usersideapi_GetQuota_FullMethodName    = "/videoapi.usersideapi/GetQuota"
	Usersideapi_CreateVideo_FullMethodName = "/videoapi.usersideapi/CreateVideo"
	Usersideapi_GetVideo_FullMethodName    = "/videoapi.usersideapi/GetVideo"
	Usersideapi_GetVideos_FullMethodName   = "/videoapi.usersideapi/GetVideos"
	Usersideapi_DeleteVideo_FullMethodName = "/videoapi.usersideapi/DeleteVideo"
)

// UsersideapiClient is the client API for Usersideapi service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UsersideapiClient interface {
	GetQuota(ctx context.Context, in *GetQuotaRequest, opts ...grpc.CallOption) (*QuotaResponse, error)
	CreateVideo(ctx context.Context, in *CreateVideoRequest, opts ...grpc.CallOption) (*VideoResponse, error)
	GetVideo(ctx context.Context, in *VideoRequest, opts ...grpc.CallOption) (*VideoResponse, error)
	GetVideos(ctx context.Context, in *GetVideosRequest, opts ...grpc.CallOption) (*VideosResponse, error)
	DeleteVideo(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteVideoResponse, error)
}

type usersideapiClient struct {
	cc grpc.ClientConnInterface
}

func NewUsersideapiClient(cc grpc.ClientConnInterface) UsersideapiClient {
	return &usersideapiClient{cc}
}

func (c *usersideapiClient) GetQuota(ctx context.Context, in *GetQuotaRequest, opts ...grpc.CallOption) (*QuotaResponse, error) {
	out := new(QuotaResponse)
	err := c.cc.Invoke(ctx, Usersideapi_GetQuota_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usersideapiClient) CreateVideo(ctx context.Context, in *CreateVideoRequest, opts ...grpc.CallOption) (*VideoResponse, error) {
	out := new(VideoResponse)
	err := c.cc.Invoke(ctx, Usersideapi_CreateVideo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usersideapiClient) GetVideo(ctx context.Context, in *VideoRequest, opts ...grpc.CallOption) (*VideoResponse, error) {
	out := new(VideoResponse)
	err := c.cc.Invoke(ctx, Usersideapi_GetVideo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usersideapiClient) GetVideos(ctx context.Context, in *GetVideosRequest, opts ...grpc.CallOption) (*VideosResponse, error) {
	out := new(VideosResponse)
	err := c.cc.Invoke(ctx, Usersideapi_GetVideos_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usersideapiClient) DeleteVideo(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteVideoResponse, error) {
	out := new(DeleteVideoResponse)
	err := c.cc.Invoke(ctx, Usersideapi_DeleteVideo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UsersideapiServer is the server API for Usersideapi service.
// All implementations must embed UnimplementedUsersideapiServer
// for forward compatibility
type UsersideapiServer interface {
	GetQuota(context.Context, *GetQuotaRequest) (*QuotaResponse, error)
	CreateVideo(context.Context, *CreateVideoRequest) (*VideoResponse, error)
	GetVideo(context.Context, *VideoRequest) (*VideoResponse, error)
	GetVideos(context.Context, *GetVideosRequest) (*VideosResponse, error)
	DeleteVideo(context.Context, *DeleteRequest) (*DeleteVideoResponse, error)
	mustEmbedUnimplementedUsersideapiServer()
}

// UnimplementedUsersideapiServer must be embedded to have forward compatible implementations.
type UnimplementedUsersideapiServer struct {
}

func (UnimplementedUsersideapiServer) GetQuota(context.Context, *GetQuotaRequest) (*QuotaResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetQuota not implemented")
}
func (UnimplementedUsersideapiServer) CreateVideo(context.Context, *CreateVideoRequest) (*VideoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateVideo not implemented")
}
func (UnimplementedUsersideapiServer) GetVideo(context.Context, *VideoRequest) (*VideoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetVideo not implemented")
}
func (UnimplementedUsersideapiServer) GetVideos(context.Context, *GetVideosRequest) (*VideosResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetVideos not implemented")
}
func (UnimplementedUsersideapiServer) DeleteVideo(context.Context, *DeleteRequest) (*DeleteVideoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteVideo not implemented")
}
func (UnimplementedUsersideapiServer) mustEmbedUnimplementedUsersideapiServer() {}

// UnsafeUsersideapiServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UsersideapiServer will
// result in compilation errors.
type UnsafeUsersideapiServer interface {
	mustEmbedUnimplementedUsersideapiServer()
}

func RegisterUsersideapiServer(s grpc.ServiceRegistrar, srv UsersideapiServer) {
	s.RegisterService(&Usersideapi_ServiceDesc, srv)
}

func _Usersideapi_GetQuota_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetQuotaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsersideapiServer).GetQuota(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Usersideapi_GetQuota_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsersideapiServer).GetQuota(ctx, req.(*GetQuotaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Usersideapi_CreateVideo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateVideoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsersideapiServer).CreateVideo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Usersideapi_CreateVideo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsersideapiServer).CreateVideo(ctx, req.(*CreateVideoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Usersideapi_GetVideo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VideoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsersideapiServer).GetVideo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Usersideapi_GetVideo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsersideapiServer).GetVideo(ctx, req.(*VideoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Usersideapi_GetVideos_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetVideosRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsersideapiServer).GetVideos(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Usersideapi_GetVideos_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsersideapiServer).GetVideos(ctx, req.(*GetVideosRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Usersideapi_DeleteVideo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsersideapiServer).DeleteVideo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Usersideapi_DeleteVideo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsersideapiServer).DeleteVideo(ctx, req.(*DeleteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Usersideapi_ServiceDesc is the grpc.ServiceDesc for Usersideapi service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Usersideapi_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "videoapi.usersideapi",
	HandlerType: (*UsersideapiServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetQuota",
			Handler:    _Usersideapi_GetQuota_Handler,
		},
		{
			MethodName: "CreateVideo",
			Handler:    _Usersideapi_CreateVideo_Handler,
		},
		{
			MethodName: "GetVideo",
			Handler:    _Usersideapi_GetVideo_Handler,
		},
		{
			MethodName: "GetVideos",
			Handler:    _Usersideapi_GetVideos_Handler,
		},
		{
			MethodName: "DeleteVideo",
			Handler:    _Usersideapi_DeleteVideo_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "internal/api/video/grpc/protobuf/user.proto",
}