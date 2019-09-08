// Code generated by protoc-gen-go. DO NOT EDIT.
// source: orders.proto

package pb

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type ValidatableCodeRequest struct {
	Datapoolid           string   `protobuf:"bytes,1,opt,name=datapoolid,proto3" json:"datapoolid,omitempty"`
	Add                  int32    `protobuf:"varint,2,opt,name=add,proto3" json:"add,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ValidatableCodeRequest) Reset()         { *m = ValidatableCodeRequest{} }
func (m *ValidatableCodeRequest) String() string { return proto.CompactTextString(m) }
func (*ValidatableCodeRequest) ProtoMessage()    {}
func (*ValidatableCodeRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_e0f5d4cf0fc9e41b, []int{0}
}

func (m *ValidatableCodeRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ValidatableCodeRequest.Unmarshal(m, b)
}
func (m *ValidatableCodeRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ValidatableCodeRequest.Marshal(b, m, deterministic)
}
func (m *ValidatableCodeRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ValidatableCodeRequest.Merge(m, src)
}
func (m *ValidatableCodeRequest) XXX_Size() int {
	return xxx_messageInfo_ValidatableCodeRequest.Size(m)
}
func (m *ValidatableCodeRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ValidatableCodeRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ValidatableCodeRequest proto.InternalMessageInfo

func (m *ValidatableCodeRequest) GetDatapoolid() string {
	if m != nil {
		return m.Datapoolid
	}
	return ""
}

func (m *ValidatableCodeRequest) GetAdd() int32 {
	if m != nil {
		return m.Add
	}
	return 0
}

type ValidatableCode struct {
	Data                 int32    `protobuf:"varint,2,opt,name=data,proto3" json:"data,omitempty"`
	Add                  int32    `protobuf:"varint,3,opt,name=add,proto3" json:"add,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ValidatableCode) Reset()         { *m = ValidatableCode{} }
func (m *ValidatableCode) String() string { return proto.CompactTextString(m) }
func (*ValidatableCode) ProtoMessage()    {}
func (*ValidatableCode) Descriptor() ([]byte, []int) {
	return fileDescriptor_e0f5d4cf0fc9e41b, []int{1}
}

func (m *ValidatableCode) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ValidatableCode.Unmarshal(m, b)
}
func (m *ValidatableCode) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ValidatableCode.Marshal(b, m, deterministic)
}
func (m *ValidatableCode) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ValidatableCode.Merge(m, src)
}
func (m *ValidatableCode) XXX_Size() int {
	return xxx_messageInfo_ValidatableCode.Size(m)
}
func (m *ValidatableCode) XXX_DiscardUnknown() {
	xxx_messageInfo_ValidatableCode.DiscardUnknown(m)
}

var xxx_messageInfo_ValidatableCode proto.InternalMessageInfo

func (m *ValidatableCode) GetData() int32 {
	if m != nil {
		return m.Data
	}
	return 0
}

func (m *ValidatableCode) GetAdd() int32 {
	if m != nil {
		return m.Add
	}
	return 0
}

type ValidationResult struct {
	AddedData            string   `protobuf:"bytes,1,opt,name=added_data,json=addedData,proto3" json:"added_data,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ValidationResult) Reset()         { *m = ValidationResult{} }
func (m *ValidationResult) String() string { return proto.CompactTextString(m) }
func (*ValidationResult) ProtoMessage()    {}
func (*ValidationResult) Descriptor() ([]byte, []int) {
	return fileDescriptor_e0f5d4cf0fc9e41b, []int{2}
}

func (m *ValidationResult) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ValidationResult.Unmarshal(m, b)
}
func (m *ValidationResult) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ValidationResult.Marshal(b, m, deterministic)
}
func (m *ValidationResult) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ValidationResult.Merge(m, src)
}
func (m *ValidationResult) XXX_Size() int {
	return xxx_messageInfo_ValidationResult.Size(m)
}
func (m *ValidationResult) XXX_DiscardUnknown() {
	xxx_messageInfo_ValidationResult.DiscardUnknown(m)
}

var xxx_messageInfo_ValidationResult proto.InternalMessageInfo

func (m *ValidationResult) GetAddedData() string {
	if m != nil {
		return m.AddedData
	}
	return ""
}

type CommitResult struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CommitResult) Reset()         { *m = CommitResult{} }
func (m *CommitResult) String() string { return proto.CompactTextString(m) }
func (*CommitResult) ProtoMessage()    {}
func (*CommitResult) Descriptor() ([]byte, []int) {
	return fileDescriptor_e0f5d4cf0fc9e41b, []int{3}
}

func (m *CommitResult) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CommitResult.Unmarshal(m, b)
}
func (m *CommitResult) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CommitResult.Marshal(b, m, deterministic)
}
func (m *CommitResult) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CommitResult.Merge(m, src)
}
func (m *CommitResult) XXX_Size() int {
	return xxx_messageInfo_CommitResult.Size(m)
}
func (m *CommitResult) XXX_DiscardUnknown() {
	xxx_messageInfo_CommitResult.DiscardUnknown(m)
}

var xxx_messageInfo_CommitResult proto.InternalMessageInfo

func init() {
	proto.RegisterType((*ValidatableCodeRequest)(nil), "pb.ValidatableCodeRequest")
	proto.RegisterType((*ValidatableCode)(nil), "pb.ValidatableCode")
	proto.RegisterType((*ValidationResult)(nil), "pb.ValidationResult")
	proto.RegisterType((*CommitResult)(nil), "pb.CommitResult")
}

func init() { proto.RegisterFile("orders.proto", fileDescriptor_e0f5d4cf0fc9e41b) }

var fileDescriptor_e0f5d4cf0fc9e41b = []byte{
	// 231 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x90, 0x41, 0x4b, 0x03, 0x31,
	0x10, 0x85, 0x9b, 0x56, 0x0b, 0x1d, 0x8a, 0x86, 0x51, 0x4a, 0x59, 0x50, 0x4a, 0x4e, 0x3d, 0x2d,
	0xa8, 0x07, 0x2f, 0xde, 0xea, 0x49, 0x0f, 0x42, 0x0e, 0x5e, 0x25, 0x61, 0x72, 0x08, 0xa4, 0x4e,
	0xdc, 0x4d, 0x7f, 0x89, 0x7f, 0x58, 0x92, 0x5d, 0xdb, 0xa5, 0xec, 0x6d, 0xf2, 0xf2, 0x5e, 0xe6,
	0x7b, 0x81, 0x25, 0x37, 0xe4, 0x9a, 0xb6, 0x8e, 0x0d, 0x27, 0xc6, 0x69, 0xb4, 0xea, 0x0d, 0x56,
	0x9f, 0x26, 0x78, 0x32, 0xc9, 0xd8, 0xe0, 0x76, 0x4c, 0x4e, 0xbb, 0x9f, 0x83, 0x6b, 0x13, 0xde,
	0x03, 0x64, 0x35, 0x32, 0x07, 0x4f, 0x6b, 0xb1, 0x11, 0xdb, 0x85, 0x1e, 0x28, 0x28, 0x61, 0x66,
	0x88, 0xd6, 0xd3, 0x8d, 0xd8, 0x5e, 0xea, 0x3c, 0xaa, 0x67, 0xb8, 0x3e, 0x7b, 0x0b, 0x11, 0x2e,
	0xf2, 0xb1, 0x77, 0x95, 0xf9, 0x3f, 0x38, 0x3b, 0x05, 0x1f, 0x40, 0xf6, 0x41, 0xcf, 0xdf, 0xda,
	0xb5, 0x87, 0x90, 0xf0, 0x0e, 0xc0, 0x10, 0x39, 0xfa, 0x2a, 0xf9, 0x6e, 0xfd, 0xa2, 0x28, 0xaf,
	0x26, 0x19, 0x75, 0x05, 0xcb, 0x1d, 0xef, 0xf7, 0x3e, 0x75, 0xf6, 0xc7, 0x5f, 0x01, 0xf3, 0x8f,
	0x52, 0x0e, 0xdf, 0x61, 0xd5, 0x77, 0x38, 0xa7, 0xa9, 0xea, 0x68, 0xeb, 0xf1, 0xba, 0xd5, 0xcd,
	0xc8, 0x9d, 0x9a, 0xe0, 0x0b, 0xc8, 0x6e, 0xcf, 0x09, 0x10, 0x6f, 0x07, 0xd6, 0x23, 0x70, 0x25,
	0xb3, 0x3a, 0x64, 0x52, 0x13, 0x3b, 0x2f, 0x1f, 0xfd, 0xf4, 0x17, 0x00, 0x00, 0xff, 0xff, 0x0d,
	0x8b, 0x01, 0xc6, 0x78, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// OrdersClient is the client API for Orders service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type OrdersClient interface {
	RequestValidatableCode(ctx context.Context, in *ValidatableCodeRequest, opts ...grpc.CallOption) (*ValidatableCode, error)
	CommitValidation(ctx context.Context, in *ValidationResult, opts ...grpc.CallOption) (*CommitResult, error)
}

type ordersClient struct {
	cc *grpc.ClientConn
}

func NewOrdersClient(cc *grpc.ClientConn) OrdersClient {
	return &ordersClient{cc}
}

func (c *ordersClient) RequestValidatableCode(ctx context.Context, in *ValidatableCodeRequest, opts ...grpc.CallOption) (*ValidatableCode, error) {
	out := new(ValidatableCode)
	err := c.cc.Invoke(ctx, "/pb.Orders/RequestValidatableCode", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ordersClient) CommitValidation(ctx context.Context, in *ValidationResult, opts ...grpc.CallOption) (*CommitResult, error) {
	out := new(CommitResult)
	err := c.cc.Invoke(ctx, "/pb.Orders/CommitValidation", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// OrdersServer is the server API for Orders service.
type OrdersServer interface {
	RequestValidatableCode(context.Context, *ValidatableCodeRequest) (*ValidatableCode, error)
	CommitValidation(context.Context, *ValidationResult) (*CommitResult, error)
}

func RegisterOrdersServer(s *grpc.Server, srv OrdersServer) {
	s.RegisterService(&_Orders_serviceDesc, srv)
}

func _Orders_RequestValidatableCode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidatableCodeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrdersServer).RequestValidatableCode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.Orders/RequestValidatableCode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrdersServer).RequestValidatableCode(ctx, req.(*ValidatableCodeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Orders_CommitValidation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidationResult)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrdersServer).CommitValidation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.Orders/CommitValidation",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrdersServer).CommitValidation(ctx, req.(*ValidationResult))
	}
	return interceptor(ctx, in, info, handler)
}

var _Orders_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pb.Orders",
	HandlerType: (*OrdersServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RequestValidatableCode",
			Handler:    _Orders_RequestValidatableCode_Handler,
		},
		{
			MethodName: "CommitValidation",
			Handler:    _Orders_CommitValidation_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "orders.proto",
}
