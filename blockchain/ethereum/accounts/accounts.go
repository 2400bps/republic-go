package accounts

import (
	"context"
	"encoding/base64"
	"log"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/republicprotocol/republic-go/blockchain/ethereum"
	"github.com/republicprotocol/republic-go/order"
)

// RenExAccounts implements the cal.DarkpoolAccounts interface
type RenExAccounts struct {
	mu *sync.RWMutex

	network      ethereum.Network
	context      context.Context
	conn         ethereum.Conn
	transactOpts *bind.TransactOpts
	callOpts     *bind.CallOpts
	binding      *RenExSettlement
	address      common.Address
}

// NewRenExAccounts returns a new RenExAccounts
func NewRenExAccounts(ctx context.Context, conn ethereum.Conn, transactOpts *bind.TransactOpts, callOpts *bind.CallOpts) (RenExAccounts, error) {
	contract, err := NewRenExSettlement(common.HexToAddress(conn.Config.RenExAccountsAddress), bind.ContractBackend(conn.Client))
	if err != nil {
		return RenExAccounts{}, err
	}

	return RenExAccounts{
		mu: new(sync.RWMutex),

		network:      conn.Config.Network,
		context:      ctx,
		conn:         conn,
		transactOpts: transactOpts,
		callOpts:     callOpts,
		binding:      contract,
		address:      common.HexToAddress(conn.Config.RenExAccountsAddress),
	}, nil
}

// SubmitOrder to the RenEx accounts
func (accounts *RenExAccounts) SubmitOrder(ord order.Order) error {
	accounts.mu.Lock()
	defer accounts.mu.Unlock()

	return accounts.submitOrder(ord)
}

// SubmitMatch will submit a matched order pair to the RenEx accounts
func (accounts *RenExAccounts) SubmitMatch(buy, sell order.ID) error {
	accounts.mu.Lock()
	defer accounts.mu.Unlock()

	return accounts.submitMatch(buy, sell)
}

// Settle implements the cal.DarkpoolAccounts interface
func (accounts *RenExAccounts) Settle(buy order.Order, sell order.Order) error {
	accounts.mu.Lock()
	defer accounts.mu.Unlock()

	err := accounts.submitOrder(buy)
	if err != nil {
		return err
	}
	err = accounts.submitOrder(sell)
	if err != nil {
		return err
	}

	return accounts.submitMatch(buy.ID, sell.ID)
}

func (accounts *RenExAccounts) SettlementDetail(buy, sell order.ID) (*big.Int, *big.Int, *big.Int, *big.Int, *big.Int, error) {
	accounts.mu.RLock()
	defer accounts.mu.RUnlock()

	price, lowVolume, highVolume, lowFee, highFee, err := accounts.binding.GetSettlementDetails(accounts.callOpts, buy, sell)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	return price, lowVolume, highVolume, lowFee, highFee, nil
}

func (accounts *RenExAccounts) submitMatch(buy, sell order.ID) error {
	accounts.transactOpts.GasLimit = 500000
	tx, err := accounts.binding.SubmitMatch(accounts.transactOpts, buy, sell)
	if err != nil {
		return err
	}
	_, err = accounts.conn.PatchedWaitMined(accounts.context, tx)
	return err
}

func (accounts *RenExAccounts) submitOrder(ord order.Order) error {
	nonceHash := big.NewInt(0).SetBytes(ord.BytesFromNonce())
	log.Printf("[submit order] id: %v,tokens:%d, priceCo:%v, priceExp:%v, volumeCo:%v, volumeExp:%v, minVol:%v, minVolExp:%v", base64.StdEncoding.EncodeToString(ord.ID[:]), uint64(ord.Tokens), uint16(ord.Price.Co), uint16(ord.Price.Exp), uint16(ord.Volume.Co), uint16(ord.Volume.Exp), uint16(ord.MinimumVolume.Co), uint16(ord.MinimumVolume.Exp))
	tx, err := accounts.binding.SubmitOrder(accounts.transactOpts, uint8(ord.Type), uint8(ord.Parity), uint64(ord.Expiry.Unix()), uint64(ord.Tokens), uint16(ord.Price.Co), uint16(ord.Price.Exp), uint16(ord.Volume.Co), uint16(ord.Volume.Exp), uint16(ord.MinimumVolume.Co), uint16(ord.MinimumVolume.Exp), nonceHash)
	if err != nil {
		return err
	}
	_, err = accounts.conn.PatchedWaitMined(accounts.context, tx)
	return err
}
