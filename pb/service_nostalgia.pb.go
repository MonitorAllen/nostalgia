// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.5
// 	protoc        v5.29.3
// source: service_nostalgia.proto

package pb

import (
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_service_nostalgia_proto protoreflect.FileDescriptor

var file_service_nostalgia_proto_rawDesc = string([]byte{
	0x0a, 0x17, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x6e, 0x6f, 0x73, 0x74, 0x61, 0x6c,
	0x67, 0x69, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x70, 0x62, 0x1a, 0x1c, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x76,
	0x32, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x16, 0x72, 0x70, 0x63,
	0x5f, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x5f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x16, 0x72, 0x70, 0x63, 0x5f, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x5f,
	0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x15, 0x72, 0x70, 0x63,
	0x5f, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x5f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x16, 0x72, 0x70, 0x63, 0x5f, 0x6c, 0x6f, 0x67, 0x6f, 0x75, 0x74, 0x5f, 0x61,
	0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x14, 0x72, 0x70, 0x63, 0x5f,
	0x61, 0x64, 0x6d, 0x69, 0x6e, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x17, 0x72, 0x70, 0x63, 0x5f, 0x69, 0x6e, 0x69, 0x74, 0x5f, 0x73, 0x79, 0x73, 0x5f, 0x6d,
	0x65, 0x6e, 0x75, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x18, 0x72, 0x70, 0x63, 0x5f, 0x63,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x5f, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x18, 0x72, 0x70, 0x63, 0x5f, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x5f,
	0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x72,
	0x70, 0x63, 0x5f, 0x6c, 0x69, 0x73, 0x74, 0x5f, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x15, 0x72, 0x70, 0x63, 0x5f, 0x67, 0x65, 0x74, 0x5f,
	0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x18, 0x72,
	0x70, 0x63, 0x5f, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x5f, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x72, 0x70, 0x63, 0x5f, 0x72, 0x65, 0x6e,
	0x65, 0x77, 0x5f, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x15, 0x72, 0x70, 0x63, 0x5f, 0x75, 0x70, 0x6c, 0x6f, 0x61,
	0x64, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x32, 0xfd, 0x0f, 0x0a,
	0x09, 0x4e, 0x6f, 0x73, 0x74, 0x61, 0x6c, 0x67, 0x69, 0x61, 0x12, 0x8e, 0x01, 0x0a, 0x0a, 0x4c,
	0x6f, 0x67, 0x69, 0x6e, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x12, 0x15, 0x2e, 0x70, 0x62, 0x2e, 0x4c,
	0x6f, 0x67, 0x69, 0x6e, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x16, 0x2e, 0x70, 0x62, 0x2e, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x41, 0x64, 0x6d, 0x69, 0x6e,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x51, 0x92, 0x41, 0x34, 0x0a, 0x04, 0x41,
	0x75, 0x74, 0x68, 0x12, 0x0d, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x20, 0x62, 0x61, 0x63, 0x6b, 0x65,
	0x6e, 0x64, 0x1a, 0x1d, 0x55, 0x73, 0x65, 0x20, 0x74, 0x68, 0x69, 0x73, 0x20, 0x41, 0x50, 0x49,
	0x20, 0x74, 0x6f, 0x20, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x20, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e,
	0x64, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x14, 0x3a, 0x01, 0x2a, 0x22, 0x0f, 0x2f, 0x76, 0x31, 0x2f,
	0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x12, 0x94, 0x01, 0x0a, 0x0b,
	0x4c, 0x6f, 0x67, 0x6f, 0x75, 0x74, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x12, 0x16, 0x2e, 0x70, 0x62,
	0x2e, 0x4c, 0x6f, 0x67, 0x6f, 0x75, 0x74, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x70, 0x62, 0x2e, 0x4c, 0x6f, 0x67, 0x6f, 0x75, 0x74, 0x41,
	0x64, 0x6d, 0x69, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x54, 0x92, 0x41,
	0x36, 0x0a, 0x04, 0x41, 0x75, 0x74, 0x68, 0x12, 0x0e, 0x6c, 0x6f, 0x67, 0x6f, 0x75, 0x74, 0x20,
	0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x1a, 0x1e, 0x55, 0x73, 0x65, 0x20, 0x74, 0x68, 0x69,
	0x73, 0x20, 0x41, 0x50, 0x49, 0x20, 0x74, 0x6f, 0x20, 0x6c, 0x6f, 0x67, 0x6f, 0x75, 0x74, 0x20,
	0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x15, 0x3a, 0x01, 0x2a,
	0x22, 0x10, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x6c, 0x6f, 0x67, 0x6f,
	0x75, 0x74, 0x12, 0xb3, 0x01, 0x0a, 0x10, 0x52, 0x65, 0x6e, 0x65, 0x77, 0x41, 0x63, 0x63, 0x65,
	0x73, 0x73, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x1b, 0x2e, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x6e,
	0x65, 0x77, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x6e, 0x65, 0x77, 0x41,
	0x63, 0x63, 0x65, 0x73, 0x73, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x64, 0x92, 0x41, 0x40, 0x0a, 0x04, 0x41, 0x75, 0x74, 0x68, 0x12, 0x12, 0x72,
	0x65, 0x6e, 0x65, 0x77, 0x20, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x20, 0x74, 0x6f, 0x6b, 0x65,
	0x6e, 0x1a, 0x24, 0x55, 0x73, 0x65, 0x20, 0x74, 0x68, 0x69, 0x73, 0x20, 0x41, 0x50, 0x49, 0x20,
	0x74, 0x6f, 0x20, 0x72, 0x65, 0x6e, 0x65, 0x77, 0x20, 0x61, 0x20, 0x61, 0x63, 0x63, 0x65, 0x73,
	0x73, 0x20, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1b, 0x3a, 0x01, 0x2a,
	0x22, 0x16, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x72, 0x65, 0x6e, 0x65,
	0x77, 0x5f, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12, 0x92, 0x01, 0x0a, 0x0b, 0x49, 0x6e, 0x69,
	0x74, 0x53, 0x79, 0x73, 0x4d, 0x65, 0x6e, 0x75, 0x12, 0x16, 0x2e, 0x70, 0x62, 0x2e, 0x49, 0x6e,
	0x69, 0x74, 0x53, 0x79, 0x73, 0x4d, 0x65, 0x6e, 0x75, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x17, 0x2e, 0x70, 0x62, 0x2e, 0x49, 0x6e, 0x69, 0x74, 0x53, 0x79, 0x73, 0x4d, 0x65, 0x6e,
	0x75, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x52, 0x92, 0x41, 0x3a, 0x0a, 0x04,
	0x4d, 0x65, 0x6e, 0x75, 0x12, 0x10, 0x69, 0x6e, 0x69, 0x74, 0x20, 0x73, 0x79, 0x73, 0x74, 0x65,
	0x6d, 0x20, 0x6d, 0x65, 0x6e, 0x75, 0x1a, 0x20, 0x55, 0x73, 0x65, 0x20, 0x74, 0x68, 0x69, 0x73,
	0x20, 0x41, 0x50, 0x49, 0x20, 0x74, 0x6f, 0x20, 0x69, 0x6e, 0x69, 0x74, 0x20, 0x73, 0x79, 0x73,
	0x74, 0x65, 0x6d, 0x20, 0x6d, 0x65, 0x6e, 0x75, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0f, 0x12, 0x0d,
	0x2f, 0x76, 0x31, 0x2f, 0x6d, 0x65, 0x6e, 0x75, 0x2f, 0x69, 0x6e, 0x69, 0x74, 0x12, 0x94, 0x01,
	0x0a, 0x0b, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x12, 0x16, 0x2e,
	0x70, 0x62, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x70, 0x62, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x54,
	0x92, 0x41, 0x3d, 0x0a, 0x05, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x12, 0x10, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x20, 0x6e, 0x65, 0x77, 0x20, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x1a, 0x22, 0x55, 0x73,
	0x65, 0x20, 0x74, 0x68, 0x69, 0x73, 0x20, 0x41, 0x50, 0x49, 0x20, 0x74, 0x6f, 0x20, 0x63, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x20, 0x61, 0x20, 0x6e, 0x65, 0x77, 0x20, 0x61, 0x64, 0x6d, 0x69, 0x6e,
	0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0e, 0x3a, 0x01, 0x2a, 0x22, 0x09, 0x2f, 0x76, 0x31, 0x2f, 0x61,
	0x64, 0x6d, 0x69, 0x6e, 0x12, 0x94, 0x01, 0x0a, 0x0b, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x41,
	0x64, 0x6d, 0x69, 0x6e, 0x12, 0x16, 0x2e, 0x70, 0x62, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x41, 0x64, 0x6d, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x70,
	0x62, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x54, 0x92, 0x41, 0x3d, 0x0a, 0x05, 0x41, 0x64, 0x6d, 0x69,
	0x6e, 0x12, 0x10, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x20, 0x6e, 0x65, 0x77, 0x20, 0x61, 0x64,
	0x6d, 0x69, 0x6e, 0x1a, 0x22, 0x55, 0x73, 0x65, 0x20, 0x74, 0x68, 0x69, 0x73, 0x20, 0x41, 0x50,
	0x49, 0x20, 0x74, 0x6f, 0x20, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x20, 0x61, 0x20, 0x6e, 0x65,
	0x77, 0x20, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0e, 0x3a, 0x01, 0x2a,
	0x32, 0x09, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x12, 0x8a, 0x01, 0x0a, 0x09,
	0x41, 0x64, 0x6d, 0x69, 0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x14, 0x2e, 0x70, 0x62, 0x2e, 0x41,
	0x64, 0x6d, 0x69, 0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x15, 0x2e, 0x70, 0x62, 0x2e, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x50, 0x92, 0x41, 0x37, 0x0a, 0x05, 0x41, 0x64, 0x6d,
	0x69, 0x6e, 0x12, 0x0e, 0x67, 0x65, 0x74, 0x20, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x20, 0x69, 0x6e,
	0x66, 0x6f, 0x1a, 0x1e, 0x55, 0x73, 0x65, 0x20, 0x74, 0x68, 0x69, 0x73, 0x20, 0x41, 0x50, 0x49,
	0x20, 0x74, 0x6f, 0x20, 0x67, 0x65, 0x74, 0x20, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x20, 0x69, 0x6e,
	0x66, 0x6f, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x10, 0x12, 0x0e, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x64,
	0x6d, 0x69, 0x6e, 0x2f, 0x69, 0x6e, 0x66, 0x6f, 0x12, 0xad, 0x01, 0x0a, 0x0d, 0x43, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x12, 0x18, 0x2e, 0x70, 0x62, 0x2e,
	0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x70, 0x62, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x67, 0x92, 0x41, 0x4d, 0x0a, 0x07, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x12, 0x15, 0x63,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x20, 0x61, 0x6e, 0x20, 0x6e, 0x65, 0x77, 0x20, 0x61, 0x72, 0x74,
	0x69, 0x63, 0x6c, 0x65, 0x1a, 0x2b, 0x55, 0x73, 0x65, 0x20, 0x74, 0x68, 0x69, 0x73, 0x20, 0x41,
	0x50, 0x49, 0x20, 0x74, 0x6f, 0x20, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x20, 0x63, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x20, 0x61, 0x20, 0x6e, 0x65, 0x77, 0x20, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c,
	0x65, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x11, 0x3a, 0x01, 0x2a, 0x22, 0x0c, 0x2f, 0x76, 0x31, 0x2f,
	0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x73, 0x12, 0x9e, 0x01, 0x0a, 0x0d, 0x44, 0x65, 0x6c,
	0x65, 0x74, 0x65, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x12, 0x18, 0x2e, 0x70, 0x62, 0x2e,
	0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x70, 0x62, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65,
	0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x58, 0x92, 0x41, 0x3c, 0x0a, 0x07, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x12, 0x11, 0x64,
	0x65, 0x6c, 0x65, 0x74, 0x65, 0x20, 0x61, 0x6e, 0x20, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65,
	0x1a, 0x1e, 0x55, 0x73, 0x65, 0x20, 0x74, 0x68, 0x69, 0x73, 0x20, 0x41, 0x50, 0x49, 0x20, 0x74,
	0x6f, 0x20, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x20, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65,
	0x82, 0xd3, 0xe4, 0x93, 0x02, 0x13, 0x2a, 0x11, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x72, 0x74, 0x69,
	0x63, 0x6c, 0x65, 0x73, 0x2f, 0x7b, 0x69, 0x64, 0x7d, 0x12, 0x94, 0x01, 0x0a, 0x0c, 0x4c, 0x69,
	0x73, 0x74, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x73, 0x12, 0x17, 0x2e, 0x70, 0x62, 0x2e,
	0x4c, 0x69, 0x73, 0x74, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x70, 0x62, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x41, 0x72, 0x74,
	0x69, 0x63, 0x6c, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x51, 0x92,
	0x41, 0x3a, 0x0a, 0x07, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x12, 0x0d, 0x6c, 0x69, 0x73,
	0x74, 0x20, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x73, 0x1a, 0x20, 0x55, 0x73, 0x65, 0x20,
	0x74, 0x68, 0x69, 0x73, 0x20, 0x41, 0x50, 0x49, 0x20, 0x74, 0x6f, 0x20, 0x67, 0x65, 0x74, 0x20,
	0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x20, 0x6c, 0x69, 0x73, 0x74, 0x82, 0xd3, 0xe4, 0x93,
	0x02, 0x0e, 0x12, 0x0c, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x73,
	0x12, 0xa5, 0x01, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x12,
	0x15, 0x2e, 0x70, 0x62, 0x2e, 0x47, 0x65, 0x74, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x70, 0x62, 0x2e, 0x47, 0x65, 0x74, 0x41,
	0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x68,
	0x92, 0x41, 0x3d, 0x0a, 0x07, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x12, 0x10, 0x67, 0x65,
	0x74, 0x20, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x20, 0x69, 0x6e, 0x66, 0x6f, 0x1a, 0x20,
	0x55, 0x73, 0x65, 0x20, 0x74, 0x68, 0x69, 0x73, 0x20, 0x41, 0x50, 0x49, 0x20, 0x74, 0x6f, 0x20,
	0x67, 0x65, 0x74, 0x20, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x20, 0x69, 0x6e, 0x66, 0x6f,
	0x82, 0xd3, 0xe4, 0x93, 0x02, 0x22, 0x12, 0x20, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x72, 0x74, 0x69,
	0x63, 0x6c, 0x65, 0x73, 0x2f, 0x7b, 0x69, 0x64, 0x7d, 0x2f, 0x7b, 0x6e, 0x65, 0x65, 0x64, 0x5f,
	0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x7d, 0x12, 0xa3, 0x01, 0x0a, 0x0d, 0x55, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x12, 0x18, 0x2e, 0x70, 0x62, 0x2e,
	0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x70, 0x62, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x5d, 0x92, 0x41, 0x43, 0x0a, 0x07, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x12, 0x13, 0x75,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x20, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x20, 0x69, 0x6e,
	0x66, 0x6f, 0x1a, 0x23, 0x55, 0x73, 0x65, 0x20, 0x74, 0x68, 0x69, 0x73, 0x20, 0x41, 0x50, 0x49,
	0x20, 0x74, 0x6f, 0x20, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x20, 0x61, 0x72, 0x74, 0x69, 0x63,
	0x6c, 0x65, 0x20, 0x69, 0x6e, 0x66, 0x6f, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x11, 0x3a, 0x01, 0x2a,
	0x32, 0x0c, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x73, 0x12, 0x8b,
	0x01, 0x0a, 0x0a, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x15, 0x2e,
	0x70, 0x62, 0x2e, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x70, 0x62, 0x2e, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64,
	0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x4e, 0x92, 0x41,
	0x2c, 0x0a, 0x04, 0x55, 0x74, 0x69, 0x6c, 0x12, 0x0b, 0x75, 0x70, 0x6c, 0x61, 0x6f, 0x64, 0x20,
	0x66, 0x69, 0x6c, 0x65, 0x1a, 0x17, 0x55, 0x73, 0x65, 0x20, 0x74, 0x68, 0x69, 0x73, 0x20, 0x41,
	0x50, 0x49, 0x20, 0x75, 0x6c, 0x6f, 0x61, 0x64, 0x20, 0x66, 0x69, 0x6c, 0x65, 0x82, 0xd3, 0xe4,
	0x93, 0x02, 0x19, 0x3a, 0x01, 0x2a, 0x22, 0x14, 0x2f, 0x76, 0x31, 0x2f, 0x75, 0x74, 0x69, 0x6c,
	0x2f, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x42, 0x9b, 0x01, 0x92,
	0x41, 0x72, 0x12, 0x70, 0x0a, 0x0d, 0x4e, 0x6f, 0x73, 0x74, 0x61, 0x6c, 0x67, 0x69, 0x61, 0x20,
	0x41, 0x50, 0x49, 0x22, 0x5a, 0x0a, 0x1b, 0x4d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x41, 0x6c,
	0x6c, 0x65, 0x6e, 0x20, 0x42, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x20, 0x4d, 0x61, 0x73, 0x74,
	0x65, 0x72, 0x12, 0x1f, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x4d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x41, 0x6c,
	0x6c, 0x65, 0x6e, 0x1a, 0x1a, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x61, 0x6c, 0x6c, 0x65,
	0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x40, 0x67, 0x6d, 0x61, 0x69, 0x6c, 0x2e, 0x63, 0x6f, 0x6d, 0x32,
	0x03, 0x31, 0x2e, 0x30, 0x5a, 0x24, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x4d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x41, 0x6c, 0x6c, 0x65, 0x6e, 0x2f, 0x6e, 0x6f,
	0x73, 0x74, 0x61, 0x6c, 0x67, 0x69, 0x61, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
})

var file_service_nostalgia_proto_goTypes = []any{
	(*LoginAdminRequest)(nil),        // 0: pb.LoginAdminRequest
	(*LogoutAdminRequest)(nil),       // 1: pb.LogoutAdminRequest
	(*RenewAccessTokenRequest)(nil),  // 2: pb.RenewAccessTokenRequest
	(*InitSysMenuRequest)(nil),       // 3: pb.InitSysMenuRequest
	(*CreateAdminRequest)(nil),       // 4: pb.CreateAdminRequest
	(*UpdateAdminRequest)(nil),       // 5: pb.UpdateAdminRequest
	(*AdminInfoRequest)(nil),         // 6: pb.AdminInfoRequest
	(*CreateArticleRequest)(nil),     // 7: pb.CreateArticleRequest
	(*DeleteArticleRequest)(nil),     // 8: pb.DeleteArticleRequest
	(*ListArticlesRequest)(nil),      // 9: pb.ListArticlesRequest
	(*GetArticleRequest)(nil),        // 10: pb.GetArticleRequest
	(*UpdateArticleRequest)(nil),     // 11: pb.UpdateArticleRequest
	(*UploadFileRequest)(nil),        // 12: pb.UploadFileRequest
	(*LoginAdminResponse)(nil),       // 13: pb.LoginAdminResponse
	(*LogoutAdminResponse)(nil),      // 14: pb.LogoutAdminResponse
	(*RenewAccessTokenResponse)(nil), // 15: pb.RenewAccessTokenResponse
	(*InitSysMenuResponse)(nil),      // 16: pb.InitSysMenuResponse
	(*CreateAdminResponse)(nil),      // 17: pb.CreateAdminResponse
	(*UpdateAdminResponse)(nil),      // 18: pb.UpdateAdminResponse
	(*AdminInfoResponse)(nil),        // 19: pb.AdminInfoResponse
	(*CreateArticleResponse)(nil),    // 20: pb.CreateArticleResponse
	(*DeleteArticleResponse)(nil),    // 21: pb.DeleteArticleResponse
	(*ListArticlesResponse)(nil),     // 22: pb.ListArticlesResponse
	(*GetArticleResponse)(nil),       // 23: pb.GetArticleResponse
	(*UpdateArticleResponse)(nil),    // 24: pb.UpdateArticleResponse
	(*UploadFileResponse)(nil),       // 25: pb.UploadFileResponse
}
var file_service_nostalgia_proto_depIdxs = []int32{
	0,  // 0: pb.Nostalgia.LoginAdmin:input_type -> pb.LoginAdminRequest
	1,  // 1: pb.Nostalgia.LogoutAdmin:input_type -> pb.LogoutAdminRequest
	2,  // 2: pb.Nostalgia.RenewAccessToken:input_type -> pb.RenewAccessTokenRequest
	3,  // 3: pb.Nostalgia.InitSysMenu:input_type -> pb.InitSysMenuRequest
	4,  // 4: pb.Nostalgia.CreateAdmin:input_type -> pb.CreateAdminRequest
	5,  // 5: pb.Nostalgia.UpdateAdmin:input_type -> pb.UpdateAdminRequest
	6,  // 6: pb.Nostalgia.AdminInfo:input_type -> pb.AdminInfoRequest
	7,  // 7: pb.Nostalgia.CreateArticle:input_type -> pb.CreateArticleRequest
	8,  // 8: pb.Nostalgia.DeleteArticle:input_type -> pb.DeleteArticleRequest
	9,  // 9: pb.Nostalgia.ListArticles:input_type -> pb.ListArticlesRequest
	10, // 10: pb.Nostalgia.GetArticle:input_type -> pb.GetArticleRequest
	11, // 11: pb.Nostalgia.UpdateArticle:input_type -> pb.UpdateArticleRequest
	12, // 12: pb.Nostalgia.UploadFile:input_type -> pb.UploadFileRequest
	13, // 13: pb.Nostalgia.LoginAdmin:output_type -> pb.LoginAdminResponse
	14, // 14: pb.Nostalgia.LogoutAdmin:output_type -> pb.LogoutAdminResponse
	15, // 15: pb.Nostalgia.RenewAccessToken:output_type -> pb.RenewAccessTokenResponse
	16, // 16: pb.Nostalgia.InitSysMenu:output_type -> pb.InitSysMenuResponse
	17, // 17: pb.Nostalgia.CreateAdmin:output_type -> pb.CreateAdminResponse
	18, // 18: pb.Nostalgia.UpdateAdmin:output_type -> pb.UpdateAdminResponse
	19, // 19: pb.Nostalgia.AdminInfo:output_type -> pb.AdminInfoResponse
	20, // 20: pb.Nostalgia.CreateArticle:output_type -> pb.CreateArticleResponse
	21, // 21: pb.Nostalgia.DeleteArticle:output_type -> pb.DeleteArticleResponse
	22, // 22: pb.Nostalgia.ListArticles:output_type -> pb.ListArticlesResponse
	23, // 23: pb.Nostalgia.GetArticle:output_type -> pb.GetArticleResponse
	24, // 24: pb.Nostalgia.UpdateArticle:output_type -> pb.UpdateArticleResponse
	25, // 25: pb.Nostalgia.UploadFile:output_type -> pb.UploadFileResponse
	13, // [13:26] is the sub-list for method output_type
	0,  // [0:13] is the sub-list for method input_type
	0,  // [0:0] is the sub-list for extension type_name
	0,  // [0:0] is the sub-list for extension extendee
	0,  // [0:0] is the sub-list for field type_name
}

func init() { file_service_nostalgia_proto_init() }
func file_service_nostalgia_proto_init() {
	if File_service_nostalgia_proto != nil {
		return
	}
	file_rpc_create_admin_proto_init()
	file_rpc_update_admin_proto_init()
	file_rpc_login_admin_proto_init()
	file_rpc_logout_admin_proto_init()
	file_rpc_admin_info_proto_init()
	file_rpc_init_sys_menu_proto_init()
	file_rpc_create_article_proto_init()
	file_rpc_delete_article_proto_init()
	file_rpc_list_articles_proto_init()
	file_rpc_get_article_proto_init()
	file_rpc_update_article_proto_init()
	file_rpc_renew_access_token_proto_init()
	file_rpc_upload_file_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_service_nostalgia_proto_rawDesc), len(file_service_nostalgia_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_service_nostalgia_proto_goTypes,
		DependencyIndexes: file_service_nostalgia_proto_depIdxs,
	}.Build()
	File_service_nostalgia_proto = out.File
	file_service_nostalgia_proto_goTypes = nil
	file_service_nostalgia_proto_depIdxs = nil
}
