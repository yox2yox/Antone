// Code generated by protoc-gen-go. DO NOT EDIT.
// source: accounting.proto

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

type SignupWorkerRequest struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Addr                 string   `protobuf:"bytes,2,opt,name=addr,proto3" json:"addr,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SignupWorkerRequest) Reset()         { *m = SignupWorkerRequest{} }
func (m *SignupWorkerRequest) String() string { return proto.CompactTextString(m) }
func (*SignupWorkerRequest) ProtoMessage()    {}
func (*SignupWorkerRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_cb3d75761beb5907, []int{0}
}

func (m *SignupWorkerRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SignupWorkerRequest.Unmarshal(m, b)
}
func (m *SignupWorkerRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SignupWorkerRequest.Marshal(b, m, deterministic)
}
func (m *SignupWorkerRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SignupWorkerRequest.Merge(m, src)
}
func (m *SignupWorkerRequest) XXX_Size() int {
	return xxx_messageInfo_SignupWorkerRequest.Size(m)
}
func (m *SignupWorkerRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SignupWorkerRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SignupWorkerRequest proto.InternalMessageInfo

func (m *SignupWorkerRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *SignupWorkerRequest) GetAddr() string {
	if m != nil {
		return m.Addr
	}
	return ""
}

type GetWorkerRequest struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetWorkerRequest) Reset()         { *m = GetWorkerRequest{} }
func (m *GetWorkerRequest) String() string { return proto.CompactTextString(m) }
func (*GetWorkerRequest) ProtoMessage()    {}
func (*GetWorkerRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_cb3d75761beb5907, []int{1}
}

func (m *GetWorkerRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetWorkerRequest.Unmarshal(m, b)
}
func (m *GetWorkerRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetWorkerRequest.Marshal(b, m, deterministic)
}
func (m *GetWorkerRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetWorkerRequest.Merge(m, src)
}
func (m *GetWorkerRequest) XXX_Size() int {
	return xxx_messageInfo_GetWorkerRequest.Size(m)
}
func (m *GetWorkerRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetWorkerRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetWorkerRequest proto.InternalMessageInfo

func (m *GetWorkerRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type WorkerAccount struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Reputation           int32    `protobuf:"varint,2,opt,name=reputation,proto3" json:"reputation,omitempty"`
	Balance              int32    `protobuf:"varint,3,opt,name=balance,proto3" json:"balance,omitempty"`
	Holdings             []string `protobuf:"bytes,4,rep,name=holdings,proto3" json:"holdings,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *WorkerAccount) Reset()         { *m = WorkerAccount{} }
func (m *WorkerAccount) String() string { return proto.CompactTextString(m) }
func (*WorkerAccount) ProtoMessage()    {}
func (*WorkerAccount) Descriptor() ([]byte, []int) {
	return fileDescriptor_cb3d75761beb5907, []int{2}
}

func (m *WorkerAccount) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_WorkerAccount.Unmarshal(m, b)
}
func (m *WorkerAccount) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_WorkerAccount.Marshal(b, m, deterministic)
}
func (m *WorkerAccount) XXX_Merge(src proto.Message) {
	xxx_messageInfo_WorkerAccount.Merge(m, src)
}
func (m *WorkerAccount) XXX_Size() int {
	return xxx_messageInfo_WorkerAccount.Size(m)
}
func (m *WorkerAccount) XXX_DiscardUnknown() {
	xxx_messageInfo_WorkerAccount.DiscardUnknown(m)
}

var xxx_messageInfo_WorkerAccount proto.InternalMessageInfo

func (m *WorkerAccount) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *WorkerAccount) GetReputation() int32 {
	if m != nil {
		return m.Reputation
	}
	return 0
}

func (m *WorkerAccount) GetBalance() int32 {
	if m != nil {
		return m.Balance
	}
	return 0
}

func (m *WorkerAccount) GetHoldings() []string {
	if m != nil {
		return m.Holdings
	}
	return nil
}

type SignupClientRequest struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Addr                 string   `protobuf:"bytes,2,opt,name=addr,proto3" json:"addr,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SignupClientRequest) Reset()         { *m = SignupClientRequest{} }
func (m *SignupClientRequest) String() string { return proto.CompactTextString(m) }
func (*SignupClientRequest) ProtoMessage()    {}
func (*SignupClientRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_cb3d75761beb5907, []int{3}
}

func (m *SignupClientRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SignupClientRequest.Unmarshal(m, b)
}
func (m *SignupClientRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SignupClientRequest.Marshal(b, m, deterministic)
}
func (m *SignupClientRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SignupClientRequest.Merge(m, src)
}
func (m *SignupClientRequest) XXX_Size() int {
	return xxx_messageInfo_SignupClientRequest.Size(m)
}
func (m *SignupClientRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SignupClientRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SignupClientRequest proto.InternalMessageInfo

func (m *SignupClientRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *SignupClientRequest) GetAddr() string {
	if m != nil {
		return m.Addr
	}
	return ""
}

type GetClientRequest struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetClientRequest) Reset()         { *m = GetClientRequest{} }
func (m *GetClientRequest) String() string { return proto.CompactTextString(m) }
func (*GetClientRequest) ProtoMessage()    {}
func (*GetClientRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_cb3d75761beb5907, []int{4}
}

func (m *GetClientRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetClientRequest.Unmarshal(m, b)
}
func (m *GetClientRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetClientRequest.Marshal(b, m, deterministic)
}
func (m *GetClientRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetClientRequest.Merge(m, src)
}
func (m *GetClientRequest) XXX_Size() int {
	return xxx_messageInfo_GetClientRequest.Size(m)
}
func (m *GetClientRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetClientRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetClientRequest proto.InternalMessageInfo

func (m *GetClientRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type ClientAccount struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Balance              int32    `protobuf:"varint,3,opt,name=balance,proto3" json:"balance,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ClientAccount) Reset()         { *m = ClientAccount{} }
func (m *ClientAccount) String() string { return proto.CompactTextString(m) }
func (*ClientAccount) ProtoMessage()    {}
func (*ClientAccount) Descriptor() ([]byte, []int) {
	return fileDescriptor_cb3d75761beb5907, []int{5}
}

func (m *ClientAccount) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ClientAccount.Unmarshal(m, b)
}
func (m *ClientAccount) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ClientAccount.Marshal(b, m, deterministic)
}
func (m *ClientAccount) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ClientAccount.Merge(m, src)
}
func (m *ClientAccount) XXX_Size() int {
	return xxx_messageInfo_ClientAccount.Size(m)
}
func (m *ClientAccount) XXX_DiscardUnknown() {
	xxx_messageInfo_ClientAccount.DiscardUnknown(m)
}

var xxx_messageInfo_ClientAccount proto.InternalMessageInfo

func (m *ClientAccount) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *ClientAccount) GetBalance() int32 {
	if m != nil {
		return m.Balance
	}
	return 0
}

func init() {
	proto.RegisterType((*SignupWorkerRequest)(nil), "pb.SignupWorkerRequest")
	proto.RegisterType((*GetWorkerRequest)(nil), "pb.GetWorkerRequest")
	proto.RegisterType((*WorkerAccount)(nil), "pb.WorkerAccount")
	proto.RegisterType((*SignupClientRequest)(nil), "pb.SignupClientRequest")
	proto.RegisterType((*GetClientRequest)(nil), "pb.GetClientRequest")
	proto.RegisterType((*ClientAccount)(nil), "pb.ClientAccount")
}

func init() { proto.RegisterFile("accounting.proto", fileDescriptor_cb3d75761beb5907) }

var fileDescriptor_cb3d75761beb5907 = []byte{
	// 276 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x92, 0xb1, 0x4f, 0xeb, 0x30,
	0x10, 0xc6, 0x5b, 0xb7, 0xef, 0x41, 0x4f, 0x04, 0x95, 0x03, 0x09, 0xab, 0x03, 0xaa, 0x3c, 0x75,
	0xca, 0x00, 0x53, 0x11, 0x4b, 0xc5, 0x80, 0x58, 0xc3, 0xc0, 0xec, 0xc4, 0x26, 0x58, 0x44, 0xb6,
	0x71, 0xed, 0x9d, 0x3f, 0x1d, 0xd5, 0x16, 0x69, 0x02, 0x6d, 0x11, 0x5b, 0x7c, 0x77, 0xbf, 0xbb,
	0xfb, 0xee, 0x0b, 0x4c, 0x79, 0x55, 0x99, 0xa0, 0xbd, 0xd2, 0x75, 0x6e, 0x9d, 0xf1, 0x06, 0x89,
	0x2d, 0xd9, 0x12, 0xce, 0x9f, 0x54, 0xad, 0x83, 0x7d, 0x36, 0xee, 0x4d, 0xba, 0x42, 0xbe, 0x07,
	0xb9, 0xf6, 0x78, 0x0a, 0x44, 0x09, 0x3a, 0x9c, 0x0f, 0x17, 0x93, 0x82, 0x28, 0x81, 0x08, 0x63,
	0x2e, 0x84, 0xa3, 0x24, 0x46, 0xe2, 0x37, 0x63, 0x30, 0x7d, 0x90, 0xfe, 0x20, 0xc7, 0x02, 0x64,
	0xa9, 0x60, 0x95, 0x86, 0xff, 0x68, 0x7c, 0x05, 0xe0, 0xa4, 0x0d, 0x9e, 0x7b, 0x65, 0x74, 0x6c,
	0xff, 0xaf, 0xe8, 0x44, 0x90, 0xc2, 0x51, 0xc9, 0x1b, 0xae, 0x2b, 0x49, 0x47, 0x31, 0xf9, 0xf5,
	0xc4, 0x19, 0x1c, 0xbf, 0x9a, 0x46, 0x28, 0x5d, 0xaf, 0xe9, 0x78, 0x3e, 0x5a, 0x4c, 0x8a, 0xf6,
	0xbd, 0x55, 0x75, 0xdf, 0x28, 0xa9, 0xfd, 0xdf, 0x55, 0x1d, 0xe4, 0xd8, 0x12, 0xb2, 0x54, 0xb0,
	0x4f, 0xd5, 0xde, 0xad, 0xaf, 0x3f, 0x08, 0xc0, 0xaa, 0x35, 0x02, 0xef, 0xe0, 0xa4, 0x7b, 0x7e,
	0xbc, 0xcc, 0x6d, 0x99, 0xef, 0x30, 0x64, 0x76, 0xb6, 0x49, 0xf4, 0x4e, 0xc9, 0x06, 0x5b, 0x3a,
	0x6d, 0xd3, 0xa5, 0x7b, 0x02, 0x12, 0xdd, 0x5b, 0x99, 0x0d, 0xf0, 0x16, 0xb2, 0xd6, 0xbf, 0x47,
	0xfd, 0x62, 0xf0, 0x62, 0x53, 0xf5, 0xdd, 0xd2, 0xdd, 0x93, 0x13, 0x9b, 0x3a, 0xf6, 0xd8, 0xdf,
	0xe7, 0x96, 0xff, 0xe3, 0xdf, 0x77, 0xf3, 0x19, 0x00, 0x00, 0xff, 0xff, 0xb1, 0x4d, 0x1e, 0x33,
	0x91, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// AccountingClient is the client API for Accounting service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type AccountingClient interface {
	SignupWorker(ctx context.Context, in *SignupWorkerRequest, opts ...grpc.CallOption) (*WorkerAccount, error)
	SignupClient(ctx context.Context, in *SignupClientRequest, opts ...grpc.CallOption) (*ClientAccount, error)
	GetWorkerInfo(ctx context.Context, in *GetWorkerRequest, opts ...grpc.CallOption) (*WorkerAccount, error)
	GetClientInfo(ctx context.Context, in *GetClientRequest, opts ...grpc.CallOption) (*ClientAccount, error)
}

type accountingClient struct {
	cc *grpc.ClientConn
}

func NewAccountingClient(cc *grpc.ClientConn) AccountingClient {
	return &accountingClient{cc}
}

func (c *accountingClient) SignupWorker(ctx context.Context, in *SignupWorkerRequest, opts ...grpc.CallOption) (*WorkerAccount, error) {
	out := new(WorkerAccount)
	err := c.cc.Invoke(ctx, "/pb.Accounting/SignupWorker", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountingClient) SignupClient(ctx context.Context, in *SignupClientRequest, opts ...grpc.CallOption) (*ClientAccount, error) {
	out := new(ClientAccount)
	err := c.cc.Invoke(ctx, "/pb.Accounting/SignupClient", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountingClient) GetWorkerInfo(ctx context.Context, in *GetWorkerRequest, opts ...grpc.CallOption) (*WorkerAccount, error) {
	out := new(WorkerAccount)
	err := c.cc.Invoke(ctx, "/pb.Accounting/GetWorkerInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *accountingClient) GetClientInfo(ctx context.Context, in *GetClientRequest, opts ...grpc.CallOption) (*ClientAccount, error) {
	out := new(ClientAccount)
	err := c.cc.Invoke(ctx, "/pb.Accounting/GetClientInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AccountingServer is the server API for Accounting service.
type AccountingServer interface {
	SignupWorker(context.Context, *SignupWorkerRequest) (*WorkerAccount, error)
	SignupClient(context.Context, *SignupClientRequest) (*ClientAccount, error)
	GetWorkerInfo(context.Context, *GetWorkerRequest) (*WorkerAccount, error)
	GetClientInfo(context.Context, *GetClientRequest) (*ClientAccount, error)
}

func RegisterAccountingServer(s *grpc.Server, srv AccountingServer) {
	s.RegisterService(&_Accounting_serviceDesc, srv)
}

func _Accounting_SignupWorker_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignupWorkerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountingServer).SignupWorker(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.Accounting/SignupWorker",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountingServer).SignupWorker(ctx, req.(*SignupWorkerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Accounting_SignupClient_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignupClientRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountingServer).SignupClient(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.Accounting/SignupClient",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountingServer).SignupClient(ctx, req.(*SignupClientRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Accounting_GetWorkerInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetWorkerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountingServer).GetWorkerInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.Accounting/GetWorkerInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountingServer).GetWorkerInfo(ctx, req.(*GetWorkerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Accounting_GetClientInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetClientRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AccountingServer).GetClientInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.Accounting/GetClientInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AccountingServer).GetClientInfo(ctx, req.(*GetClientRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Accounting_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pb.Accounting",
	HandlerType: (*AccountingServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SignupWorker",
			Handler:    _Accounting_SignupWorker_Handler,
		},
		{
			MethodName: "SignupClient",
			Handler:    _Accounting_SignupClient_Handler,
		},
		{
			MethodName: "GetWorkerInfo",
			Handler:    _Accounting_GetWorkerInfo_Handler,
		},
		{
			MethodName: "GetClientInfo",
			Handler:    _Accounting_GetClientInfo_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "accounting.proto",
}
