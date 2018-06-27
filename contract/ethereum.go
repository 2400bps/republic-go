package contract

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
)

// Conn contains the client and the contracts deployed to it
type Conn struct {
	RawClient *ethrpc.Client
	Client    *ethclient.Client
	Config    Config
}

// Connect to a URI.
func Connect(config Config) (Conn, error) {
	if config.URI == "" {
		switch config.Network {
		case NetworkTestnet:
			config.URI = "https://kovan.infura.io"
		case NetworkFalcon:
			config.URI = "https://kovan.infura.io"
		case NetworkNightly:
			config.URI = "https://kovan.infura.io"
		case NetworkLocal:
			config.URI = "http://localhost:8545"
		default:
			return Conn{}, fmt.Errorf("cannot connect to %s: unsupported", config.Network)
		}
	}

	if config.RepublicTokenAddress == "" {
		switch config.Network {
		case NetworkTestnet:
			config.RepublicTokenAddress = "0x6f429121a3bd3e6c1c17edbc676eec44cf117faf"
		case NetworkFalcon:
			config.RepublicTokenAddress = "0x87e83f957a2f3a2e5fe16d5c6b22e38fd28bdc06"
		case NetworkNightly:
			config.RepublicTokenAddress = "0x15f692d6b9ba8cec643c7d16909e8acdec431bf6"
		case NetworkLocal:
		default:
			return Conn{}, fmt.Errorf("no default contract address on %s", config.Network)
		}
	}
	if config.DarknodeRegistryAddress == "" {
		switch config.Network {
		case NetworkTestnet:
			config.DarknodeRegistryAddress = "0x5d09eb34ce084bece690651f147dbd8ff41007bf"
		case NetworkFalcon:
			config.DarknodeRegistryAddress = "0x7352e7244899b7cb5d803cc02741c8910d3b75de"
		case NetworkNightly:
			config.DarknodeRegistryAddress = "0xc735241f93f87d4dbea499ee6e1d41ec50e3d8ce"
		case NetworkLocal:
		default:
			return Conn{}, fmt.Errorf("no default contract address on %s", config.Network)
		}
	}
	if config.OrderbookAddress == "" {
		switch config.Network {
		case NetworkTestnet:
			config.OrderbookAddress = "0xB01219Cf49e92ffcd48fecC96241dBd1372B8Bb1"
		case NetworkFalcon:
			config.OrderbookAddress = "0x20949251119d77471a40f456a8a9d39b1847db8d"
		case NetworkNightly:
			config.OrderbookAddress = "0x42c72b4090ed0627c85ed878f699b2db254beeca"
		case NetworkLocal:
		default:
			return Conn{}, fmt.Errorf("no default contract address on %s", config.Network)
		}
	}
	if config.RewardVaultAddress == "" {
		switch config.Network {
		case NetworkTestnet:
			config.RewardVaultAddress = "0x5d62ccc1086f38286dc152962a4f3e337eec1ec1"
		case NetworkFalcon:
			config.RewardVaultAddress = "0x0e6bbbb35835cc3624a000e1698b7b68e9eec7df"
		case NetworkNightly:
			config.RewardVaultAddress = "0x65129f15fc0bfd901ce99c71147a93256fa094e6"
		case NetworkLocal:
		default:
			return Conn{}, fmt.Errorf("no default contract address on %s", config.Network)
		}
	}
	if config.RenExBalancesAddress == "" {
		switch config.Network {
		case NetworkTestnet:
			config.RenExBalancesAddress = "0xc5b98949AB0dfa0A7d4c07Bb29B002D6d6DA3e25"
		case NetworkFalcon:
			config.RenExBalancesAddress = "0x3083e5ba36c6b42ca93c22c803013a4539eedc7f"
		case NetworkNightly:
			config.RenExBalancesAddress = "0x6268002a734edcde6c2111ae339e0d92b1ed2bfa"
		case NetworkLocal:
		default:
			return Conn{}, fmt.Errorf("no default contract address on %s", config.Network)
		}
	}
	if config.RenExSettlementAddress == "" {
		switch config.Network {
		case NetworkTestnet:
			config.RenExSettlementAddress = "0xc53abbc5713e606a86533088707e80fcae33eff8"
		case NetworkFalcon:
			config.RenExSettlementAddress = "0x038b63c120a7e60946d6ebaa6dcfc3a475108cc9"
		case NetworkNightly:
			config.RenExSettlementAddress = "0xd42f9dd5e66627aa836b206edd76025b26a89dea"
		case NetworkLocal:
		default:
			return Conn{}, fmt.Errorf("no default contract address on %s", config.Network)
		}
	}

	ethclient, err := ethclient.Dial(config.URI)
	if err != nil {
		return Conn{}, err
	}

	return Conn{
		Client: ethclient,
		Config: config,
	}, nil
}

// PatchedWaitMined waits for tx to be mined on the blockchain.
// It stops waiting when the context is canceled.
//
// TODO: THIS DOES NOT WORK WITH PARITY, WHICH SENDS A TRANSACTION RECEIPT UPON
// RECEIVING A TX, NOT AFTER IT'S MINED
func (conn *Conn) PatchedWaitMined(ctx context.Context, tx *types.Transaction) (*types.Receipt, error) {
	switch conn.Config.Network {
	case NetworkLocal:
		time.Sleep(100 * time.Millisecond)
		return nil, nil
	default:
		return bind.WaitMined(ctx, conn.Client, tx)
	}
}

// PatchedWaitDeployed waits for a contract deployment transaction and returns the on-chain
// contract address when it is mined. It stops waiting when ctx is canceled.
//
// TODO: THIS DOES NOT WORK WITH PARITY, WHICH SENDS A TRANSACTION RECEIPT UPON
// RECEIVING A TX, NOT AFTER IT'S MINED
func (conn *Conn) PatchedWaitDeployed(ctx context.Context, tx *types.Transaction) (common.Address, error) {
	switch conn.Config.Network {
	case NetworkLocal:
		time.Sleep(100 * time.Millisecond)
		return common.Address{}, nil
	default:
		return bind.WaitDeployed(ctx, conn.Client, tx)
	}
}

// TransferEth is a helper function for sending ETH to an address
func (conn *Conn) TransferEth(ctx context.Context, from *bind.TransactOpts, to common.Address, value *big.Int) error {
	transactor := &bind.TransactOpts{
		From:     from.From,
		Nonce:    from.Nonce,
		Signer:   from.Signer,
		Value:    value,
		GasPrice: from.GasPrice,
		GasLimit: 30000,
		Context:  from.Context,
	}

	// Why is there no ethclient.Transfer?
	bound := bind.NewBoundContract(to, abi.ABI{}, nil, conn.Client, nil)
	tx, err := bound.Transfer(transactor)
	if err != nil {
		return err
	}
	_, err = conn.PatchedWaitMined(ctx, tx)
	return err
}
