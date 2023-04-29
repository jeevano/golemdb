// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.21.12
// source: internal/rpc/kv.proto

package rpc

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
	Kv_Put_FullMethodName = "/rpc.Kv/Put"
	Kv_Get_FullMethodName = "/rpc.Kv/Get"
)

// KvClient is the client API for Kv service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type KvClient interface {
	Put(ctx context.Context, in *PutRequest, opts ...grpc.CallOption) (*Empty, error)
	Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error)
}

type kvClient struct {
	cc grpc.ClientConnInterface
}

func NewKvClient(cc grpc.ClientConnInterface) KvClient {
	return &kvClient{cc}
}

func (c *kvClient) Put(ctx context.Context, in *PutRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, Kv_Put_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kvClient) Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error) {
	out := new(GetResponse)
	err := c.cc.Invoke(ctx, Kv_Get_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// KvServer is the server API for Kv service.
// All implementations must embed UnimplementedKvServer
// for forward compatibility
type KvServer interface {
	Put(context.Context, *PutRequest) (*Empty, error)
	Get(context.Context, *GetRequest) (*GetResponse, error)
	mustEmbedUnimplementedKvServer()
}

// UnimplementedKvServer must be embedded to have forward compatible implementations.
type UnimplementedKvServer struct {
}

func (UnimplementedKvServer) Put(context.Context, *PutRequest) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Put not implemented")
}
func (UnimplementedKvServer) Get(context.Context, *GetRequest) (*GetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedKvServer) mustEmbedUnimplementedKvServer() {}

// UnsafeKvServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to KvServer will
// result in compilation errors.
type UnsafeKvServer interface {
	mustEmbedUnimplementedKvServer()
}

func RegisterKvServer(s grpc.ServiceRegistrar, srv KvServer) {
	s.RegisterService(&Kv_ServiceDesc, srv)
}

func _Kv_Put_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PutRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KvServer).Put(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Kv_Put_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KvServer).Put(ctx, req.(*PutRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Kv_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KvServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Kv_Get_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KvServer).Get(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Kv_ServiceDesc is the grpc.ServiceDesc for Kv service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Kv_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "rpc.Kv",
	HandlerType: (*KvServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Put",
			Handler:    _Kv_Put_Handler,
		},
		{
			MethodName: "Get",
			Handler:    _Kv_Get_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "internal/rpc/kv.proto",
}
