// Code generated by protoc-gen-go. DO NOT EDIT.
// source: worker.proto

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
	Bridgeid             string   `protobuf:"bytes,1,opt,name=bridgeid,proto3" json:"bridgeid,omitempty"`
	Datapoolid           string   `protobuf:"bytes,2,opt,name=datapoolid,proto3" json:"datapoolid,omitempty"`
	Add                  int32    `protobuf:"varint,3,opt,name=add,proto3" json:"add,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ValidatableCodeRequest) Reset()         { *m = ValidatableCodeRequest{} }
func (m *ValidatableCodeRequest) String() string { return proto.CompactTextString(m) }
func (*ValidatableCodeRequest) ProtoMessage()    {}
func (*ValidatableCodeRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_e4ff6184b07e587a, []int{0}
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

func (m *ValidatableCodeRequest) GetBridgeid() string {
	if m != nil {
		return m.Bridgeid
	}
	return ""
}

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
	Data                 int32    `protobuf:"varint,1,opt,name=data,proto3" json:"data,omitempty"`
	Add                  int32    `protobuf:"varint,2,opt,name=add,proto3" json:"add,omitempty"`
	Reputation           int32    `protobuf:"varint,3,opt,name=reputation,proto3" json:"reputation,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ValidatableCode) Reset()         { *m = ValidatableCode{} }
func (m *ValidatableCode) String() string { return proto.CompactTextString(m) }
func (*ValidatableCode) ProtoMessage()    {}
func (*ValidatableCode) Descriptor() ([]byte, []int) {
	return fileDescriptor_e4ff6184b07e587a, []int{1}
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

func (m *ValidatableCode) GetReputation() int32 {
	if m != nil {
		return m.Reputation
	}
	return 0
}

type ValidationResult struct {
	Pool                 int32    `protobuf:"varint,1,opt,name=pool,proto3" json:"pool,omitempty"`
	Reject               bool     `protobuf:"varint,2,opt,name=reject,proto3" json:"reject,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ValidationResult) Reset()         { *m = ValidationResult{} }
func (m *ValidationResult) String() string { return proto.CompactTextString(m) }
func (*ValidationResult) ProtoMessage()    {}
func (*ValidationResult) Descriptor() ([]byte, []int) {
	return fileDescriptor_e4ff6184b07e587a, []int{2}
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

func (m *ValidationResult) GetPool() int32 {
	if m != nil {
		return m.Pool
	}
	return 0
}

func (m *ValidationResult) GetReject() bool {
	if m != nil {
		return m.Reject
	}
	return false
}

func init() {
	proto.RegisterType((*ValidatableCodeRequest)(nil), "pb.ValidatableCodeRequest")
	proto.RegisterType((*ValidatableCode)(nil), "pb.ValidatableCode")
	proto.RegisterType((*ValidationResult)(nil), "pb.ValidationResult")
}

func init() { proto.RegisterFile("worker.proto", fileDescriptor_e4ff6184b07e587a) }

var fileDescriptor_e4ff6184b07e587a = []byte{
	// 252 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x91, 0x31, 0x4f, 0xc3, 0x30,
	0x10, 0x85, 0x9b, 0x94, 0x46, 0xe5, 0x84, 0xd4, 0xea, 0x40, 0x55, 0x94, 0x01, 0x55, 0x9e, 0x3a,
	0x65, 0x80, 0xbd, 0x0b, 0x43, 0x47, 0x24, 0x0f, 0x74, 0xb6, 0xf1, 0x81, 0x0c, 0x51, 0x6d, 0xdc,
	0x8b, 0xf8, 0x1d, 0xfc, 0x63, 0x64, 0x2b, 0x4d, 0xa3, 0x28, 0xdb, 0xbd, 0x17, 0xbd, 0xef, 0x5d,
	0xce, 0x70, 0xf7, 0xeb, 0xc2, 0x37, 0x85, 0xda, 0x07, 0xc7, 0x0e, 0x73, 0xaf, 0xc5, 0x07, 0x6c,
	0xde, 0x54, 0x63, 0x8d, 0x62, 0xa5, 0x1b, 0x7a, 0x71, 0x86, 0x24, 0xfd, 0xb4, 0x74, 0x66, 0xac,
	0x60, 0xa9, 0x83, 0x35, 0x9f, 0x64, 0x4d, 0x99, 0x6d, 0xb3, 0xdd, 0xad, 0xec, 0x35, 0x3e, 0x02,
	0xc4, 0x84, 0x77, 0xae, 0xb1, 0xa6, 0xcc, 0xd3, 0xd7, 0x81, 0x83, 0x6b, 0x98, 0x2b, 0x63, 0xca,
	0xf9, 0x36, 0xdb, 0x2d, 0x64, 0x1c, 0xc5, 0x11, 0x56, 0xa3, 0x1e, 0x44, 0xb8, 0x89, 0x32, 0xc1,
	0x17, 0x32, 0xcd, 0x97, 0x60, 0xde, 0x07, 0x63, 0x55, 0x20, 0xdf, 0xb2, 0x62, 0xeb, 0x4e, 0x1d,
	0x71, 0xe0, 0x88, 0x3d, 0xac, 0x3b, 0xb0, 0x75, 0x27, 0x49, 0xe7, 0xb6, 0xe1, 0x48, 0x8e, 0x8b,
	0x5c, 0xc8, 0x71, 0xc6, 0x0d, 0x14, 0x81, 0xbe, 0xe8, 0x9d, 0x13, 0x7c, 0x29, 0x3b, 0xf5, 0xf4,
	0x97, 0x41, 0x71, 0x4c, 0x57, 0xc1, 0x3d, 0xac, 0x5e, 0x83, 0xa1, 0x70, 0xe5, 0xe1, 0x7d, 0xed,
	0x75, 0x3d, 0x5a, 0xbc, 0x7a, 0x18, 0x98, 0x7d, 0xa9, 0x98, 0xe1, 0x01, 0xf0, 0x40, 0x3c, 0xfe,
	0xcd, 0x6a, 0x02, 0xd1, 0xdd, 0xb8, 0x9a, 0xc2, 0x8b, 0x99, 0x2e, 0xd2, 0xfb, 0x3c, 0xff, 0x07,
	0x00, 0x00, 0xff, 0xff, 0x58, 0xf1, 0xe2, 0x38, 0xaf, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// WorkerClient is the client API for Worker service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type WorkerClient interface {
	OrderValidation(ctx context.Context, in *ValidatableCode, opts ...grpc.CallOption) (*ValidationResult, error)
	GetValidatableCode(ctx context.Context, in *ValidatableCodeRequest, opts ...grpc.CallOption) (*ValidatableCode, error)
}

type workerClient struct {
	cc *grpc.ClientConn
}

func NewWorkerClient(cc *grpc.ClientConn) WorkerClient {
	return &workerClient{cc}
}

func (c *workerClient) OrderValidation(ctx context.Context, in *ValidatableCode, opts ...grpc.CallOption) (*ValidationResult, error) {
	out := new(ValidationResult)
	err := c.cc.Invoke(ctx, "/pb.Worker/OrderValidation", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workerClient) GetValidatableCode(ctx context.Context, in *ValidatableCodeRequest, opts ...grpc.CallOption) (*ValidatableCode, error) {
	out := new(ValidatableCode)
	err := c.cc.Invoke(ctx, "/pb.Worker/GetValidatableCode", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WorkerServer is the server API for Worker service.
type WorkerServer interface {
	OrderValidation(context.Context, *ValidatableCode) (*ValidationResult, error)
	GetValidatableCode(context.Context, *ValidatableCodeRequest) (*ValidatableCode, error)
}

func RegisterWorkerServer(s *grpc.Server, srv WorkerServer) {
	s.RegisterService(&_Worker_serviceDesc, srv)
}

func _Worker_OrderValidation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidatableCode)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkerServer).OrderValidation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.Worker/OrderValidation",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkerServer).OrderValidation(ctx, req.(*ValidatableCode))
	}
	return interceptor(ctx, in, info, handler)
}

func _Worker_GetValidatableCode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidatableCodeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkerServer).GetValidatableCode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.Worker/GetValidatableCode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkerServer).GetValidatableCode(ctx, req.(*ValidatableCodeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Worker_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pb.Worker",
	HandlerType: (*WorkerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "OrderValidation",
			Handler:    _Worker_OrderValidation_Handler,
		},
		{
			MethodName: "GetValidatableCode",
			Handler:    _Worker_GetValidatableCode_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "worker.proto",
}
