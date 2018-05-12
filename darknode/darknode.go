package darknode

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/republicprotocol/republic-go/blockchain/ethereum"
	"github.com/republicprotocol/republic-go/blockchain/ethereum/dnr"
	"github.com/republicprotocol/republic-go/crypto"
	"github.com/republicprotocol/republic-go/darkocean"
	"github.com/republicprotocol/republic-go/dispatch"
	"github.com/republicprotocol/republic-go/grpc"
	"github.com/republicprotocol/republic-go/grpc/status"
	"github.com/republicprotocol/republic-go/identity"
	"github.com/republicprotocol/republic-go/logger"
	"github.com/republicprotocol/republic-go/order"
	"github.com/republicprotocol/republic-go/orderbook"
	"github.com/republicprotocol/republic-go/smpc"
	"google.golang.org/grpc"
)

// Darknodes is an alias.
type Darknodes []Darknode

type Darknode struct {
	Config *Config
	Logger *logger.Logger

	multiAddress identity.MultiAddress
	address      identity.Address
	id           identity.ID
	orderbook    orderbook.Orderbook
	crypter      crypto.Crypter

	darknodeRegistry   dnr.DarknodeRegistry

	orderFragments         chan order.Fragment
	orderFragmentsCanceled chan order.ID

	rpc  *rpc.RPC
	smpc smpc.SmpcComputer
}

// NewDarknode returns a new Darknode.
func NewDarknode(multiAddr identity.MultiAddress, config *Config) (Darknode, error) {
	node := Darknode{
		Config: config,
		Logger: logger.StdoutLogger,
	}

	// Get identity information from the Config
	node.multiAddress = multiAddr
	node.address = node.Config.Address
	node.id = node.address.ID()
	node.orderbook = orderbook.NewOrderbook()

	// Open a connection to the Ethereum network
	transactOpts := bind.NewKeyedTransactor(config.Keystore.EcdsaKey.PrivateKey)

	ethclient, err := ethereum.Connect(config.Ethereum)
	if err != nil {
		return node, err
	}

	// Create bindings to the DarknodeRegistry and Ocean
	darknodeRegistry, err := dnr.NewDarknodeRegistry(context.Background(), ethclient, transactOpts, &bind.CallOpts{})
	if err != nil {
		return Darknode{}, err
	}
	node.darknodeRegistry = darknodeRegistry

	crypter := darkocean.NewCrypter(node.Config.Keystore, node.darknodeRegistry, 256, time.Minute)
	node.crypter = &crypter

	node.orderFragments = make(chan order.Fragment, 1)
	node.orderFragmentsCanceled = make(chan order.ID, 1)
	node.rpc = rpc.NewRPC(node.crypter, node.multiAddress, &node.orderbook)
	node.rpc.OnOpenOrder(func(sig []byte, orderFragment order.Fragment) error {
		if err := node.orderbook.Open(order.Order{
			ID: orderFragment.OrderID,
		}); err != nil {
			return err
		}
		node.orderFragments <- orderFragment
		return nil
	})

	node.rpc.OnCancelOrder(func(sig []byte, orderID order.ID) error {
		if err := node.orderbook.Cancel(order.Order{
			ID: orderID,
		}); err != nil {
			return err
		}
		node.orderFragmentsCanceled <- orderID
		return nil
	})

	return node, nil
}

// Run is the recommended way to turn on a Darknode.
func (node *Darknode) Run(done <-chan struct{}) <-chan error {
	errs := make(chan error, 1)

	go func() {
		defer close(errs)

		// Wait until registration is approved
		node.Logger.Info("waiting for registration...")
		if err := node.darknodeRegistry.WaitUntilRegistration(node.ID()[:]); err != nil {
			errs <- err
			return
		}

		// Start serving
		go func() {

			node.Logger.Info("serving gRPC services...")
			if err := node.Serve(done); err != nil {
				errs <- err
				return
			}
		}()
		time.Sleep(time.Second)

		// Bootstrap into the network and stop after all search paths are
		// exhausted, or one minute has passed
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		dispatch.Pipe(done, node.Bootstrap(ctx), errs)

		// Run epochs
		dispatch.Pipe(done, node.RunEpochs(done), errs)
	}()

	return errs
}

// Bootstrap the Darknode into the swarm network. The Darknode will query all
// reachable nodes for itself, updating its dht.DHT as it connects to other
// nodes. Calls to Darknode.Bootstrap are not blocking, and return a channel of
// errors encountered. Users should not call Darknode.Bootstrap until the
// Darknode is registered, and the its registration is approved.
func (node *Darknode) Bootstrap(ctx context.Context) <-chan error {
	return node.rpc.SwarmerClient().Bootstrap(ctx, node.Config.BootstrapMultiAddresses, -1)
}

// Serve the Darknode services until the done channel is closed.
func (node *Darknode) Serve(done <-chan struct{}) error {
	server := grpc.NewServer()
	listener, err := net.Listen("tcp", fmt.Sprintf("%v:%v", node.Config.Host, node.Config.Port))
	if err != nil {
		return err
	}
	node.rpc.Relayer().Register(server)
	node.rpc.Smpcer().Register(server)
	node.rpc.Swarmer().Register(server)
	status.RegisterStatusServer(server, node.rpc)

	go func() {
		node.Logger.Info("listening at " + node.Config.Host + " " + node.Config.Port)
		if err = server.Serve(listener); err != nil {
			return
		}
	}()
	<-done
	server.Stop()
	return err
}

// RunEpochs will watch for changes to the Ocean and run the secure
// multi-party computation with new Pools. Stops when the done channel is
// closed, and will attempt to recover from errors encountered while
// interacting with the Ocean.
func (node *Darknode) RunEpochs(done <-chan struct{}) <-chan error {
	errs := make(chan error, 1)

	go func() {
		// Maintain multiple done channels so that multiple epochs can be running
		// in parallel
		var prevDone chan struct{}
		var currDone chan struct{}
		defer func() {
			if prevDone != nil {
				close(prevDone)
			}
			if currDone != nil {
				close(currDone)
			}
		}()

		// Looping until the done channel is closed will recover from errors
		// returned by watching the Ocean
		for {
			select {
			case <-done:
				return
			default:
			}

			// Start watching epochs
			epochs, epochErrs := RunEpochWatcher(done, node.darknodeRegistry)
			go dispatch.Pipe(done, epochErrs, errs)

			for quit := false; !quit; {
				select {

				case <-done:
					return

				case err, ok := <-errs:
					if !ok {
						quit = true
						break
					}
					node.Logger.Network(logger.Error, err.Error())

				case epoch, ok := <-epochs:
					if !ok {
						quit = true
						break
					}
					if prevDone != nil {
						close(prevDone)
					}
					prevDone = currDone
					currDone = make(chan struct{})

					darknodeIDs, err := node.darknodeRegistry.GetAllNodes()
					if err != nil {
						// FIXME: Do not skip the epoch. Retry with a backoff.
						errs <- err
						continue
					}

					darkOcean, err := darkocean.NewDarkOcean(&node.darknodeRegistry, epoch.Blockhash, darknodeIDs)
					if err != nil {
						// FIXME: Do not skip the epoch. Retry with a backoff.
						errs <- err
						continue
					}

					deltas, deltaErrs := node.RunEpochProcess(currDone, darkOcean)
					go dispatch.Pipe(done, deltaErrs, errs)
					go func() {
						for dlt := range deltas {
							if dlt.IsMatch(smpc.Prime) {
								node.Logger.OrderMatch(logger.Info, dlt.ID.String(), dlt.BuyOrderID.String(), dlt.SellOrderID.String())
							}
						}
					}()
				}
			}
		}
	}()

	return errs
}

// ID returns the ID of the Darknode.
func (node *Darknode) ID() identity.ID {
	return node.id
}

// Address returns the Address of the Darknode.
func (node *Darknode) Address() identity.Address {
	return node.address
}

// MultiAddress returns the MultiAddress of the Darknode.
func (node *Darknode) MultiAddress() identity.MultiAddress {
	return node.multiAddress
}

// OnOpenOrder implements the rpc.RelayDelegate interface.
func (node *Darknode) OnOpenOrder(from identity.MultiAddress, orderFragment *order.Fragment) {
	node.orderFragments <- *orderFragment
	ord := order.Order{
		ID: orderFragment.OrderID,
	}
	err := node.orderbook.Open(ord)
	if err != nil {
		node.Logger.Compute(logger.Error, err.Error())
	}
}

// OnReleaseOrder re-opens an order that was previously in a different state.
func (node *Darknode) OnReleaseOrder(orderID order.ID) {
	if err := node.orderbook.Release(order.Order{
		ID: orderID,
	}); err != nil {
		node.Logger.Compute(logger.Error, err.Error())
	}
}
