package smpcer

import (
	"errors"
	fmt "fmt"

	"github.com/republicprotocol/republic-go/crypto"
	"github.com/republicprotocol/republic-go/grpc/client"
	"github.com/republicprotocol/republic-go/identity"
	"github.com/republicprotocol/republic-go/order"
	"golang.org/x/net/context"
)

var ErrConnectToSelf = errors.New("connect to self")

type Client struct {
	crypter      crypto.Crypter
	multiAddress identity.MultiAddress
	rendezvous   Rendezvous
	streamer     Streamer
	connPool     *client.ConnPool
}

func NewClient(crypter crypto.Crypter, multiAddress identity.MultiAddress, connPool *client.ConnPool) Client {
	return Client{
		crypter:      crypter,
		multiAddress: multiAddress,
		rendezvous:   NewRendezvous(),
		streamer:     NewStreamer(multiAddress, connPool),
		connPool:     connPool,
	}
}

func (client *Client) OpenOrder(ctx context.Context, multiAddr identity.MultiAddress, orderFragment order.Fragment) error {
	conn, err := client.connPool.Dial(ctx, multiAddr)
	if err != nil {
		return fmt.Errorf("cannot dial %v: %v", multiAddr, err)
	}
	defer conn.Close()

	smpcerClient := NewSmpcClient(conn.ClientConn)
	orderFragmentData, err := MarshalOrderFragment(multiAddr.Address().String(), client.crypter, &orderFragment)
	if err != nil {
		return fmt.Errorf("cannot marshal order fragment: %v", err)
	}
	request := &OpenOrderRequest{
		Signature:     []byte{},
		OrderFragment: orderFragmentData,
	}
	_, err = smpcerClient.OpenOrder(ctx, request)
	return err
}

func (client *Client) CloseOrder(ctx context.Context, multiAddr identity.MultiAddress, orderID []byte) error {
	conn, err := client.connPool.Dial(ctx, multiAddr)
	if err != nil {
		return fmt.Errorf("cannot dial %v:%v", multiAddr, err)
	}
	defer conn.Close()

	smpcerClient := NewSmpcClient(conn.ClientConn)
	request := &CancelOrderRequest{
		Signature: []byte{}, // FIXME: Provide verifiable signature
		OrderId:   orderID,
	}

	_, err = smpcerClient.CancelOrder(ctx, request)
	return err
}

func (client *Client) Compute(ctx context.Context, multiAddress identity.MultiAddress, sender <-chan interface{}) (<-chan interface{}, <-chan interface{}) {
	if client.Address() == multiAddress.Address() {
		// The Client is attempting to connect to itself
		receiver := make(chan interface{})
		defer close(receiver)
		errs := make(chan interface{}, 1)
		defer close(errs)
		errs <- ErrConnectToSelf
		return receiver, errs
	}

	if client.Address() < multiAddress.Address() {
		// The Client should open a gRPC stream
		return client.connect(ctx, multiAddress, sender)
	}

	// The Client must wait for the Smpc service to accept a gRPC stream from
	// a Client on another machine
	return client.wait(ctx, multiAddress, sender), nil
}

// Address of the Client.
func (client *Client) Address() identity.Address {
	return client.multiAddress.Address()
}

// MultiAddress of the Client.
func (client *Client) MultiAddress() identity.MultiAddress {
	return client.multiAddress
}

func (client *Client) connect(ctx context.Context, multiAddress identity.MultiAddress, sender <-chan interface{}) (<-chan interface{}, <-chan interface{}) {
	return client.streamer.connect(multiAddress, ctx.Done(), sender)
}

func (client *Client) wait(ctx context.Context, multiAddress identity.MultiAddress, sender <-chan interface{}) <-chan interface{} {
	return client.rendezvous.wait(multiAddress.Address(), ctx.Done(), sender)
}
