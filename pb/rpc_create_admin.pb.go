// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.5
// 	protoc        v5.29.3
// source: rpc_create_admin.proto

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

type CreateAdminRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Password      string                 `protobuf:"bytes,2,opt,name=password,proto3" json:"password,omitempty"`
	IsActive      *bool                  `protobuf:"varint,3,opt,name=is_active,json=isActive,proto3,oneof" json:"is_active,omitempty"`
	RoleId        *int64                 `protobuf:"varint,4,opt,name=role_id,json=roleId,proto3,oneof" json:"role_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateAdminRequest) Reset() {
	*x = CreateAdminRequest{}
	mi := &file_rpc_create_admin_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateAdminRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateAdminRequest) ProtoMessage() {}

func (x *CreateAdminRequest) ProtoReflect() protoreflect.Message {
	mi := &file_rpc_create_admin_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateAdminRequest.ProtoReflect.Descriptor instead.
func (*CreateAdminRequest) Descriptor() ([]byte, []int) {
	return file_rpc_create_admin_proto_rawDescGZIP(), []int{0}
}

func (x *CreateAdminRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *CreateAdminRequest) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

func (x *CreateAdminRequest) GetIsActive() bool {
	if x != nil && x.IsActive != nil {
		return *x.IsActive
	}
	return false
}

func (x *CreateAdminRequest) GetRoleId() int64 {
	if x != nil && x.RoleId != nil {
		return *x.RoleId
	}
	return 0
}

type CreateAdminResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Admin         *Admin                 `protobuf:"bytes,1,opt,name=admin,proto3" json:"admin,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateAdminResponse) Reset() {
	*x = CreateAdminResponse{}
	mi := &file_rpc_create_admin_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateAdminResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateAdminResponse) ProtoMessage() {}

func (x *CreateAdminResponse) ProtoReflect() protoreflect.Message {
	mi := &file_rpc_create_admin_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateAdminResponse.ProtoReflect.Descriptor instead.
func (*CreateAdminResponse) Descriptor() ([]byte, []int) {
	return file_rpc_create_admin_proto_rawDescGZIP(), []int{1}
}

func (x *CreateAdminResponse) GetAdmin() *Admin {
	if x != nil {
		return x.Admin
	}
	return nil
}

var File_rpc_create_admin_proto protoreflect.FileDescriptor

var file_rpc_create_admin_proto_rawDesc = string([]byte{
	0x0a, 0x16, 0x72, 0x70, 0x63, 0x5f, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x5f, 0x61, 0x64, 0x6d,
	0x69, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x70, 0x62, 0x1a, 0x0b, 0x61, 0x64,
	0x6d, 0x69, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x9e, 0x01, 0x0a, 0x12, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64,
	0x12, 0x20, 0x0a, 0x09, 0x69, 0x73, 0x5f, 0x61, 0x63, 0x74, 0x69, 0x76, 0x65, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x08, 0x48, 0x00, 0x52, 0x08, 0x69, 0x73, 0x41, 0x63, 0x74, 0x69, 0x76, 0x65, 0x88,
	0x01, 0x01, 0x12, 0x1c, 0x0a, 0x07, 0x72, 0x6f, 0x6c, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x03, 0x48, 0x01, 0x52, 0x06, 0x72, 0x6f, 0x6c, 0x65, 0x49, 0x64, 0x88, 0x01, 0x01,
	0x42, 0x0c, 0x0a, 0x0a, 0x5f, 0x69, 0x73, 0x5f, 0x61, 0x63, 0x74, 0x69, 0x76, 0x65, 0x42, 0x0a,
	0x0a, 0x08, 0x5f, 0x72, 0x6f, 0x6c, 0x65, 0x5f, 0x69, 0x64, 0x22, 0x36, 0x0a, 0x13, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x1f, 0x0a, 0x05, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x09, 0x2e, 0x70, 0x62, 0x2e, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x52, 0x05, 0x61, 0x64, 0x6d,
	0x69, 0x6e, 0x42, 0x26, 0x5a, 0x24, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x4d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x41, 0x6c, 0x6c, 0x65, 0x6e, 0x2f, 0x6e, 0x6f,
	0x73, 0x74, 0x61, 0x6c, 0x67, 0x69, 0x61, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
})

var (
	file_rpc_create_admin_proto_rawDescOnce sync.Once
	file_rpc_create_admin_proto_rawDescData []byte
)

func file_rpc_create_admin_proto_rawDescGZIP() []byte {
	file_rpc_create_admin_proto_rawDescOnce.Do(func() {
		file_rpc_create_admin_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_rpc_create_admin_proto_rawDesc), len(file_rpc_create_admin_proto_rawDesc)))
	})
	return file_rpc_create_admin_proto_rawDescData
}

var file_rpc_create_admin_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_rpc_create_admin_proto_goTypes = []any{
	(*CreateAdminRequest)(nil),  // 0: pb.CreateAdminRequest
	(*CreateAdminResponse)(nil), // 1: pb.CreateAdminResponse
	(*Admin)(nil),               // 2: pb.Admin
}
var file_rpc_create_admin_proto_depIdxs = []int32{
	2, // 0: pb.CreateAdminResponse.admin:type_name -> pb.Admin
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_rpc_create_admin_proto_init() }
func file_rpc_create_admin_proto_init() {
	if File_rpc_create_admin_proto != nil {
		return
	}
	file_admin_proto_init()
	file_rpc_create_admin_proto_msgTypes[0].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_rpc_create_admin_proto_rawDesc), len(file_rpc_create_admin_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_rpc_create_admin_proto_goTypes,
		DependencyIndexes: file_rpc_create_admin_proto_depIdxs,
		MessageInfos:      file_rpc_create_admin_proto_msgTypes,
	}.Build()
	File_rpc_create_admin_proto = out.File
	file_rpc_create_admin_proto_goTypes = nil
	file_rpc_create_admin_proto_depIdxs = nil
}
