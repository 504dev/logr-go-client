// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.25.3
// source: logr.proto

package logrpcv1

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

// LogrpcClient is the client API for Logrpc service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LogrpcClient interface {
	Push(ctx context.Context, in *LogrpcPackage, opts ...grpc.CallOption) (*Response, error)
}

type logrpcClient struct {
	cc grpc.ClientConnInterface
}

func NewLogrpcClient(cc grpc.ClientConnInterface) LogrpcClient {
	return &logrpcClient{cc}
}

func (c *logrpcClient) Push(ctx context.Context, in *LogrpcPackage, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/logr.Logrpc/Push", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LogrpcServer is the server API for Logrpc service.
// All implementations must embed UnimplementedLogrpcServer
// for forward compatibility
type LogrpcServer interface {
	Push(context.Context, *LogrpcPackage) (*Response, error)
	mustEmbedUnimplementedLogrpcServer()
}

// UnimplementedLogrpcServer must be embedded to have forward compatible implementations.
type UnimplementedLogrpcServer struct {
}

func (UnimplementedLogrpcServer) Push(context.Context, *LogrpcPackage) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Push not implemented")
}
func (UnimplementedLogrpcServer) mustEmbedUnimplementedLogrpcServer() {}

// UnsafeLogrpcServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LogrpcServer will
// result in compilation errors.
type UnsafeLogrpcServer interface {
	mustEmbedUnimplementedLogrpcServer()
}

func RegisterLogrpcServer(s grpc.ServiceRegistrar, srv LogrpcServer) {
	s.RegisterService(&Logrpc_ServiceDesc, srv)
}

func _Logrpc_Push_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LogrpcPackage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LogrpcServer).Push(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/logr.Logrpc/Push",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LogrpcServer).Push(ctx, req.(*LogrpcPackage))
	}
	return interceptor(ctx, in, info, handler)
}

// Logrpc_ServiceDesc is the grpc.ServiceDesc for Logrpc service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Logrpc_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "logr.Logrpc",
	HandlerType: (*LogrpcServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Push",
			Handler:    _Logrpc_Push_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "logr.proto",
}
