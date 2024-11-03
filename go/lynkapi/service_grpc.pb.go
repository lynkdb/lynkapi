// Copyright 2024 Eryx <evorui at gmail dot com>, All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.20.3
// source: lynkapi/service.proto

package lynkapi

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
	LynkService_ApiList_FullMethodName     = "/lynkapi.LynkService/ApiList"
	LynkService_Exec_FullMethodName        = "/lynkapi.LynkService/Exec"
	LynkService_DataProject_FullMethodName = "/lynkapi.LynkService/DataProject"
	LynkService_DataQuery_FullMethodName   = "/lynkapi.LynkService/DataQuery"
	LynkService_DataUpsert_FullMethodName  = "/lynkapi.LynkService/DataUpsert"
	LynkService_DataIgsert_FullMethodName  = "/lynkapi.LynkService/DataIgsert"
)

// LynkServiceClient is the client API for LynkService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LynkServiceClient interface {
	ApiList(ctx context.Context, in *ApiListRequest, opts ...grpc.CallOption) (*ApiListResponse, error)
	Exec(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
	DataProject(ctx context.Context, in *DataProjectRequest, opts ...grpc.CallOption) (*DataProjectResponse, error)
	DataQuery(ctx context.Context, in *DataQuery, opts ...grpc.CallOption) (*DataResult, error)
	DataUpsert(ctx context.Context, in *DataInsert, opts ...grpc.CallOption) (*DataResult, error)
	DataIgsert(ctx context.Context, in *DataInsert, opts ...grpc.CallOption) (*DataResult, error)
}

type lynkServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewLynkServiceClient(cc grpc.ClientConnInterface) LynkServiceClient {
	return &lynkServiceClient{cc}
}

func (c *lynkServiceClient) ApiList(ctx context.Context, in *ApiListRequest, opts ...grpc.CallOption) (*ApiListResponse, error) {
	out := new(ApiListResponse)
	err := c.cc.Invoke(ctx, LynkService_ApiList_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *lynkServiceClient) Exec(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, LynkService_Exec_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *lynkServiceClient) DataProject(ctx context.Context, in *DataProjectRequest, opts ...grpc.CallOption) (*DataProjectResponse, error) {
	out := new(DataProjectResponse)
	err := c.cc.Invoke(ctx, LynkService_DataProject_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *lynkServiceClient) DataQuery(ctx context.Context, in *DataQuery, opts ...grpc.CallOption) (*DataResult, error) {
	out := new(DataResult)
	err := c.cc.Invoke(ctx, LynkService_DataQuery_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *lynkServiceClient) DataUpsert(ctx context.Context, in *DataInsert, opts ...grpc.CallOption) (*DataResult, error) {
	out := new(DataResult)
	err := c.cc.Invoke(ctx, LynkService_DataUpsert_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *lynkServiceClient) DataIgsert(ctx context.Context, in *DataInsert, opts ...grpc.CallOption) (*DataResult, error) {
	out := new(DataResult)
	err := c.cc.Invoke(ctx, LynkService_DataIgsert_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LynkServiceServer is the server API for LynkService service.
// All implementations must embed UnimplementedLynkServiceServer
// for forward compatibility
type LynkServiceServer interface {
	ApiList(context.Context, *ApiListRequest) (*ApiListResponse, error)
	Exec(context.Context, *Request) (*Response, error)
	DataProject(context.Context, *DataProjectRequest) (*DataProjectResponse, error)
	DataQuery(context.Context, *DataQuery) (*DataResult, error)
	DataUpsert(context.Context, *DataInsert) (*DataResult, error)
	DataIgsert(context.Context, *DataInsert) (*DataResult, error)
	mustEmbedUnimplementedLynkServiceServer()
}

// UnimplementedLynkServiceServer must be embedded to have forward compatible implementations.
type UnimplementedLynkServiceServer struct {
}

func (UnimplementedLynkServiceServer) ApiList(context.Context, *ApiListRequest) (*ApiListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ApiList not implemented")
}
func (UnimplementedLynkServiceServer) Exec(context.Context, *Request) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Exec not implemented")
}
func (UnimplementedLynkServiceServer) DataProject(context.Context, *DataProjectRequest) (*DataProjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DataProject not implemented")
}
func (UnimplementedLynkServiceServer) DataQuery(context.Context, *DataQuery) (*DataResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DataQuery not implemented")
}
func (UnimplementedLynkServiceServer) DataUpsert(context.Context, *DataInsert) (*DataResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DataUpsert not implemented")
}
func (UnimplementedLynkServiceServer) DataIgsert(context.Context, *DataInsert) (*DataResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DataIgsert not implemented")
}
func (UnimplementedLynkServiceServer) mustEmbedUnimplementedLynkServiceServer() {}

// UnsafeLynkServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LynkServiceServer will
// result in compilation errors.
type UnsafeLynkServiceServer interface {
	mustEmbedUnimplementedLynkServiceServer()
}

func RegisterLynkServiceServer(s grpc.ServiceRegistrar, srv LynkServiceServer) {
	s.RegisterService(&LynkService_ServiceDesc, srv)
}

func _LynkService_ApiList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ApiListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LynkServiceServer).ApiList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LynkService_ApiList_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LynkServiceServer).ApiList(ctx, req.(*ApiListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LynkService_Exec_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LynkServiceServer).Exec(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LynkService_Exec_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LynkServiceServer).Exec(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _LynkService_DataProject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DataProjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LynkServiceServer).DataProject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LynkService_DataProject_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LynkServiceServer).DataProject(ctx, req.(*DataProjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LynkService_DataQuery_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DataQuery)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LynkServiceServer).DataQuery(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LynkService_DataQuery_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LynkServiceServer).DataQuery(ctx, req.(*DataQuery))
	}
	return interceptor(ctx, in, info, handler)
}

func _LynkService_DataUpsert_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DataInsert)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LynkServiceServer).DataUpsert(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LynkService_DataUpsert_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LynkServiceServer).DataUpsert(ctx, req.(*DataInsert))
	}
	return interceptor(ctx, in, info, handler)
}

func _LynkService_DataIgsert_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DataInsert)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LynkServiceServer).DataIgsert(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LynkService_DataIgsert_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LynkServiceServer).DataIgsert(ctx, req.(*DataInsert))
	}
	return interceptor(ctx, in, info, handler)
}

// LynkService_ServiceDesc is the grpc.ServiceDesc for LynkService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var LynkService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "lynkapi.LynkService",
	HandlerType: (*LynkServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ApiList",
			Handler:    _LynkService_ApiList_Handler,
		},
		{
			MethodName: "Exec",
			Handler:    _LynkService_Exec_Handler,
		},
		{
			MethodName: "DataProject",
			Handler:    _LynkService_DataProject_Handler,
		},
		{
			MethodName: "DataQuery",
			Handler:    _LynkService_DataQuery_Handler,
		},
		{
			MethodName: "DataUpsert",
			Handler:    _LynkService_DataUpsert_Handler,
		},
		{
			MethodName: "DataIgsert",
			Handler:    _LynkService_DataIgsert_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "lynkapi/service.proto",
}
