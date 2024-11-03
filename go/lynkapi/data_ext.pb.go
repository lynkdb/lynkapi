// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v3.20.3
// source: lynkapi/data_ext.proto

package lynkapi

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

type DataExt_BaseItem struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id     string `protobuf:"bytes,3,opt,name=id,proto3" json:"id,omitempty" toml:"id,omitempty" yaml:"id,omitempty" x_attrs:"primary_key,object_id(12)"`
	Pid    string `protobuf:"bytes,4,opt,name=pid,proto3" json:"pid,omitempty" toml:"pid,omitempty" yaml:"pid,omitempty"`                     // parent-id
	UserId string `protobuf:"bytes,5,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty" toml:"user_id,omitempty" yaml:"user_id,omitempty"`
	DictId string `protobuf:"bytes,8,opt,name=dict_id,json=dictId,proto3" json:"dict_id,omitempty" toml:"dict_id,omitempty" yaml:"dict_id,omitempty"`
	DictNs string `protobuf:"bytes,9,opt,name=dict_ns,json=dictNs,proto3" json:"dict_ns,omitempty" toml:"dict_ns,omitempty" yaml:"dict_ns,omitempty"`
}

func (x *DataExt_BaseItem) Reset() {
	*x = DataExt_BaseItem{}
	if protoimpl.UnsafeEnabled {
		mi := &file_lynkapi_data_ext_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DataExt_BaseItem) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DataExt_BaseItem) ProtoMessage() {}

func (x *DataExt_BaseItem) ProtoReflect() protoreflect.Message {
	mi := &file_lynkapi_data_ext_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DataExt_BaseItem.ProtoReflect.Descriptor instead.
func (*DataExt_BaseItem) Descriptor() ([]byte, []int) {
	return file_lynkapi_data_ext_proto_rawDescGZIP(), []int{0}
}

func (x *DataExt_BaseItem) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *DataExt_BaseItem) GetPid() string {
	if x != nil {
		return x.Pid
	}
	return ""
}

func (x *DataExt_BaseItem) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *DataExt_BaseItem) GetDictId() string {
	if x != nil {
		return x.DictId
	}
	return ""
}

func (x *DataExt_BaseItem) GetDictNs() string {
	if x != nil {
		return x.DictNs
	}
	return ""
}

var File_lynkapi_data_ext_proto protoreflect.FileDescriptor

var file_lynkapi_data_ext_proto_rawDesc = []byte{
	0x0a, 0x16, 0x6c, 0x79, 0x6e, 0x6b, 0x61, 0x70, 0x69, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x5f, 0x65,
	0x78, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x6c, 0x79, 0x6e, 0x6b, 0x61, 0x70,
	0x69, 0x22, 0x7f, 0x0a, 0x10, 0x44, 0x61, 0x74, 0x61, 0x45, 0x78, 0x74, 0x5f, 0x42, 0x61, 0x73,
	0x65, 0x49, 0x74, 0x65, 0x6d, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x70, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x70, 0x69, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x75, 0x73, 0x65, 0x72, 0x5f,
	0x69, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x64,
	0x12, 0x17, 0x0a, 0x07, 0x64, 0x69, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x08, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x64, 0x69, 0x63, 0x74, 0x49, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x64, 0x69, 0x63,
	0x74, 0x5f, 0x6e, 0x73, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x64, 0x69, 0x63, 0x74,
	0x4e, 0x73, 0x42, 0x30, 0x48, 0x03, 0x5a, 0x2c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x6c, 0x79, 0x6e, 0x6b, 0x64, 0x62, 0x2f, 0x6c, 0x79, 0x6e, 0x6b, 0x61, 0x70,
	0x69, 0x2f, 0x67, 0x6f, 0x2f, 0x6c, 0x79, 0x6e, 0x6b, 0x61, 0x70, 0x69, 0x3b, 0x6c, 0x79, 0x6e,
	0x6b, 0x61, 0x70, 0x69, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_lynkapi_data_ext_proto_rawDescOnce sync.Once
	file_lynkapi_data_ext_proto_rawDescData = file_lynkapi_data_ext_proto_rawDesc
)

func file_lynkapi_data_ext_proto_rawDescGZIP() []byte {
	file_lynkapi_data_ext_proto_rawDescOnce.Do(func() {
		file_lynkapi_data_ext_proto_rawDescData = protoimpl.X.CompressGZIP(file_lynkapi_data_ext_proto_rawDescData)
	})
	return file_lynkapi_data_ext_proto_rawDescData
}

var file_lynkapi_data_ext_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_lynkapi_data_ext_proto_goTypes = []interface{}{
	(*DataExt_BaseItem)(nil), // 0: lynkapi.DataExt_BaseItem
}
var file_lynkapi_data_ext_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_lynkapi_data_ext_proto_init() }
func file_lynkapi_data_ext_proto_init() {
	if File_lynkapi_data_ext_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_lynkapi_data_ext_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DataExt_BaseItem); i {
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
			RawDescriptor: file_lynkapi_data_ext_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_lynkapi_data_ext_proto_goTypes,
		DependencyIndexes: file_lynkapi_data_ext_proto_depIdxs,
		MessageInfos:      file_lynkapi_data_ext_proto_msgTypes,
	}.Build()
	File_lynkapi_data_ext_proto = out.File
	file_lynkapi_data_ext_proto_rawDesc = nil
	file_lynkapi_data_ext_proto_goTypes = nil
	file_lynkapi_data_ext_proto_depIdxs = nil
}
