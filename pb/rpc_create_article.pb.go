// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.5
// 	protoc        v5.29.3
// source: rpc_create_article.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type CreateArticleRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Title         *string                `protobuf:"bytes,1,opt,name=title,proto3,oneof" json:"title,omitempty"`
	Summary       *string                `protobuf:"bytes,2,opt,name=summary,proto3,oneof" json:"summary,omitempty"`
	Content       *string                `protobuf:"bytes,3,opt,name=content,proto3,oneof" json:"content,omitempty"`
	IsPublish     *bool                  `protobuf:"varint,4,opt,name=is_publish,json=isPublish,proto3,oneof" json:"is_publish,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateArticleRequest) Reset() {
	*x = CreateArticleRequest{}
	mi := &file_rpc_create_article_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateArticleRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateArticleRequest) ProtoMessage() {}

func (x *CreateArticleRequest) ProtoReflect() protoreflect.Message {
	mi := &file_rpc_create_article_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateArticleRequest.ProtoReflect.Descriptor instead.
func (*CreateArticleRequest) Descriptor() ([]byte, []int) {
	return file_rpc_create_article_proto_rawDescGZIP(), []int{0}
}

func (x *CreateArticleRequest) GetTitle() string {
	if x != nil && x.Title != nil {
		return *x.Title
	}
	return ""
}

func (x *CreateArticleRequest) GetSummary() string {
	if x != nil && x.Summary != nil {
		return *x.Summary
	}
	return ""
}

func (x *CreateArticleRequest) GetContent() string {
	if x != nil && x.Content != nil {
		return *x.Content
	}
	return ""
}

func (x *CreateArticleRequest) GetIsPublish() bool {
	if x != nil && x.IsPublish != nil {
		return *x.IsPublish
	}
	return false
}

type CreateArticleResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Article       *Article               `protobuf:"bytes,1,opt,name=article,proto3" json:"article,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateArticleResponse) Reset() {
	*x = CreateArticleResponse{}
	mi := &file_rpc_create_article_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateArticleResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateArticleResponse) ProtoMessage() {}

func (x *CreateArticleResponse) ProtoReflect() protoreflect.Message {
	mi := &file_rpc_create_article_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateArticleResponse.ProtoReflect.Descriptor instead.
func (*CreateArticleResponse) Descriptor() ([]byte, []int) {
	return file_rpc_create_article_proto_rawDescGZIP(), []int{1}
}

func (x *CreateArticleResponse) GetArticle() *Article {
	if x != nil {
		return x.Article
	}
	return nil
}

var File_rpc_create_article_proto protoreflect.FileDescriptor

var file_rpc_create_article_proto_rawDesc = string([]byte{
	0x0a, 0x18, 0x72, 0x70, 0x63, 0x5f, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x5f, 0x61, 0x72, 0x74,
	0x69, 0x63, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x70, 0x62, 0x1a, 0x0d,
	0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xc4, 0x01,
	0x0a, 0x14, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x19, 0x0a, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x88, 0x01,
	0x01, 0x12, 0x1d, 0x0a, 0x07, 0x73, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x48, 0x01, 0x52, 0x07, 0x73, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x88, 0x01, 0x01,
	0x12, 0x1d, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x48, 0x02, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x88, 0x01, 0x01, 0x12,
	0x22, 0x0a, 0x0a, 0x69, 0x73, 0x5f, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x08, 0x48, 0x03, 0x52, 0x09, 0x69, 0x73, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68,
	0x88, 0x01, 0x01, 0x42, 0x08, 0x0a, 0x06, 0x5f, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x42, 0x0a, 0x0a,
	0x08, 0x5f, 0x73, 0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x42, 0x0a, 0x0a, 0x08, 0x5f, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x42, 0x0d, 0x0a, 0x0b, 0x5f, 0x69, 0x73, 0x5f, 0x70, 0x75, 0x62,
	0x6c, 0x69, 0x73, 0x68, 0x22, 0x3e, 0x0a, 0x15, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x72,
	0x74, 0x69, 0x63, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x25, 0x0a,
	0x07, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0b,
	0x2e, 0x70, 0x62, 0x2e, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x52, 0x07, 0x61, 0x72, 0x74,
	0x69, 0x63, 0x6c, 0x65, 0x42, 0x26, 0x5a, 0x24, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x4d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x41, 0x6c, 0x6c, 0x65, 0x6e, 0x2f,
	0x6e, 0x6f, 0x73, 0x74, 0x61, 0x6c, 0x67, 0x69, 0x61, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_rpc_create_article_proto_rawDescOnce sync.Once
	file_rpc_create_article_proto_rawDescData []byte
)

func file_rpc_create_article_proto_rawDescGZIP() []byte {
	file_rpc_create_article_proto_rawDescOnce.Do(func() {
		file_rpc_create_article_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_rpc_create_article_proto_rawDesc), len(file_rpc_create_article_proto_rawDesc)))
	})
	return file_rpc_create_article_proto_rawDescData
}

var file_rpc_create_article_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_rpc_create_article_proto_goTypes = []any{
	(*CreateArticleRequest)(nil),  // 0: pb.CreateArticleRequest
	(*CreateArticleResponse)(nil), // 1: pb.CreateArticleResponse
	(*Article)(nil),               // 2: pb.Article
}
var file_rpc_create_article_proto_depIdxs = []int32{
	2, // 0: pb.CreateArticleResponse.article:type_name -> pb.Article
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_rpc_create_article_proto_init() }
func file_rpc_create_article_proto_init() {
	if File_rpc_create_article_proto != nil {
		return
	}
	file_article_proto_init()
	file_rpc_create_article_proto_msgTypes[0].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_rpc_create_article_proto_rawDesc), len(file_rpc_create_article_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_rpc_create_article_proto_goTypes,
		DependencyIndexes: file_rpc_create_article_proto_depIdxs,
		MessageInfos:      file_rpc_create_article_proto_msgTypes,
	}.Build()
	File_rpc_create_article_proto = out.File
	file_rpc_create_article_proto_goTypes = nil
	file_rpc_create_article_proto_depIdxs = nil
}
