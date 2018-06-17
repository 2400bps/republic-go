package ome

import (
	"errors"
	"sync"
	"time"

	"github.com/republicprotocol/republic-go/cal"
	"github.com/republicprotocol/republic-go/order"
)

// ErrOrderNotConfirmed is an internal error that is returned when an
// order.Order is not confirmed.
var ErrOrderNotConfirmed = errors.New("order not confirmed")

// A Confirmer consumes Computations that been resolved to a match, usually by
// a call to Matcher.Resolve, and reaches consensus with other Darknodes that
// this match has happened. This prevents the occurrence of conflicting matches
// across the parallel resolving done by different Darknode pods.
type Confirmer interface {

	// Confirm Computations that have resolved in a match by reaching consensus
	// with other Darknodes. The input channel will be consumed and an output
	// channel of confirmed Computations is produced. Stop the Confirmer by
	// closing the done channel.
	Confirm(done <-chan struct{}, coms <-chan Computation) (<-chan Computation, <-chan error)
}

type confirmer struct {
	storer Storer

	renLedger             cal.RenLedger
	renLedgerPollInterval time.Duration
	renLedgerBlockDepth   uint

	confirmingMu         *sync.Mutex
	confirmingBuyOrders  map[order.ID]struct{}
	confirmingSellOrders map[order.ID]struct{}
}

// NewConfirmer returns a Confirmer that submits Computations to the
// cal.RenLedger for confirmation. It polls the cal.RenLedger on an interval
// and checks for consensus on confirmations by waiting until a submitted
// Computation has been confirmed has the confirmation has passed the block
// depth limit.
func NewConfirmer(storer Storer, renLedger cal.RenLedger, renLedgerPollInterval time.Duration, renLedgerBlockDepth uint) Confirmer {
	return &confirmer{
		storer: storer,

		renLedger:             renLedger,
		renLedgerPollInterval: renLedgerPollInterval,
		renLedgerBlockDepth:   renLedgerBlockDepth,

		confirmingMu:         new(sync.Mutex),
		confirmingBuyOrders:  map[order.ID]struct{}{},
		confirmingSellOrders: map[order.ID]struct{}{},
	}
}

// Confirm implements the Confirmer interface.
func (confirmer *confirmer) Confirm(done <-chan struct{}, coms <-chan Computation) (<-chan Computation, <-chan error) {
	confirmations := make(chan Computation, 64)
	errs := make(chan error, 1)

	// Two background processes are run that must end before cleanup can
	// happen safely
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for {
			select {
			case <-done:
				return
			case com, ok := <-coms:
				if !ok {
					return
				}
				// Confirm Computations on the blockchain and register them for
				// observation (we need to wait for finality)
				if err := confirmer.beginConfirmation(com); err != nil {
					select {
					case <-done:
						return
					case errs <- err:
						continue
					}
				}

				// Wait for the confirmation of these orders to pass the depth
				// limit
				confirmer.confirmingMu.Lock()
				confirmer.confirmingBuyOrders[com.Buy] = struct{}{}
				confirmer.confirmingSellOrders[com.Sell] = struct{}{}
				confirmer.confirmingMu.Unlock()
			}
		}
	}()

	// Periodically poll the cal.RenLedger to observe the state of
	// confirmations that have passed the block depth limit
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(confirmer.renLedgerPollInterval)
		defer ticker.Stop()

		for {
			select {

			// Graceful shutdown when the done channel is closed
			case <-done:
				return

			case <-ticker.C:
				confirmer.confirmingMu.Lock()
				confirmer.checkOrdersForConfirmationFinality(order.ParityBuy, done, confirmations, errs)
				confirmer.checkOrdersForConfirmationFinality(order.ParitySell, done, confirmations, errs)
				confirmer.confirmingMu.Unlock()
			}
		}
	}()

	// Cleanup
	go func() {
		defer close(confirmations)
		defer close(errs)
		wg.Wait()
	}()

	return confirmations, errs
}

func (confirmer *confirmer) beginConfirmation(orderMatch Computation) error {
	return confirmer.renLedger.ConfirmOrder(orderMatch.Buy, orderMatch.Sell)
}

func (confirmer *confirmer) checkOrdersForConfirmationFinality(orderParity order.Parity, done <-chan struct{}, confirmations chan<- Computation, errs chan<- error) {
	var confirmingOrders map[order.ID]struct{}
	if orderParity == order.ParityBuy {
		confirmingOrders = confirmer.confirmingBuyOrders
	} else {
		confirmingOrders = confirmer.confirmingSellOrders
	}
	for ord := range confirmingOrders {
		ordMatch, err := confirmer.checkOrderForConfirmationFinality(ord, orderParity)
		if err != nil {
			if err == ErrOrderNotConfirmed {
				continue
			}
			select {
			case <-done:
				return
			case errs <- err:
				continue
			}
		}

		com, err := confirmer.computationFromOrders(orderParity, ord, ordMatch)
		if err != nil {
			select {
			case <-done:
				return
			case errs <- err:
				continue
			}
		}
		if err := confirmer.storer.InsertComputation(com); err != nil {
			select {
			case <-done:
				return
			case errs <- err:
			}
		}

		select {
		case <-done:
			return
		case confirmations <- com:
			delete(confirmer.confirmingBuyOrders, com.Buy)
			delete(confirmer.confirmingSellOrders, com.Sell)
		}
	}
}

func (confirmer *confirmer) checkOrderForConfirmationFinality(ord order.ID, orderParity order.Parity) (order.ID, error) {
	// Ignore orders that are not pass the depth limit
	depth, err := confirmer.renLedger.Depth(ord)
	if err != nil {
		return order.ID{}, err
	}
	if depth < confirmer.renLedgerBlockDepth {
		return order.ID{}, ErrOrderNotConfirmed
	}

	// Purge orders that are not confirmed
	status, err := confirmer.renLedger.Status(ord)
	if err != nil {
		return order.ID{}, err
	}
	if status != cal.StatusConfirmed {
		if orderParity == order.ParityBuy {
			delete(confirmer.confirmingBuyOrders, ord)
		} else {
			delete(confirmer.confirmingSellOrders, ord)
		}
		return order.ID{}, ErrOrderNotConfirmed
	}

	match, err := confirmer.renLedger.OrderMatch(ord)
	if err != nil {
		return order.ID{}, err
	}
	return match, nil
}

func (confirmer *confirmer) computationFromOrders(orderParity order.Parity, ord, ordMatch order.ID) (Computation, error) {
	var comID ComputationID
	if orderParity == order.ParityBuy {
		comID = NewComputationID(ord, ordMatch)
	} else {
		comID = NewComputationID(ordMatch, ord)
	}
	com, err := confirmer.storer.Computation(comID)
	if err != nil {
		return com, err
	}
	com.State = ComputationStateAccepted
	com.Timestamp = time.Now()
	return com, nil
}
