// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v3.17.3
// source: internal/api/video/grpc/protobuf/service.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type GetByStatusRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status string `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
}

func (x *GetByStatusRequest) Reset() {
	*x = GetByStatusRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_api_video_grpc_protobuf_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetByStatusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetByStatusRequest) ProtoMessage() {}

func (x *GetByStatusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_api_video_grpc_protobuf_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetByStatusRequest.ProtoReflect.Descriptor instead.
func (*GetByStatusRequest) Descriptor() ([]byte, []int) {
	return file_internal_api_video_grpc_protobuf_service_proto_rawDescGZIP(), []int{0}
}

func (x *GetByStatusRequest) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

type VideoListResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Videos []*Video `protobuf:"bytes,1,rep,name=videos,proto3" json:"videos,omitempty"`
}

func (x *VideoListResponse) Reset() {
	*x = VideoListResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_api_video_grpc_protobuf_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *VideoListResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VideoListResponse) ProtoMessage() {}

func (x *VideoListResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_api_video_grpc_protobuf_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VideoListResponse.ProtoReflect.Descriptor instead.
func (*VideoListResponse) Descriptor() ([]byte, []int) {
	return file_internal_api_video_grpc_protobuf_service_proto_rawDescGZIP(), []int{1}
}

func (x *VideoListResponse) GetVideos() []*Video {
	if x != nil {
		return x.Videos
	}
	return nil
}

type Video struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Status    string `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
	CreatedAt string `protobuf:"bytes,3,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
}

func (x *Video) Reset() {
	*x = Video{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_api_video_grpc_protobuf_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Video) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Video) ProtoMessage() {}

func (x *Video) ProtoReflect() protoreflect.Message {
	mi := &file_internal_api_video_grpc_protobuf_service_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Video.ProtoReflect.Descriptor instead.
func (*Video) Descriptor() ([]byte, []int) {
	return file_internal_api_video_grpc_protobuf_service_proto_rawDescGZIP(), []int{2}
}

func (x *Video) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Video) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *Video) GetCreatedAt() string {
	if x != nil {
		return x.CreatedAt
	}
	return ""
}

type UpdateVideoRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id           string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Status       string `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
	Location     string `protobuf:"bytes,3,opt,name=location,proto3" json:"location,omitempty"`
	PlaybackMeta []byte `protobuf:"bytes,4,opt,name=playback_meta,json=playbackMeta,proto3" json:"playback_meta,omitempty"`
}

func (x *UpdateVideoRequest) Reset() {
	*x = UpdateVideoRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_api_video_grpc_protobuf_service_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateVideoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateVideoRequest) ProtoMessage() {}

func (x *UpdateVideoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_api_video_grpc_protobuf_service_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateVideoRequest.ProtoReflect.Descriptor instead.
func (*UpdateVideoRequest) Descriptor() ([]byte, []int) {
	return file_internal_api_video_grpc_protobuf_service_proto_rawDescGZIP(), []int{3}
}

func (x *UpdateVideoRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *UpdateVideoRequest) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *UpdateVideoRequest) GetLocation() string {
	if x != nil {
		return x.Location
	}
	return ""
}

func (x *UpdateVideoRequest) GetPlaybackMeta() []byte {
	if x != nil {
		return x.PlaybackMeta
	}
	return nil
}

type UpdateVideoResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *UpdateVideoResponse) Reset() {
	*x = UpdateVideoResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_api_video_grpc_protobuf_service_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateVideoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateVideoResponse) ProtoMessage() {}

func (x *UpdateVideoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_api_video_grpc_protobuf_service_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateVideoResponse.ProtoReflect.Descriptor instead.
func (*UpdateVideoResponse) Descriptor() ([]byte, []int) {
	return file_internal_api_video_grpc_protobuf_service_proto_rawDescGZIP(), []int{4}
}

var File_internal_api_video_grpc_protobuf_service_proto protoreflect.FileDescriptor

var file_internal_api_video_grpc_protobuf_service_proto_rawDesc = []byte{
	0x0a, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76,
	0x69, 0x64, 0x65, 0x6f, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x08, 0x76, 0x69, 0x64, 0x65, 0x6f, 0x61, 0x70, 0x69, 0x22, 0x2c, 0x0a, 0x12, 0x47, 0x65,
	0x74, 0x42, 0x79, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x3c, 0x0a, 0x11, 0x56, 0x69, 0x64, 0x65,
	0x6f, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x27, 0x0a,
	0x06, 0x76, 0x69, 0x64, 0x65, 0x6f, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e,
	0x76, 0x69, 0x64, 0x65, 0x6f, 0x61, 0x70, 0x69, 0x2e, 0x56, 0x69, 0x64, 0x65, 0x6f, 0x52, 0x06,
	0x76, 0x69, 0x64, 0x65, 0x6f, 0x73, 0x22, 0x4e, 0x0a, 0x05, 0x56, 0x69, 0x64, 0x65, 0x6f, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12,
	0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x63, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x22, 0x7d, 0x0a, 0x12, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x56, 0x69, 0x64, 0x65, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x16, 0x0a, 0x06,
	0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x23, 0x0a, 0x0d, 0x70, 0x6c, 0x61, 0x79, 0x62, 0x61, 0x63, 0x6b, 0x5f, 0x6d, 0x65, 0x74,
	0x61, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0c, 0x70, 0x6c, 0x61, 0x79, 0x62, 0x61, 0x63,
	0x6b, 0x4d, 0x65, 0x74, 0x61, 0x22, 0x15, 0x0a, 0x13, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x56,
	0x69, 0x64, 0x65, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0xa6, 0x01, 0x0a,
	0x08, 0x76, 0x69, 0x64, 0x65, 0x6f, 0x61, 0x70, 0x69, 0x12, 0x4e, 0x0a, 0x11, 0x47, 0x65, 0x74,
	0x56, 0x69, 0x64, 0x65, 0x6f, 0x73, 0x42, 0x79, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1c,
	0x2e, 0x76, 0x69, 0x64, 0x65, 0x6f, 0x61, 0x70, 0x69, 0x2e, 0x47, 0x65, 0x74, 0x42, 0x79, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x76,
	0x69, 0x64, 0x65, 0x6f, 0x61, 0x70, 0x69, 0x2e, 0x56, 0x69, 0x64, 0x65, 0x6f, 0x4c, 0x69, 0x73,
	0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x4a, 0x0a, 0x0b, 0x55, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x56, 0x69, 0x64, 0x65, 0x6f, 0x12, 0x1c, 0x2e, 0x76, 0x69, 0x64, 0x65, 0x6f,
	0x61, 0x70, 0x69, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x56, 0x69, 0x64, 0x65, 0x6f, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x76, 0x69, 0x64, 0x65, 0x6f, 0x61, 0x70,
	0x69, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x56, 0x69, 0x64, 0x65, 0x6f, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x27, 0x5a, 0x25, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61,
	0x6c, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x69, 0x64, 0x65, 0x6f, 0x2f, 0x67, 0x72, 0x70, 0x63,
	0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x70, 0x62, 0x3b, 0x70, 0x62, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_internal_api_video_grpc_protobuf_service_proto_rawDescOnce sync.Once
	file_internal_api_video_grpc_protobuf_service_proto_rawDescData = file_internal_api_video_grpc_protobuf_service_proto_rawDesc
)

func file_internal_api_video_grpc_protobuf_service_proto_rawDescGZIP() []byte {
	file_internal_api_video_grpc_protobuf_service_proto_rawDescOnce.Do(func() {
		file_internal_api_video_grpc_protobuf_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_internal_api_video_grpc_protobuf_service_proto_rawDescData)
	})
	return file_internal_api_video_grpc_protobuf_service_proto_rawDescData
}

var file_internal_api_video_grpc_protobuf_service_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_internal_api_video_grpc_protobuf_service_proto_goTypes = []interface{}{
	(*GetByStatusRequest)(nil),  // 0: videoapi.GetByStatusRequest
	(*VideoListResponse)(nil),   // 1: videoapi.VideoListResponse
	(*Video)(nil),               // 2: videoapi.Video
	(*UpdateVideoRequest)(nil),  // 3: videoapi.UpdateVideoRequest
	(*UpdateVideoResponse)(nil), // 4: videoapi.UpdateVideoResponse
}
var file_internal_api_video_grpc_protobuf_service_proto_depIdxs = []int32{
	2, // 0: videoapi.VideoListResponse.videos:type_name -> videoapi.Video
	0, // 1: videoapi.videoapi.GetVideosByStatus:input_type -> videoapi.GetByStatusRequest
	3, // 2: videoapi.videoapi.UpdateVideo:input_type -> videoapi.UpdateVideoRequest
	1, // 3: videoapi.videoapi.GetVideosByStatus:output_type -> videoapi.VideoListResponse
	4, // 4: videoapi.videoapi.UpdateVideo:output_type -> videoapi.UpdateVideoResponse
	3, // [3:5] is the sub-list for method output_type
	1, // [1:3] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_internal_api_video_grpc_protobuf_service_proto_init() }
func file_internal_api_video_grpc_protobuf_service_proto_init() {
	if File_internal_api_video_grpc_protobuf_service_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_internal_api_video_grpc_protobuf_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetByStatusRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_api_video_grpc_protobuf_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*VideoListResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_api_video_grpc_protobuf_service_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Video); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_api_video_grpc_protobuf_service_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateVideoRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_internal_api_video_grpc_protobuf_service_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateVideoResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_internal_api_video_grpc_protobuf_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_internal_api_video_grpc_protobuf_service_proto_goTypes,
		DependencyIndexes: file_internal_api_video_grpc_protobuf_service_proto_depIdxs,
		MessageInfos:      file_internal_api_video_grpc_protobuf_service_proto_msgTypes,
	}.Build()
	File_internal_api_video_grpc_protobuf_service_proto = out.File
	file_internal_api_video_grpc_protobuf_service_proto_rawDesc = nil
	file_internal_api_video_grpc_protobuf_service_proto_goTypes = nil
	file_internal_api_video_grpc_protobuf_service_proto_depIdxs = nil
}
