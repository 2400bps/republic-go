package contract

import (
	"context"
	"errors"
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
			config.DarknodeRegistryAddress = "0x372b6204263c6867f81e2a9e11057ff43efea14b"
		case NetworkFalcon:
			config.DarknodeRegistryAddress = "0xfafd5c83d1e21763b79418c4ecb5d62b4970df8e"
		case NetworkNightly:
			config.DarknodeRegistryAddress = "0x8a31d477267a5af1bc5142904ef0afa31d326e03"
		case NetworkLocal:
		default:
			return Conn{}, fmt.Errorf("no default contract address on %s", config.Network)
		}
	}
	if config.DarknodeRewardVaultAddress == "" {
		switch config.Network {
		case NetworkTestnet:
			config.DarknodeRewardVaultAddress = "0x5d62ccc1086f38286dc152962a4f3e337eec1ec1"
		case NetworkFalcon:
			config.DarknodeRewardVaultAddress = "0x0e6bbbb35835cc3624a000e1698b7b68e9eec7df"
		case NetworkNightly:
			config.DarknodeRewardVaultAddress = "0xda43560f5fe6c6b5e062c06fee0f6fbc71bbf18a"
		case NetworkLocal:
		default:
			return Conn{}, fmt.Errorf("no default contract address on %s", config.Network)
		}
	}
	if config.DarknodeSlasherAddress == "" {
		switch config.Network {
		case NetworkTestnet:
			config.DarknodeSlasherAddress = "0x38458ef4a185455cba57a7594b0143c53ad057c1"
		case NetworkFalcon:
			config.DarknodeSlasherAddress = "0x38458ef4a185455cba57a7594b0143c53ad057c1"
		case NetworkNightly:
			config.DarknodeSlasherAddress = "0x38458ef4a185455cba57a7594b0143c53ad057c1"
		case NetworkLocal:
		default:
			return Conn{}, fmt.Errorf("no default contract address on %s", config.Network)
		}
	}
	if config.OrderbookAddress == "" {
		switch config.Network {
		case NetworkTestnet:
			config.OrderbookAddress = "0xa7caa4780a39d8b8acd6a0bdfb5b906210bc76cd"
		case NetworkFalcon:
			config.OrderbookAddress = "0x044b08eec761c39ac32aee1d6ef0583812f21699"
		case NetworkNightly:
			config.OrderbookAddress = "0x376127adc18260fc238ebfb6626b2f4b59ec9b66"
		case NetworkLocal:
		default:
			return Conn{}, fmt.Errorf("no default contract address on %s", config.Network)
		}
	}
	if config.SettlementRegistryAddress == "" {
		switch config.Network {
		case NetworkTestnet:
			config.SettlementRegistryAddress = "0x399a70ed71897836468fd74ea19138df90a78d79"
		case NetworkFalcon:
			config.SettlementRegistryAddress = "0x399a70ed71897836468fd74ea19138df90a78d79"
		case NetworkNightly:
			config.SettlementRegistryAddress = "0x399a70ed71897836468fd74ea19138df90a78d79"
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
		receipt, err := bind.WaitMined(ctx, conn.Client, tx)
		if err != nil {
			return nil, err
		}
		if receipt.Status != types.ReceiptStatusSuccessful {
			return receipt, errors.New("transaction reverted")
		}
		return receipt, nil
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
		GasLimit: 21000,
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

// SendEth is a helper function for sending ETH to an address
func (conn *Conn) SendEth(ctx context.Context, from *bind.TransactOpts, to common.Address, value *big.Int) (*types.Transaction, error) {
	transactor := &bind.TransactOpts{
		From:     from.From,
		Nonce:    from.Nonce,
		Signer:   from.Signer,
		Value:    value,
		GasPrice: from.GasPrice,
		GasLimit: 21000,
		Context:  from.Context,
	}

	// Why is there no ethclient.Transfer?
	bound := bind.NewBoundContract(to, abi.ABI{}, nil, conn.Client, nil)
	return bound.Transfer(transactor)
}

// TokenAddresses returns the tokens for the provided network
func TokenAddresses(network Network) map[string]string {
	tokens := map[string]string{}
	switch network {
	case NetworkTestnet:
		tokens["ABC"] = "0x289f785d9137ecf38a46a678cf4e9e98d32a06d4"
		tokens["DGX"] = "0x0798297a11cefef7479e40e67839fee3c025691e"
		tokens["REN"] = "0x6f429121a3bd3e6c1c17edbc676eec44cf117faf"
		tokens["PQR"] = "0x099ea44e49e34250e247a150c66c89b314216e34"
		tokens["UVW"] = "0x58bc110f70291e5c731af8bf99cd8d209c0dfd3e"
		tokens["XYZ"] = "0x0f48986df7b79fbb085753dc2fefe10dde7dd232"
	case NetworkFalcon:
		tokens["ABC"] = "0x1c428ab82c06dbe9af414e6c923862d4c3ae0579"
		tokens["DGX"] = "0xf4faf1b22cee0a024ad6b12bb29ec0e13f5827c2"
		tokens["REN"] = "0x87e83f957a2f3a2e5fe16d5c6b22e38fd28bdc06"
		tokens["PQR"] = "0x295a3894fc98b021735a760dbc7aed265663ca42"
		tokens["UVW"] = "0x011c45eaa4cf4ad49978887e02f944434a5033b9"
		tokens["XYZ"] = "0x21c1ba3ea123eb23815c689ee05a944119c7f428"
	case NetworkNightly:
		tokens["ABC"] = "0xa86c6a3322efa371faad6a9b04708788e3592615"
		tokens["DGX"] = "0x092ece29781777604afac04887af30042c3bc5df"
		tokens["REN"] = "0x15f692d6b9ba8cec643c7d16909e8acdec431bf6"
		tokens["PQR"] = "0xeb5a7335e850176b44ca1990730d1a2433e195f3"
		tokens["UVW"] = "0x7dd5f16f2e0a0030e9512e0f888443c4408dffb0"
		tokens["XYZ"] = "0x69440b57b52e323cbd12a162a5f9870f61182918"
	default:
		panic("unknown network")
	}
	return tokens
}
