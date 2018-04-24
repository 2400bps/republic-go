package hyperdrive

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/republicprotocol/republic-go/identity"
)

// Txs must not store any Nonce more than once within any Tx.
type Txs []Tx

// A Tx stores Nonces alongside a Keccak256 Hash of the Nonces. A valid Tx must
// not store any Nonce more than once.
type Tx struct {
	identity.Hash
	Nonces
}

func NewTx(nonces ...Nonce) Tx {
	sliceNonces := make([][]byte, len(nonces))
	for i := range sliceNonces {
		sliceNonces[i] = nonces[i][:]
	}
	var hash [32]byte
	copy(hash[:], crypto.Keccak256(sliceNonces...))
	return Tx{
		Hash:   hash,
		Nonces: nonces,
	}
}

func NewTxFromByteSlices(nonces ...[]byte) Tx {
	var hash [32]byte
	copy(hash[:], crypto.Keccak256(nonces...))

	noncesArray := make([]Nonce, len(nonces))
	for i := range nonces {
		copy(noncesArray[i][:], nonces[i])
	}
	return Tx{
		Hash:   hash,
		Nonces: noncesArray,
	}
}

type TxWithBlockNumber struct {
	Hash        common.Hash
	BlockNumber uint64
}

func NewTxWithBlockNumber(hash common.Hash, blockNumber uint64) TxWithBlockNumber {
	return TxWithBlockNumber{
		Hash:        hash,
		BlockNumber: blockNumber,
	}
}

// Nonces must not store any Nonce more than once.
type Nonces []Nonce

// A Nonce is a unique 256-bit value that makes up a Tx. It must be unique
// within the entire Hyperdrive blockchain.
type Nonce = [32]byte
