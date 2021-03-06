// Code generated by protoc-gen-go. DO NOT EDIT.
// source: grpc.proto

/*
Package grpc is a generated protocol buffer package.

It is generated from these files:
	grpc.proto

It has these top-level messages:
	MultiAddress
	PingRequest
	PingResponse
	PongRequest
	PongResponse
	QueryRequest
	QueryResponse
	StreamMessage
	OpenOrderRequest
	OpenOrderResponse
	EncryptedOrderFragment
	EncryptedCoExpShare
	OrderFragmentCommitment
	CoExpCommitment
	StatusRequest
	StatusResponse
	UpdateMidpointRequest
	UpdateMidpointResponse
*/
package grpc

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc1 "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type OrderType int32

const (
	OrderType_Midpoint     OrderType = 0
	OrderType_Limit        OrderType = 1
	OrderType_Midpoint_FOK OrderType = 2
	OrderType_Limit_FOK    OrderType = 3
)

var OrderType_name = map[int32]string{
	0: "Midpoint",
	1: "Limit",
	2: "Midpoint_FOK",
	3: "Limit_FOK",
}
var OrderType_value = map[string]int32{
	"Midpoint":     0,
	"Limit":        1,
	"Midpoint_FOK": 2,
	"Limit_FOK":    3,
}

func (x OrderType) String() string {
	return proto.EnumName(OrderType_name, int32(x))
}
func (OrderType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type OrderParity int32

const (
	OrderParity_Buy  OrderParity = 0
	OrderParity_Sell OrderParity = 1
)

var OrderParity_name = map[int32]string{
	0: "Buy",
	1: "Sell",
}
var OrderParity_value = map[string]int32{
	"Buy":  0,
	"Sell": 1,
}

func (x OrderParity) String() string {
	return proto.EnumName(OrderParity_name, int32(x))
}
func (OrderParity) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type OrderSettlement int32

const (
	OrderSettlement_Nil         OrderSettlement = 0
	OrderSettlement_RenEx       OrderSettlement = 1
	OrderSettlement_RenExAtomic OrderSettlement = 2
)

var OrderSettlement_name = map[int32]string{
	0: "Nil",
	1: "RenEx",
	2: "RenExAtomic",
}
var OrderSettlement_value = map[string]int32{
	"Nil":         0,
	"RenEx":       1,
	"RenExAtomic": 2,
}

func (x OrderSettlement) String() string {
	return proto.EnumName(OrderSettlement_name, int32(x))
}
func (OrderSettlement) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type MultiAddress struct {
	Signature         []byte `protobuf:"bytes,1,opt,name=signature,proto3" json:"signature,omitempty"`
	MultiAddress      string `protobuf:"bytes,2,opt,name=multiAddress" json:"multiAddress,omitempty"`
	MultiAddressNonce uint64 `protobuf:"varint,3,opt,name=multiAddressNonce" json:"multiAddressNonce,omitempty"`
}

func (m *MultiAddress) Reset()                    { *m = MultiAddress{} }
func (m *MultiAddress) String() string            { return proto.CompactTextString(m) }
func (*MultiAddress) ProtoMessage()               {}
func (*MultiAddress) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *MultiAddress) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

func (m *MultiAddress) GetMultiAddress() string {
	if m != nil {
		return m.MultiAddress
	}
	return ""
}

func (m *MultiAddress) GetMultiAddressNonce() uint64 {
	if m != nil {
		return m.MultiAddressNonce
	}
	return 0
}

type PingRequest struct {
	MultiAddress *MultiAddress `protobuf:"bytes,1,opt,name=multiAddress" json:"multiAddress,omitempty"`
}

func (m *PingRequest) Reset()                    { *m = PingRequest{} }
func (m *PingRequest) String() string            { return proto.CompactTextString(m) }
func (*PingRequest) ProtoMessage()               {}
func (*PingRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *PingRequest) GetMultiAddress() *MultiAddress {
	if m != nil {
		return m.MultiAddress
	}
	return nil
}

type PingResponse struct {
}

func (m *PingResponse) Reset()                    { *m = PingResponse{} }
func (m *PingResponse) String() string            { return proto.CompactTextString(m) }
func (*PingResponse) ProtoMessage()               {}
func (*PingResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type PongRequest struct {
	MultiAddress *MultiAddress `protobuf:"bytes,1,opt,name=multiAddress" json:"multiAddress,omitempty"`
}

func (m *PongRequest) Reset()                    { *m = PongRequest{} }
func (m *PongRequest) String() string            { return proto.CompactTextString(m) }
func (*PongRequest) ProtoMessage()               {}
func (*PongRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *PongRequest) GetMultiAddress() *MultiAddress {
	if m != nil {
		return m.MultiAddress
	}
	return nil
}

type PongResponse struct {
}

func (m *PongResponse) Reset()                    { *m = PongResponse{} }
func (m *PongResponse) String() string            { return proto.CompactTextString(m) }
func (*PongResponse) ProtoMessage()               {}
func (*PongResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type QueryRequest struct {
	Address string `protobuf:"bytes,1,opt,name=address" json:"address,omitempty"`
}

func (m *QueryRequest) Reset()                    { *m = QueryRequest{} }
func (m *QueryRequest) String() string            { return proto.CompactTextString(m) }
func (*QueryRequest) ProtoMessage()               {}
func (*QueryRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *QueryRequest) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

type QueryResponse struct {
	MultiAddresses []*MultiAddress `protobuf:"bytes,1,rep,name=multiAddresses" json:"multiAddresses,omitempty"`
}

func (m *QueryResponse) Reset()                    { *m = QueryResponse{} }
func (m *QueryResponse) String() string            { return proto.CompactTextString(m) }
func (*QueryResponse) ProtoMessage()               {}
func (*QueryResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *QueryResponse) GetMultiAddresses() []*MultiAddress {
	if m != nil {
		return m.MultiAddresses
	}
	return nil
}

type StreamMessage struct {
	Signature []byte `protobuf:"bytes,1,opt,name=signature,proto3" json:"signature,omitempty"`
	Address   string `protobuf:"bytes,2,opt,name=address" json:"address,omitempty"`
	Network   []byte `protobuf:"bytes,3,opt,name=network,proto3" json:"network,omitempty"`
	Data      []byte `protobuf:"bytes,4,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *StreamMessage) Reset()                    { *m = StreamMessage{} }
func (m *StreamMessage) String() string            { return proto.CompactTextString(m) }
func (*StreamMessage) ProtoMessage()               {}
func (*StreamMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *StreamMessage) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

func (m *StreamMessage) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *StreamMessage) GetNetwork() []byte {
	if m != nil {
		return m.Network
	}
	return nil
}

func (m *StreamMessage) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type OpenOrderRequest struct {
	OrderFragment *EncryptedOrderFragment `protobuf:"bytes,1,opt,name=orderFragment" json:"orderFragment,omitempty"`
}

func (m *OpenOrderRequest) Reset()                    { *m = OpenOrderRequest{} }
func (m *OpenOrderRequest) String() string            { return proto.CompactTextString(m) }
func (*OpenOrderRequest) ProtoMessage()               {}
func (*OpenOrderRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *OpenOrderRequest) GetOrderFragment() *EncryptedOrderFragment {
	if m != nil {
		return m.OrderFragment
	}
	return nil
}

type OpenOrderResponse struct {
}

func (m *OpenOrderResponse) Reset()                    { *m = OpenOrderResponse{} }
func (m *OpenOrderResponse) String() string            { return proto.CompactTextString(m) }
func (*OpenOrderResponse) ProtoMessage()               {}
func (*OpenOrderResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

type EncryptedOrderFragment struct {
	OrderId         []byte                              `protobuf:"bytes,1,opt,name=orderId,proto3" json:"orderId,omitempty"`
	OrderType       OrderType                           `protobuf:"varint,2,opt,name=orderType,enum=grpc.OrderType" json:"orderType,omitempty"`
	OrderParity     OrderParity                         `protobuf:"varint,3,opt,name=orderParity,enum=grpc.OrderParity" json:"orderParity,omitempty"`
	OrderSettlement OrderSettlement                     `protobuf:"varint,4,opt,name=orderSettlement,enum=grpc.OrderSettlement" json:"orderSettlement,omitempty"`
	OrderExpiry     int64                               `protobuf:"varint,5,opt,name=orderExpiry" json:"orderExpiry,omitempty"`
	Id              []byte                              `protobuf:"bytes,6,opt,name=id,proto3" json:"id,omitempty"`
	EpochDepth      int32                               `protobuf:"varint,7,opt,name=epochDepth" json:"epochDepth,omitempty"`
	Tokens          []byte                              `protobuf:"bytes,8,opt,name=tokens,proto3" json:"tokens,omitempty"`
	Price           *EncryptedCoExpShare                `protobuf:"bytes,9,opt,name=price" json:"price,omitempty"`
	Volume          *EncryptedCoExpShare                `protobuf:"bytes,10,opt,name=volume" json:"volume,omitempty"`
	MinimumVolume   *EncryptedCoExpShare                `protobuf:"bytes,11,opt,name=minimumVolume" json:"minimumVolume,omitempty"`
	Nonce           []byte                              `protobuf:"bytes,12,opt,name=nonce,proto3" json:"nonce,omitempty"`
	Blinding        []byte                              `protobuf:"bytes,13,opt,name=blinding,proto3" json:"blinding,omitempty"`
	Commitments     map[uint64]*OrderFragmentCommitment `protobuf:"bytes,14,rep,name=commitments" json:"commitments,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *EncryptedOrderFragment) Reset()                    { *m = EncryptedOrderFragment{} }
func (m *EncryptedOrderFragment) String() string            { return proto.CompactTextString(m) }
func (*EncryptedOrderFragment) ProtoMessage()               {}
func (*EncryptedOrderFragment) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func (m *EncryptedOrderFragment) GetOrderId() []byte {
	if m != nil {
		return m.OrderId
	}
	return nil
}

func (m *EncryptedOrderFragment) GetOrderType() OrderType {
	if m != nil {
		return m.OrderType
	}
	return OrderType_Midpoint
}

func (m *EncryptedOrderFragment) GetOrderParity() OrderParity {
	if m != nil {
		return m.OrderParity
	}
	return OrderParity_Buy
}

func (m *EncryptedOrderFragment) GetOrderSettlement() OrderSettlement {
	if m != nil {
		return m.OrderSettlement
	}
	return OrderSettlement_Nil
}

func (m *EncryptedOrderFragment) GetOrderExpiry() int64 {
	if m != nil {
		return m.OrderExpiry
	}
	return 0
}

func (m *EncryptedOrderFragment) GetId() []byte {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *EncryptedOrderFragment) GetEpochDepth() int32 {
	if m != nil {
		return m.EpochDepth
	}
	return 0
}

func (m *EncryptedOrderFragment) GetTokens() []byte {
	if m != nil {
		return m.Tokens
	}
	return nil
}

func (m *EncryptedOrderFragment) GetPrice() *EncryptedCoExpShare {
	if m != nil {
		return m.Price
	}
	return nil
}

func (m *EncryptedOrderFragment) GetVolume() *EncryptedCoExpShare {
	if m != nil {
		return m.Volume
	}
	return nil
}

func (m *EncryptedOrderFragment) GetMinimumVolume() *EncryptedCoExpShare {
	if m != nil {
		return m.MinimumVolume
	}
	return nil
}

func (m *EncryptedOrderFragment) GetNonce() []byte {
	if m != nil {
		return m.Nonce
	}
	return nil
}

func (m *EncryptedOrderFragment) GetBlinding() []byte {
	if m != nil {
		return m.Blinding
	}
	return nil
}

func (m *EncryptedOrderFragment) GetCommitments() map[uint64]*OrderFragmentCommitment {
	if m != nil {
		return m.Commitments
	}
	return nil
}

type EncryptedCoExpShare struct {
	Co  []byte `protobuf:"bytes,1,opt,name=co,proto3" json:"co,omitempty"`
	Exp []byte `protobuf:"bytes,2,opt,name=exp,proto3" json:"exp,omitempty"`
}

func (m *EncryptedCoExpShare) Reset()                    { *m = EncryptedCoExpShare{} }
func (m *EncryptedCoExpShare) String() string            { return proto.CompactTextString(m) }
func (*EncryptedCoExpShare) ProtoMessage()               {}
func (*EncryptedCoExpShare) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

func (m *EncryptedCoExpShare) GetCo() []byte {
	if m != nil {
		return m.Co
	}
	return nil
}

func (m *EncryptedCoExpShare) GetExp() []byte {
	if m != nil {
		return m.Exp
	}
	return nil
}

type OrderFragmentCommitment struct {
	PriceCo          []byte `protobuf:"bytes,1,opt,name=priceCo,proto3" json:"priceCo,omitempty"`
	PriceExp         []byte `protobuf:"bytes,2,opt,name=priceExp,proto3" json:"priceExp,omitempty"`
	VolumeCo         []byte `protobuf:"bytes,3,opt,name=volumeCo,proto3" json:"volumeCo,omitempty"`
	VolumeExp        []byte `protobuf:"bytes,4,opt,name=volumeExp,proto3" json:"volumeExp,omitempty"`
	MinimumVolumeCo  []byte `protobuf:"bytes,5,opt,name=minimumVolumeCo,proto3" json:"minimumVolumeCo,omitempty"`
	MinimumVolumeExp []byte `protobuf:"bytes,6,opt,name=minimumVolumeExp,proto3" json:"minimumVolumeExp,omitempty"`
}

func (m *OrderFragmentCommitment) Reset()                    { *m = OrderFragmentCommitment{} }
func (m *OrderFragmentCommitment) String() string            { return proto.CompactTextString(m) }
func (*OrderFragmentCommitment) ProtoMessage()               {}
func (*OrderFragmentCommitment) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{12} }

func (m *OrderFragmentCommitment) GetPriceCo() []byte {
	if m != nil {
		return m.PriceCo
	}
	return nil
}

func (m *OrderFragmentCommitment) GetPriceExp() []byte {
	if m != nil {
		return m.PriceExp
	}
	return nil
}

func (m *OrderFragmentCommitment) GetVolumeCo() []byte {
	if m != nil {
		return m.VolumeCo
	}
	return nil
}

func (m *OrderFragmentCommitment) GetVolumeExp() []byte {
	if m != nil {
		return m.VolumeExp
	}
	return nil
}

func (m *OrderFragmentCommitment) GetMinimumVolumeCo() []byte {
	if m != nil {
		return m.MinimumVolumeCo
	}
	return nil
}

func (m *OrderFragmentCommitment) GetMinimumVolumeExp() []byte {
	if m != nil {
		return m.MinimumVolumeExp
	}
	return nil
}

type CoExpCommitment struct {
	Co  []byte `protobuf:"bytes,1,opt,name=co,proto3" json:"co,omitempty"`
	Exp []byte `protobuf:"bytes,2,opt,name=exp,proto3" json:"exp,omitempty"`
}

func (m *CoExpCommitment) Reset()                    { *m = CoExpCommitment{} }
func (m *CoExpCommitment) String() string            { return proto.CompactTextString(m) }
func (*CoExpCommitment) ProtoMessage()               {}
func (*CoExpCommitment) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{13} }

func (m *CoExpCommitment) GetCo() []byte {
	if m != nil {
		return m.Co
	}
	return nil
}

func (m *CoExpCommitment) GetExp() []byte {
	if m != nil {
		return m.Exp
	}
	return nil
}

type StatusRequest struct {
}

func (m *StatusRequest) Reset()                    { *m = StatusRequest{} }
func (m *StatusRequest) String() string            { return proto.CompactTextString(m) }
func (*StatusRequest) ProtoMessage()               {}
func (*StatusRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{14} }

type StatusResponse struct {
	Address      string `protobuf:"bytes,1,opt,name=address" json:"address,omitempty"`
	Bootstrapped bool   `protobuf:"varint,2,opt,name=bootstrapped" json:"bootstrapped,omitempty"`
	Peers        int64  `protobuf:"varint,3,opt,name=peers" json:"peers,omitempty"`
}

func (m *StatusResponse) Reset()                    { *m = StatusResponse{} }
func (m *StatusResponse) String() string            { return proto.CompactTextString(m) }
func (*StatusResponse) ProtoMessage()               {}
func (*StatusResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{15} }

func (m *StatusResponse) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *StatusResponse) GetBootstrapped() bool {
	if m != nil {
		return m.Bootstrapped
	}
	return false
}

func (m *StatusResponse) GetPeers() int64 {
	if m != nil {
		return m.Peers
	}
	return 0
}

type UpdateMidpointRequest struct {
	Signature []byte            `protobuf:"bytes,1,opt,name=signature,proto3" json:"signature,omitempty"`
	Prices    map[uint64]uint64 `protobuf:"bytes,2,rep,name=prices" json:"prices,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"varint,2,opt,name=value"`
	Nonce     uint64            `protobuf:"varint,3,opt,name=nonce" json:"nonce,omitempty"`
}

func (m *UpdateMidpointRequest) Reset()                    { *m = UpdateMidpointRequest{} }
func (m *UpdateMidpointRequest) String() string            { return proto.CompactTextString(m) }
func (*UpdateMidpointRequest) ProtoMessage()               {}
func (*UpdateMidpointRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{16} }

func (m *UpdateMidpointRequest) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

func (m *UpdateMidpointRequest) GetPrices() map[uint64]uint64 {
	if m != nil {
		return m.Prices
	}
	return nil
}

func (m *UpdateMidpointRequest) GetNonce() uint64 {
	if m != nil {
		return m.Nonce
	}
	return 0
}

type UpdateMidpointResponse struct {
}

func (m *UpdateMidpointResponse) Reset()                    { *m = UpdateMidpointResponse{} }
func (m *UpdateMidpointResponse) String() string            { return proto.CompactTextString(m) }
func (*UpdateMidpointResponse) ProtoMessage()               {}
func (*UpdateMidpointResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{17} }

func init() {
	proto.RegisterType((*MultiAddress)(nil), "grpc.MultiAddress")
	proto.RegisterType((*PingRequest)(nil), "grpc.PingRequest")
	proto.RegisterType((*PingResponse)(nil), "grpc.PingResponse")
	proto.RegisterType((*PongRequest)(nil), "grpc.PongRequest")
	proto.RegisterType((*PongResponse)(nil), "grpc.PongResponse")
	proto.RegisterType((*QueryRequest)(nil), "grpc.QueryRequest")
	proto.RegisterType((*QueryResponse)(nil), "grpc.QueryResponse")
	proto.RegisterType((*StreamMessage)(nil), "grpc.StreamMessage")
	proto.RegisterType((*OpenOrderRequest)(nil), "grpc.OpenOrderRequest")
	proto.RegisterType((*OpenOrderResponse)(nil), "grpc.OpenOrderResponse")
	proto.RegisterType((*EncryptedOrderFragment)(nil), "grpc.EncryptedOrderFragment")
	proto.RegisterType((*EncryptedCoExpShare)(nil), "grpc.EncryptedCoExpShare")
	proto.RegisterType((*OrderFragmentCommitment)(nil), "grpc.OrderFragmentCommitment")
	proto.RegisterType((*CoExpCommitment)(nil), "grpc.CoExpCommitment")
	proto.RegisterType((*StatusRequest)(nil), "grpc.StatusRequest")
	proto.RegisterType((*StatusResponse)(nil), "grpc.StatusResponse")
	proto.RegisterType((*UpdateMidpointRequest)(nil), "grpc.UpdateMidpointRequest")
	proto.RegisterType((*UpdateMidpointResponse)(nil), "grpc.UpdateMidpointResponse")
	proto.RegisterEnum("grpc.OrderType", OrderType_name, OrderType_value)
	proto.RegisterEnum("grpc.OrderParity", OrderParity_name, OrderParity_value)
	proto.RegisterEnum("grpc.OrderSettlement", OrderSettlement_name, OrderSettlement_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc1.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc1.SupportPackageIsVersion4

// Client API for SwarmService service

type SwarmServiceClient interface {
	Ping(ctx context.Context, in *PingRequest, opts ...grpc1.CallOption) (*PingResponse, error)
	Pong(ctx context.Context, in *PongRequest, opts ...grpc1.CallOption) (*PongResponse, error)
	Query(ctx context.Context, in *QueryRequest, opts ...grpc1.CallOption) (*QueryResponse, error)
}

type swarmServiceClient struct {
	cc *grpc1.ClientConn
}

func NewSwarmServiceClient(cc *grpc1.ClientConn) SwarmServiceClient {
	return &swarmServiceClient{cc}
}

func (c *swarmServiceClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc1.CallOption) (*PingResponse, error) {
	out := new(PingResponse)
	err := grpc1.Invoke(ctx, "/grpc.SwarmService/Ping", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *swarmServiceClient) Pong(ctx context.Context, in *PongRequest, opts ...grpc1.CallOption) (*PongResponse, error) {
	out := new(PongResponse)
	err := grpc1.Invoke(ctx, "/grpc.SwarmService/Pong", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *swarmServiceClient) Query(ctx context.Context, in *QueryRequest, opts ...grpc1.CallOption) (*QueryResponse, error) {
	out := new(QueryResponse)
	err := grpc1.Invoke(ctx, "/grpc.SwarmService/Query", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for SwarmService service

type SwarmServiceServer interface {
	Ping(context.Context, *PingRequest) (*PingResponse, error)
	Pong(context.Context, *PongRequest) (*PongResponse, error)
	Query(context.Context, *QueryRequest) (*QueryResponse, error)
}

func RegisterSwarmServiceServer(s *grpc1.Server, srv SwarmServiceServer) {
	s.RegisterService(&_SwarmService_serviceDesc, srv)
}

func _SwarmService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SwarmServiceServer).Ping(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.SwarmService/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SwarmServiceServer).Ping(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SwarmService_Pong_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(PongRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SwarmServiceServer).Pong(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.SwarmService/Pong",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SwarmServiceServer).Pong(ctx, req.(*PongRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SwarmService_Query_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SwarmServiceServer).Query(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.SwarmService/Query",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SwarmServiceServer).Query(ctx, req.(*QueryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _SwarmService_serviceDesc = grpc1.ServiceDesc{
	ServiceName: "grpc.SwarmService",
	HandlerType: (*SwarmServiceServer)(nil),
	Methods: []grpc1.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _SwarmService_Ping_Handler,
		},
		{
			MethodName: "Pong",
			Handler:    _SwarmService_Pong_Handler,
		},
		{
			MethodName: "Query",
			Handler:    _SwarmService_Query_Handler,
		},
	},
	Streams:  []grpc1.StreamDesc{},
	Metadata: "grpc.proto",
}

// Client API for StreamService service

type StreamServiceClient interface {
	Connect(ctx context.Context, opts ...grpc1.CallOption) (StreamService_ConnectClient, error)
}

type streamServiceClient struct {
	cc *grpc1.ClientConn
}

func NewStreamServiceClient(cc *grpc1.ClientConn) StreamServiceClient {
	return &streamServiceClient{cc}
}

func (c *streamServiceClient) Connect(ctx context.Context, opts ...grpc1.CallOption) (StreamService_ConnectClient, error) {
	stream, err := grpc1.NewClientStream(ctx, &_StreamService_serviceDesc.Streams[0], c.cc, "/grpc.StreamService/Connect", opts...)
	if err != nil {
		return nil, err
	}
	x := &streamServiceConnectClient{stream}
	return x, nil
}

type StreamService_ConnectClient interface {
	Send(*StreamMessage) error
	Recv() (*StreamMessage, error)
	grpc1.ClientStream
}

type streamServiceConnectClient struct {
	grpc1.ClientStream
}

func (x *streamServiceConnectClient) Send(m *StreamMessage) error {
	return x.ClientStream.SendMsg(m)
}

func (x *streamServiceConnectClient) Recv() (*StreamMessage, error) {
	m := new(StreamMessage)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for StreamService service

type StreamServiceServer interface {
	Connect(StreamService_ConnectServer) error
}

func RegisterStreamServiceServer(s *grpc1.Server, srv StreamServiceServer) {
	s.RegisterService(&_StreamService_serviceDesc, srv)
}

func _StreamService_Connect_Handler(srv interface{}, stream grpc1.ServerStream) error {
	return srv.(StreamServiceServer).Connect(&streamServiceConnectServer{stream})
}

type StreamService_ConnectServer interface {
	Send(*StreamMessage) error
	Recv() (*StreamMessage, error)
	grpc1.ServerStream
}

type streamServiceConnectServer struct {
	grpc1.ServerStream
}

func (x *streamServiceConnectServer) Send(m *StreamMessage) error {
	return x.ServerStream.SendMsg(m)
}

func (x *streamServiceConnectServer) Recv() (*StreamMessage, error) {
	m := new(StreamMessage)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _StreamService_serviceDesc = grpc1.ServiceDesc{
	ServiceName: "grpc.StreamService",
	HandlerType: (*StreamServiceServer)(nil),
	Methods:     []grpc1.MethodDesc{},
	Streams: []grpc1.StreamDesc{
		{
			StreamName:    "Connect",
			Handler:       _StreamService_Connect_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "grpc.proto",
}

// Client API for OrderbookService service

type OrderbookServiceClient interface {
	OpenOrder(ctx context.Context, in *OpenOrderRequest, opts ...grpc1.CallOption) (*OpenOrderResponse, error)
}

type orderbookServiceClient struct {
	cc *grpc1.ClientConn
}

func NewOrderbookServiceClient(cc *grpc1.ClientConn) OrderbookServiceClient {
	return &orderbookServiceClient{cc}
}

func (c *orderbookServiceClient) OpenOrder(ctx context.Context, in *OpenOrderRequest, opts ...grpc1.CallOption) (*OpenOrderResponse, error) {
	out := new(OpenOrderResponse)
	err := grpc1.Invoke(ctx, "/grpc.OrderbookService/OpenOrder", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for OrderbookService service

type OrderbookServiceServer interface {
	OpenOrder(context.Context, *OpenOrderRequest) (*OpenOrderResponse, error)
}

func RegisterOrderbookServiceServer(s *grpc1.Server, srv OrderbookServiceServer) {
	s.RegisterService(&_OrderbookService_serviceDesc, srv)
}

func _OrderbookService_OpenOrder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(OpenOrderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrderbookServiceServer).OpenOrder(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.OrderbookService/OpenOrder",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrderbookServiceServer).OpenOrder(ctx, req.(*OpenOrderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _OrderbookService_serviceDesc = grpc1.ServiceDesc{
	ServiceName: "grpc.OrderbookService",
	HandlerType: (*OrderbookServiceServer)(nil),
	Methods: []grpc1.MethodDesc{
		{
			MethodName: "OpenOrder",
			Handler:    _OrderbookService_OpenOrder_Handler,
		},
	},
	Streams:  []grpc1.StreamDesc{},
	Metadata: "grpc.proto",
}

// Client API for StatusService service

type StatusServiceClient interface {
	Status(ctx context.Context, in *StatusRequest, opts ...grpc1.CallOption) (*StatusResponse, error)
}

type statusServiceClient struct {
	cc *grpc1.ClientConn
}

func NewStatusServiceClient(cc *grpc1.ClientConn) StatusServiceClient {
	return &statusServiceClient{cc}
}

func (c *statusServiceClient) Status(ctx context.Context, in *StatusRequest, opts ...grpc1.CallOption) (*StatusResponse, error) {
	out := new(StatusResponse)
	err := grpc1.Invoke(ctx, "/grpc.StatusService/Status", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for StatusService service

type StatusServiceServer interface {
	Status(context.Context, *StatusRequest) (*StatusResponse, error)
}

func RegisterStatusServiceServer(s *grpc1.Server, srv StatusServiceServer) {
	s.RegisterService(&_StatusService_serviceDesc, srv)
}

func _StatusService_Status_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StatusServiceServer).Status(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.StatusService/Status",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StatusServiceServer).Status(ctx, req.(*StatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _StatusService_serviceDesc = grpc1.ServiceDesc{
	ServiceName: "grpc.StatusService",
	HandlerType: (*StatusServiceServer)(nil),
	Methods: []grpc1.MethodDesc{
		{
			MethodName: "Status",
			Handler:    _StatusService_Status_Handler,
		},
	},
	Streams:  []grpc1.StreamDesc{},
	Metadata: "grpc.proto",
}

// Client API for OracleService service

type OracleServiceClient interface {
	UpdateMidpoint(ctx context.Context, in *UpdateMidpointRequest, opts ...grpc1.CallOption) (*UpdateMidpointResponse, error)
}

type oracleServiceClient struct {
	cc *grpc1.ClientConn
}

func NewOracleServiceClient(cc *grpc1.ClientConn) OracleServiceClient {
	return &oracleServiceClient{cc}
}

func (c *oracleServiceClient) UpdateMidpoint(ctx context.Context, in *UpdateMidpointRequest, opts ...grpc1.CallOption) (*UpdateMidpointResponse, error) {
	out := new(UpdateMidpointResponse)
	err := grpc1.Invoke(ctx, "/grpc.OracleService/UpdateMidpoint", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for OracleService service

type OracleServiceServer interface {
	UpdateMidpoint(context.Context, *UpdateMidpointRequest) (*UpdateMidpointResponse, error)
}

func RegisterOracleServiceServer(s *grpc1.Server, srv OracleServiceServer) {
	s.RegisterService(&_OracleService_serviceDesc, srv)
}

func _OracleService_UpdateMidpoint_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateMidpointRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OracleServiceServer).UpdateMidpoint(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.OracleService/UpdateMidpoint",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OracleServiceServer).UpdateMidpoint(ctx, req.(*UpdateMidpointRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _OracleService_serviceDesc = grpc1.ServiceDesc{
	ServiceName: "grpc.OracleService",
	HandlerType: (*OracleServiceServer)(nil),
	Methods: []grpc1.MethodDesc{
		{
			MethodName: "UpdateMidpoint",
			Handler:    _OracleService_UpdateMidpoint_Handler,
		},
	},
	Streams:  []grpc1.StreamDesc{},
	Metadata: "grpc.proto",
}

func init() { proto.RegisterFile("grpc.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 1059 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x56, 0x5f, 0x6f, 0xe3, 0x44,
	0x10, 0x3f, 0xe7, 0x5f, 0x9b, 0x89, 0x93, 0xb8, 0xd3, 0x5e, 0xcf, 0x84, 0x82, 0x22, 0xbf, 0x10,
	0x55, 0xb4, 0x1c, 0xa9, 0x74, 0x07, 0x27, 0xa4, 0xea, 0x2e, 0x97, 0x13, 0xa8, 0xf4, 0x52, 0x36,
	0x70, 0x4f, 0x20, 0xe4, 0xc6, 0xab, 0xd4, 0x6a, 0xec, 0x35, 0xeb, 0x4d, 0xaf, 0x79, 0xe1, 0xa3,
	0xf0, 0x75, 0xe0, 0x5b, 0xf0, 0x55, 0xd0, 0xee, 0xda, 0xc9, 0x3a, 0x4d, 0xdb, 0x17, 0xde, 0x76,
	0x66, 0x7e, 0xf3, 0xf3, 0xec, 0xcc, 0xec, 0x8c, 0x01, 0xa6, 0x3c, 0x99, 0x1c, 0x27, 0x9c, 0x09,
	0x86, 0x15, 0x79, 0xf6, 0xfe, 0x04, 0xfb, 0x7c, 0x3e, 0x13, 0xe1, 0xeb, 0x20, 0xe0, 0x34, 0x4d,
	0xf1, 0x00, 0xea, 0x69, 0x38, 0x8d, 0x7d, 0x31, 0xe7, 0xd4, 0xb5, 0xba, 0x56, 0xcf, 0x26, 0x2b,
	0x05, 0x7a, 0x60, 0x47, 0x06, 0xda, 0x2d, 0x75, 0xad, 0x5e, 0x9d, 0x14, 0x74, 0xf8, 0x25, 0xec,
	0x98, 0xf2, 0x7b, 0x16, 0x4f, 0xa8, 0x5b, 0xee, 0x5a, 0xbd, 0x0a, 0xb9, 0x6b, 0xf0, 0x86, 0xd0,
	0xb8, 0x08, 0xe3, 0x29, 0xa1, 0x7f, 0xcc, 0x69, 0x2a, 0xf0, 0xc5, 0xda, 0x07, 0x64, 0x04, 0x8d,
	0x3e, 0x1e, 0xab, 0xb8, 0xcd, 0x40, 0x8b, 0x1f, 0xf5, 0x5a, 0x60, 0x6b, 0x9a, 0x34, 0x61, 0x71,
	0xaa, 0x69, 0xd9, 0xff, 0x43, 0xcb, 0x0c, 0xda, 0x1e, 0xd8, 0x3f, 0xcd, 0x29, 0x5f, 0xe4, 0xbc,
	0x2e, 0x6c, 0xf9, 0x06, 0x65, 0x9d, 0xe4, 0xa2, 0x77, 0x06, 0xcd, 0x0c, 0xa9, 0x5d, 0xf1, 0x15,
	0xb4, 0x4c, 0x6a, 0x2a, 0x3d, 0xca, 0xf7, 0x04, 0xb1, 0x86, 0xf4, 0xe6, 0xd0, 0x1c, 0x0b, 0x4e,
	0xfd, 0xe8, 0x9c, 0xa6, 0xa9, 0x3f, 0xa5, 0x8f, 0x54, 0xc9, 0x88, 0xaa, 0x54, 0x88, 0x4a, 0x5a,
	0x62, 0x2a, 0x3e, 0x32, 0x7e, 0xad, 0x2a, 0x62, 0x93, 0x5c, 0x44, 0x84, 0x4a, 0xe0, 0x0b, 0xdf,
	0xad, 0x28, 0xb5, 0x3a, 0x7b, 0x1f, 0xc0, 0x19, 0x25, 0x34, 0x1e, 0xf1, 0x80, 0xf2, 0xfc, 0xc6,
	0x6f, 0xa0, 0xc9, 0xa4, 0xfc, 0x8e, 0xfb, 0xd3, 0x88, 0xc6, 0x22, 0x4b, 0xe5, 0x81, 0xbe, 0xc5,
	0x30, 0x9e, 0xf0, 0x45, 0x22, 0x68, 0x30, 0x32, 0x31, 0xa4, 0xe8, 0xe2, 0xed, 0xc2, 0x8e, 0xc1,
	0x9b, 0xa5, 0xf6, 0x9f, 0x2a, 0xec, 0x6f, 0x76, 0x97, 0x51, 0x2b, 0x82, 0x1f, 0x82, 0xec, 0xae,
	0xb9, 0x88, 0x47, 0x50, 0x57, 0xc7, 0x9f, 0x17, 0x09, 0x55, 0x77, 0x6d, 0xf5, 0xdb, 0x3a, 0x92,
	0x51, 0xae, 0x26, 0x2b, 0x04, 0x9e, 0x40, 0x43, 0x09, 0x17, 0x3e, 0x0f, 0xc5, 0x42, 0xa5, 0xa0,
	0xd5, 0xdf, 0x31, 0x1c, 0xb4, 0x81, 0x98, 0x28, 0x3c, 0x85, 0xb6, 0x12, 0xc7, 0x54, 0x88, 0x19,
	0x55, 0x77, 0xae, 0x28, 0xc7, 0xa7, 0x86, 0xe3, 0xca, 0x48, 0xd6, 0xd1, 0xd8, 0xcd, 0xbe, 0x3a,
	0xbc, 0x4d, 0x42, 0xbe, 0x70, 0xab, 0x5d, 0xab, 0x57, 0x26, 0xa6, 0x0a, 0x5b, 0x50, 0x0a, 0x03,
	0xb7, 0xa6, 0xee, 0x56, 0x0a, 0x03, 0xfc, 0x1c, 0x80, 0x26, 0x6c, 0x72, 0xf5, 0x96, 0x26, 0xe2,
	0xca, 0xdd, 0xea, 0x5a, 0xbd, 0x2a, 0x31, 0x34, 0xb8, 0x0f, 0x35, 0xc1, 0xae, 0x69, 0x9c, 0xba,
	0xdb, 0xca, 0x27, 0x93, 0xf0, 0x2b, 0xa8, 0x26, 0x3c, 0x9c, 0x50, 0xb7, 0xae, 0x8a, 0xf2, 0xc9,
	0x5a, 0x51, 0x06, 0x6c, 0x78, 0x9b, 0x8c, 0xaf, 0x7c, 0x4e, 0x89, 0xc6, 0xe1, 0xd7, 0x50, 0xbb,
	0x61, 0xb3, 0x79, 0x44, 0x5d, 0x78, 0xcc, 0x23, 0x03, 0xe2, 0x29, 0x34, 0xa3, 0x30, 0x0e, 0xa3,
	0x79, 0xf4, 0x41, 0x7b, 0x36, 0x1e, 0xf3, 0x2c, 0xe2, 0x71, 0x0f, 0xaa, 0xb1, 0x9a, 0x09, 0xb6,
	0x8a, 0x5d, 0x0b, 0xd8, 0x81, 0xed, 0xcb, 0x59, 0x18, 0x07, 0x61, 0x3c, 0x75, 0x9b, 0xca, 0xb0,
	0x94, 0x71, 0x04, 0x8d, 0x09, 0x8b, 0xa2, 0x50, 0xc8, 0x74, 0xa6, 0x6e, 0x4b, 0xbd, 0x9b, 0xa3,
	0x87, 0x3a, 0xee, 0x78, 0xb0, 0xc2, 0x0f, 0x63, 0xc1, 0x17, 0xc4, 0x64, 0xe8, 0xfc, 0x06, 0xce,
	0x3a, 0x00, 0x1d, 0x28, 0x5f, 0xd3, 0x85, 0x6a, 0xb0, 0x0a, 0x91, 0x47, 0x3c, 0x81, 0xea, 0x8d,
	0x3f, 0x9b, 0xeb, 0xc6, 0x6a, 0xf4, 0x3f, 0x33, 0xca, 0x9d, 0x7f, 0x67, 0xc5, 0x42, 0x34, 0xf6,
	0x55, 0xe9, 0x1b, 0xcb, 0x7b, 0x09, 0xbb, 0x1b, 0xf2, 0x20, 0xab, 0x3c, 0x61, 0x59, 0x07, 0x97,
	0x26, 0x4c, 0x7e, 0x91, 0xde, 0x26, 0x8a, 0xdd, 0x26, 0xf2, 0xe8, 0xfd, 0x6b, 0xc1, 0xb3, 0x7b,
	0xf8, 0xe5, 0x23, 0x50, 0x35, 0x1b, 0xe4, 0x14, 0xb9, 0x28, 0x53, 0xa7, 0x8e, 0xc3, 0x25, 0xd9,
	0x52, 0x96, 0x36, 0x5d, 0xb7, 0x01, 0xcb, 0x5e, 0xfc, 0x52, 0x96, 0x43, 0x44, 0x9f, 0xa5, 0xa3,
	0x7e, 0xf7, 0x2b, 0x05, 0xf6, 0xa0, 0x5d, 0xa8, 0xdb, 0x80, 0xa9, 0xce, 0xb5, 0xc9, 0xba, 0x1a,
	0x0f, 0xc1, 0x29, 0xa8, 0x24, 0x9d, 0xee, 0xe5, 0x3b, 0x7a, 0xef, 0x04, 0xda, 0x2a, 0x23, 0xc6,
	0xc5, 0x1e, 0x4f, 0x4b, 0x5b, 0x8e, 0x3f, 0x5f, 0xcc, 0xd3, 0x6c, 0x08, 0x79, 0x01, 0xb4, 0x72,
	0x45, 0x36, 0x5d, 0xef, 0x1d, 0xc4, 0x72, 0x65, 0x5d, 0x32, 0x26, 0x52, 0xc1, 0xfd, 0x24, 0xa1,
	0x81, 0xe2, 0xdd, 0x26, 0x05, 0x9d, 0x6c, 0xc9, 0x84, 0x52, 0x9e, 0xaa, 0x14, 0x95, 0x89, 0x16,
	0xbc, 0xbf, 0x2d, 0x78, 0xfa, 0x4b, 0x12, 0xf8, 0x82, 0x9e, 0x87, 0x41, 0xc2, 0xc2, 0x58, 0xe4,
	0x43, 0xf0, 0xe1, 0xf1, 0x7b, 0x0a, 0x35, 0x95, 0x7f, 0x39, 0x7d, 0x65, 0xa7, 0x7e, 0xa1, 0x1b,
	0x67, 0x23, 0xd5, 0xf1, 0x85, 0x42, 0xea, 0x1e, 0xcd, 0xdc, 0x56, 0x2f, 0x44, 0x6f, 0x4d, 0x2d,
	0x74, 0xbe, 0x85, 0x86, 0x01, 0xde, 0xd0, 0xaf, 0x7b, 0x66, 0xbf, 0x56, 0xcc, 0x86, 0x74, 0x61,
	0x7f, 0xfd, 0xeb, 0x3a, 0x6f, 0x87, 0x43, 0xa8, 0x2f, 0x27, 0x25, 0xda, 0xb0, 0x9d, 0x03, 0x9c,
	0x27, 0x58, 0x87, 0xea, 0x8f, 0x61, 0x14, 0x0a, 0xc7, 0x42, 0x07, 0xec, 0xdc, 0xf0, 0xfb, 0xbb,
	0xd1, 0x99, 0x53, 0xc2, 0x26, 0xd4, 0x95, 0x51, 0x89, 0xe5, 0xc3, 0x2e, 0x34, 0x8c, 0xf9, 0x89,
	0x5b, 0x50, 0x7e, 0x33, 0x5f, 0x38, 0x4f, 0x70, 0x1b, 0x2a, 0x63, 0x3a, 0x9b, 0x39, 0xd6, 0xe1,
	0x0b, 0x68, 0xaf, 0x0d, 0x4a, 0x89, 0x7a, 0x1f, 0xce, 0xf4, 0x97, 0x08, 0x8d, 0x87, 0xb7, 0x8e,
	0x85, 0x6d, 0x68, 0xa8, 0xe3, 0x6b, 0xc1, 0xa2, 0x70, 0xe2, 0x94, 0xfa, 0x7f, 0x59, 0x60, 0x8f,
	0x3f, 0xfa, 0x3c, 0x1a, 0x53, 0x7e, 0x23, 0x47, 0xd6, 0x11, 0x54, 0xe4, 0xa6, 0xc7, 0x6c, 0x6c,
	0x1b, 0x3f, 0x0f, 0x1d, 0x34, 0x55, 0x59, 0x63, 0x48, 0x38, 0x33, 0xe0, 0xec, 0x2e, 0xdc, 0x58,
	0xf0, 0xf8, 0x1c, 0xaa, 0x6a, 0x6d, 0x63, 0x66, 0x34, 0xb7, 0x7d, 0x67, 0xb7, 0xa0, 0xd3, 0x1e,
	0xfd, 0xef, 0xf3, 0xdd, 0x9c, 0x07, 0xf8, 0x12, 0xb6, 0x06, 0x2c, 0x8e, 0xe9, 0x44, 0x60, 0xe6,
	0x50, 0xd8, 0xdd, 0x9d, 0x4d, 0xca, 0x9e, 0xf5, 0xdc, 0xea, 0x5f, 0x80, 0xa3, 0x52, 0x74, 0xc9,
	0xd8, 0x75, 0x4e, 0xf6, 0x1d, 0xd4, 0x97, 0xab, 0x12, 0xf7, 0xb3, 0x09, 0xb4, 0xb6, 0x93, 0x3b,
	0xcf, 0xee, 0xe8, 0xb3, 0xd8, 0xde, 0xe6, 0x0f, 0x27, 0xa7, 0x3b, 0x81, 0x9a, 0x56, 0xac, 0x42,
	0x33, 0xde, 0x55, 0x67, 0xaf, 0xa8, 0xcc, 0x58, 0x7e, 0x85, 0xe6, 0x88, 0xfb, 0x93, 0x19, 0xcd,
	0x59, 0xce, 0xa0, 0x55, 0x6c, 0x27, 0xfc, 0xf4, 0x81, 0x16, 0xef, 0x1c, 0x6c, 0x36, 0x6a, 0xf6,
	0xcb, 0x9a, 0xfa, 0x1b, 0x3d, 0xf9, 0x2f, 0x00, 0x00, 0xff, 0xff, 0xe8, 0xd5, 0xd1, 0x24, 0x9b,
	0x0a, 0x00, 0x00,
}
