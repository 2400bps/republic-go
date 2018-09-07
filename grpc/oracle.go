package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/republicprotocol/republic-go/identity"
	"github.com/republicprotocol/republic-go/logger"
	"github.com/republicprotocol/republic-go/oracle"
	"github.com/republicprotocol/republic-go/swarm"
	"google.golang.org/grpc/peer"
)

// ErrMidPointPriceIsNil is returned when the midpoint price contains
// nil fields.
var ErrMidPointPriceIsNil = errors.New("midpoint price data is nil")

// ErrMidPointRequestIsNil is returned when a gRPC request is nil or has nil
// fields.
var ErrMidPointRequestIsNil = errors.New("mid-point request is nil")

type oracleClient struct {
	addr  identity.Address
	store swarm.MultiAddressStorer
}

// NewOracleClient returns an object that implements the oracle.Client interface.
func NewOracleClient(addr identity.Address, store swarm.MultiAddressStorer) oracle.Client {
	return &oracleClient{
		addr:  addr,
		store: store,
	}
}

// UpdateMidpoint implements the oracle.Client interface.
func (client *oracleClient) UpdateMidpoint(ctx context.Context, to identity.MultiAddress, midpointPrice oracle.MidpointPrice) error {
	if midpointPrice.IsNil() {
		return ErrMidPointPriceIsNil
	}
	conn, err := Dial(ctx, to)
	if err != nil {
		logger.Network(logger.LevelError, fmt.Sprintf("cannot dial %v: %v", to, err))
		return fmt.Errorf("cannot dial %v: %v", to, err)
	}
	defer conn.Close()

	// Construct a request object and send midpoint information to a given
	// multiaddress.
	request := &UpdateMidpointRequest{
		Signature: midpointPrice.Signature,
		Prices:    midpointPrice.Prices,
		Nonce:     midpointPrice.Nonce,
	}
	if err := Backoff(ctx, func() error {
		_, err = NewOracleServiceClient(conn).UpdateMidpoint(ctx, request)
		return err
	}); err != nil {
		return err
	}

	return nil
}

// MultiAddress implements the oracle.Client interface.
func (client *oracleClient) MultiAddress() identity.MultiAddress {
	multiAddr, err := client.store.MultiAddress(client.addr)
	if err != nil {
		logger.Network(logger.LevelError, fmt.Sprintf("cannot retrieve own multiaddress: %v", err))
		return identity.MultiAddress{}
	}
	return multiAddr
}

// OracleService is a Service that implements the gRPC OracleService defined in
// protobuf. It delegates responsibility for handling the UpdateMidpoint RPCs
// to a oracle.Server.
type OracleService struct {
	server oracle.Server

	rate         time.Duration
	rateLimitsMu *sync.Mutex
	rateLimits   map[string]time.Time
}

// NewOracleService returns an OracleService that uses the oracle.Server as a
// delegate.
func NewOracleService(server oracle.Server, rate time.Duration) OracleService {
	return OracleService{
		server:       server,
		rate:         rate,
		rateLimitsMu: new(sync.Mutex),
		rateLimits:   make(map[string]time.Time),
	}
}

// Register implements the Service interface.
func (service *OracleService) Register(server *Server) {
	if server == nil {
		logger.Network(logger.LevelError, "server is nil")
		return
	}
	RegisterOracleServiceServer(server.Server, service)
}

// UpdateMidpoint is an RPC used to notify a OracleService about updated
// midpoint data. In the UpdateMidpointRequest, the client sends a signed
// MidpointPrice object and the OracleService delegates the responsibility of
// handling this signed object to its oracle.Server. If its oracle.Server
// accepts data from the client it will return an empty UpdateMidpointResponse.
func (service *OracleService) UpdateMidpoint(ctx context.Context, request *UpdateMidpointRequest) (*UpdateMidpointResponse, error) {
	// Check for empty or invalid request fields.
	if request.Signature == nil || len(request.Signature) == 0 || len(request.Prices) == 0 || request.Nonce == 0 {
		return nil, ErrMidPointRequestIsNil
	}

	if err := service.isRateLimited(ctx); err != nil {
		return nil, err
	}

	midpointPrice := oracle.MidpointPrice{
		Signature: request.Signature,
		Prices:    request.Prices,
		Nonce:     request.Nonce,
	}

	return &UpdateMidpointResponse{}, service.server.UpdateMidpoint(ctx, midpointPrice)
}

func (service *OracleService) isRateLimited(ctx context.Context) error {
	client, ok := peer.FromContext(ctx)
	if !ok {
		return fmt.Errorf("failed to get peer from ctx")
	}
	if client.Addr == net.Addr(nil) {
		return fmt.Errorf("failed to get peer address")
	}

	clientAddr := client.Addr.(*net.TCPAddr)
	clientIP := clientAddr.IP.String()

	service.rateLimitsMu.Lock()
	defer service.rateLimitsMu.Unlock()

	if lastPing, ok := service.rateLimits[clientIP]; ok {
		if service.rate > time.Since(lastPing) {
			return ErrRateLimitExceeded
		}
	}

	service.rateLimits[clientIP] = time.Now()
	return nil
}
