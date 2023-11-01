// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.12.4
// source: dataserver.proto

package dataserver

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

// DataServerClient is the client API for DataServer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DataServerClient interface {
	GetFileBlock(ctx context.Context, in *GetFileBlockRequest, opts ...grpc.CallOption) (DataServer_GetFileBlockClient, error)
	PutFileBlock(ctx context.Context, opts ...grpc.CallOption) (DataServer_PutFileBlockClient, error)
}

type dataServerClient struct {
	cc grpc.ClientConnInterface
}

func NewDataServerClient(cc grpc.ClientConnInterface) DataServerClient {
	return &dataServerClient{cc}
}

func (c *dataServerClient) GetFileBlock(ctx context.Context, in *GetFileBlockRequest, opts ...grpc.CallOption) (DataServer_GetFileBlockClient, error) {
	stream, err := c.cc.NewStream(ctx, &DataServer_ServiceDesc.Streams[0], "/dataserver.DataServer/GetFileBlock", opts...)
	if err != nil {
		return nil, err
	}
	x := &dataServerGetFileBlockClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type DataServer_GetFileBlockClient interface {
	Recv() (*GetFileBlockReply, error)
	grpc.ClientStream
}

type dataServerGetFileBlockClient struct {
	grpc.ClientStream
}

func (x *dataServerGetFileBlockClient) Recv() (*GetFileBlockReply, error) {
	m := new(GetFileBlockReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *dataServerClient) PutFileBlock(ctx context.Context, opts ...grpc.CallOption) (DataServer_PutFileBlockClient, error) {
	stream, err := c.cc.NewStream(ctx, &DataServer_ServiceDesc.Streams[1], "/dataserver.DataServer/PutFileBlock", opts...)
	if err != nil {
		return nil, err
	}
	x := &dataServerPutFileBlockClient{stream}
	return x, nil
}

type DataServer_PutFileBlockClient interface {
	Send(*PutFileBlockRequest) error
	CloseAndRecv() (*PutFileBlockReply, error)
	grpc.ClientStream
}

type dataServerPutFileBlockClient struct {
	grpc.ClientStream
}

func (x *dataServerPutFileBlockClient) Send(m *PutFileBlockRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *dataServerPutFileBlockClient) CloseAndRecv() (*PutFileBlockReply, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(PutFileBlockReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// DataServerServer is the server API for DataServer service.
// All implementations must embed UnimplementedDataServerServer
// for forward compatibility
type DataServerServer interface {
	GetFileBlock(*GetFileBlockRequest, DataServer_GetFileBlockServer) error
	PutFileBlock(DataServer_PutFileBlockServer) error
	mustEmbedUnimplementedDataServerServer()
}

// UnimplementedDataServerServer must be embedded to have forward compatible implementations.
type UnimplementedDataServerServer struct {
}

func (UnimplementedDataServerServer) GetFileBlock(*GetFileBlockRequest, DataServer_GetFileBlockServer) error {
	return status.Errorf(codes.Unimplemented, "method GetFileBlock not implemented")
}
func (UnimplementedDataServerServer) PutFileBlock(DataServer_PutFileBlockServer) error {
	return status.Errorf(codes.Unimplemented, "method PutFileBlock not implemented")
}
func (UnimplementedDataServerServer) mustEmbedUnimplementedDataServerServer() {}

// UnsafeDataServerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DataServerServer will
// result in compilation errors.
type UnsafeDataServerServer interface {
	mustEmbedUnimplementedDataServerServer()
}

func RegisterDataServerServer(s grpc.ServiceRegistrar, srv DataServerServer) {
	s.RegisterService(&DataServer_ServiceDesc, srv)
}

func _DataServer_GetFileBlock_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetFileBlockRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(DataServerServer).GetFileBlock(m, &dataServerGetFileBlockServer{stream})
}

type DataServer_GetFileBlockServer interface {
	Send(*GetFileBlockReply) error
	grpc.ServerStream
}

type dataServerGetFileBlockServer struct {
	grpc.ServerStream
}

func (x *dataServerGetFileBlockServer) Send(m *GetFileBlockReply) error {
	return x.ServerStream.SendMsg(m)
}

func _DataServer_PutFileBlock_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(DataServerServer).PutFileBlock(&dataServerPutFileBlockServer{stream})
}

type DataServer_PutFileBlockServer interface {
	SendAndClose(*PutFileBlockReply) error
	Recv() (*PutFileBlockRequest, error)
	grpc.ServerStream
}

type dataServerPutFileBlockServer struct {
	grpc.ServerStream
}

func (x *dataServerPutFileBlockServer) SendAndClose(m *PutFileBlockReply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *dataServerPutFileBlockServer) Recv() (*PutFileBlockRequest, error) {
	m := new(PutFileBlockRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// DataServer_ServiceDesc is the grpc.ServiceDesc for DataServer service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DataServer_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "dataserver.DataServer",
	HandlerType: (*DataServerServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetFileBlock",
			Handler:       _DataServer_GetFileBlock_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "PutFileBlock",
			Handler:       _DataServer_PutFileBlock_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "dataserver.proto",
}