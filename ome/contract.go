package ome

import (
	"errors"

	"github.com/republicprotocol/republic-go/order"
)

// ErrOrderNotFound is returned when attempting to load an order that cannot be
// found.
var ErrOrderNotFound = errors.New("order not found")

// ContractBinder will define all methods that the order matching
// engine will require to communicate with smart contracts. All the
// methods will be implemented in contract.Binder
type ContractBinder interface {
	ConfirmOrder(buy order.ID, sell order.ID) error

	Depth(orderID order.ID) (uint, error)

	Status(orderID order.ID) (order.Status, error)

	OrderMatch(order order.ID) (order.ID, error)

	Settle(buy order.Order, sell order.Order) error

	SubmitChallengeOrder(ord order.Order) error

	SubmitChallenge(buyID, sellID order.ID) error
}
