package order

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"math/big"
	"os"
	"time"

	"github.com/republicprotocol/republic-go/crypto"
	"github.com/republicprotocol/republic-go/shamir"
)

// A Signature is the ECDSA signature of an order ID.
type Signature [65]byte

// An ID is the keccak256 hash of an Order.
type ID [32]byte

// Equal returns an equality check between two DeltaFragmentIDs.
func (id ID) Equal(other ID) bool {
	return bytes.Equal(id[:], other[:])
}

// String returns a truncated base64 encoding of the ID.
func (id ID) String() string {
	runes := []rune(base64.StdEncoding.EncodeToString(id[:]))
	return string(runes[:4])
}

// Token is a numerical representation of a token supported by Republic
// Protocol.
type Token uint32

// Token values.
const (
	TokenBTC Token = 0
	TokenETH Token = 1
	TokenDGX Token = 256
	TokenABC Token = 257
	TokenREN Token = 65536
	TokenPQR Token = 65537
	TokenXYZ Token = 65538
)

// String returns a human-readable representation of a Token.
func (token Token) String() string {
	switch token {
	case TokenBTC:
		return "BTC"
	case TokenETH:
		return "ETH"
	case TokenDGX:
		return "DGX"
	case TokenABC:
		return "ABC"
	case TokenREN:
		return "REN"
	case TokenPQR:
		return "PQR"
	case TokenXYZ:
		return "XYZ"
	default:
		return "unexpected token"
	}
}

// Tokens are a numerical representation of the token pairings supported by
// Republic Protocol.
type Tokens uint64

// Tokens values.
const (
	TokensBTCETH = Tokens((uint64(TokenBTC) << 32) | uint64(TokenETH))
	TokensETHDGX = Tokens((uint64(TokenETH) << 32) | uint64(TokenDGX))
	TokensETHABC = Tokens((uint64(TokenETH) << 32) | uint64(TokenABC))
	TokensETHREN = Tokens((uint64(TokenETH) << 32) | uint64(TokenREN))
	TokensETHPQR = Tokens((uint64(TokenETH) << 32) | uint64(TokenPQR))
	TokensETHXYZ = Tokens((uint64(TokenETH) << 32) | uint64(TokenXYZ))
)

// PriorityToken returns the priority token of a token pair.
func (tokens Tokens) PriorityToken() Token {
	return Token(tokens & 0x00000000FFFFFFFF)
}

// NonPriorityToken returns the non-priority token of a token pair.
func (tokens Tokens) NonPriorityToken() Token {
	return Token(tokens >> 32)
}

// String returns a human-readable representation of Tokens.
func (tokens Tokens) String() string {
	switch tokens {
	case TokensBTCETH:
		return "BTC-ETH"
	case TokensETHDGX:
		return "ETH-DGX"
	case TokensETHABC:
		return "ETH-ABC"
	case TokensETHREN:
		return "ETH-REN"
	case TokensETHPQR:
		return "ETH-PQR"
	case TokensETHXYZ:
		return "ETH-XYZ"
	default:
		return "unexpected tokens"
	}
}

// A Type is a publicly bit of information that determines the type of
// trade that an Order is representing.
type Type int8

// Type values.
const (
	TypeMidpoint    Type = 0
	TypeLimit       Type = 1
	TypeMidpointFOK Type = 2
	TypeLimitFOK    Type = 3
)

// The Parity of an Order determines whether it is buy or a sell.
type Parity int8

// Parity values.
const (
	ParityBuy  Parity = 0
	ParitySell Parity = 1
)

// String returns a human-readable representation of the Parity.
func (parity Parity) String() string {
	switch parity {
	case ParityBuy:
		return "buy"
	case ParitySell:
		return "sell"
	default:
		return "unexpected parity"
	}
}

// Settlement is a unique identifier for the settlement layer used by the
// Order.
type Settlement uint64

// Values for Settlementt.
const (
	SettlementNil (Settlement) = iota
	SettlementRenEx
	SettlementRenExAtomic
)

// String implements the Stringer interface.
func (settlement Settlement) String() string {
	switch settlement {
	case SettlementRenEx:
		return "RenEx"
	case SettlementRenExAtomic:
		return "RenEx Atomic"
	default:
		return "unexpected order settlement"
	}
}

// The Status shows what status the order is in.
type Status uint8

// Status values.
const (
	Nil = Status(iota)
	Open
	Confirmed
	Canceled
)

// String implements the Stringer interface.
func (status Status) String() string {
	switch status {
	case Nil:
		return "nil"
	case Open:
		return "open"
	case Confirmed:
		return "confirmed"
	case Canceled:
		return "canceled"
	default:
		return "unexpected order status"
	}
}

// An Order represents the want to perform a trade of assets.
type Order struct {
	ID ID `json:"id"`

	Parity Parity    `json:"parity"`
	Type   Type      `json:"type"`
	Expiry time.Time `json:"expiry"`
	Nonce  uint64    `json:"nonce"`

	Settlement    Settlement `json:"settlement"`
	Tokens        Tokens     `json:"tokens"`
	Price         uint64     `json:"price"`
	Volume        uint64     `json:"volume"`
	MinimumVolume uint64     `json:"minimumVolume"`
}

// NewOrder returns a new Order and computes the ID.
func NewOrder(parity Parity, ty Type, expiry time.Time, settlement Settlement, tokens Tokens, price, volume, minimumVolume, nonce uint64) Order {
	priceCoExp := PriceToCoExp(price)
	volumeCoExp := VolumeToCoExp(volume)
	minimumVolumeCoExp := VolumeToCoExp(minimumVolume)
	order := Order{
		Parity: parity,
		Type:   ty,
		Expiry: expiry,
		Nonce:  nonce,

		Settlement:    settlement,
		Tokens:        tokens,
		Price:         PriceFromCoExp(priceCoExp.Co, priceCoExp.Exp),
		Volume:        VolumeFromCoExp(volumeCoExp.Co, volumeCoExp.Exp),
		MinimumVolume: VolumeFromCoExp(minimumVolumeCoExp.Co, minimumVolumeCoExp.Exp),
	}
	order.ID = ID(order.Hash())
	return order
}

// NewOrderFromJSONFile returns an order that is unmarshaled from a JSON file.
func NewOrderFromJSONFile(fileName string) (Order, error) {
	order := Order{}
	file, err := os.Open(fileName)
	if err != nil {
		return order, err
	}
	defer file.Close()
	if err := json.NewDecoder(file).Decode(&order); err != nil {
		return order, err
	}
	return order, nil
}

// NewOrdersFromJSONFile returns an array of orders that is unmarshaled from a JSON file.
func NewOrdersFromJSONFile(fileName string) ([]Order, error) {
	orders := []Order{}
	file, err := os.Open(fileName)
	if err != nil {
		return orders, err
	}
	defer file.Close()
	if err := json.NewDecoder(file).Decode(&orders); err != nil {
		return orders, err
	}
	return orders, nil
}

// WriteOrdersToJSONFile writes an array of orders into a JSON file.
func WriteOrdersToJSONFile(fileName string, orders []*Order) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(&orders)
}

// Split the Order into n OrderFragments, where k OrderFragments are needed to
// reconstruct the Order. Returns a slice of all n OrderFragments, or an error.
func (order *Order) Split(n, k int64) ([]Fragment, error) {
	priceCoExp := PriceToCoExp(order.Price)
	volumeCoExp := VolumeToCoExp(order.Volume)
	minimumVolumeCoExp := VolumeToCoExp(order.MinimumVolume)

	tokens, err := shamir.Split(n, k, uint64(order.Tokens))
	if err != nil {
		return nil, err
	}
	priceCos, err := shamir.Split(n, k, priceCoExp.Co)
	if err != nil {
		return nil, err
	}
	priceExps, err := shamir.Split(n, k, priceCoExp.Exp)
	if err != nil {
		return nil, err
	}
	volumeCos, err := shamir.Split(n, k, volumeCoExp.Co)
	if err != nil {
		return nil, err
	}
	volumeExps, err := shamir.Split(n, k, volumeCoExp.Exp)
	if err != nil {
		return nil, err
	}
	minimumVolumeCos, err := shamir.Split(n, k, minimumVolumeCoExp.Co)
	if err != nil {
		return nil, err
	}
	minimumVolumeExps, err := shamir.Split(n, k, minimumVolumeCoExp.Exp)
	if err != nil {
		return nil, err
	}
	nonces, err := shamir.Split(n, k, order.Nonce)
	if err != nil {
		return nil, err
	}
	fragments := make([]Fragment, n)
	for i := range fragments {
		fragments[i], err = NewFragment(
			order.ID,
			order.Type,
			order.Parity,
			order.Settlement,
			order.Expiry,
			tokens[i],
			CoExpShare{Co: priceCos[i], Exp: priceExps[i]},
			CoExpShare{Co: volumeCos[i], Exp: volumeExps[i]},
			CoExpShare{Co: minimumVolumeCos[i], Exp: minimumVolumeExps[i]},
			nonces[i],
		)
		if err != nil {
			return nil, err
		}
	}
	return fragments, nil
}

// Hash returns the Keccak256 hash of an Order. This hash is used to create the
// ID and signature for an Order. Returns a zero-d hash if the order cannot be
// marshaled into bytes.
func (order *Order) Hash() [32]byte {
	data, err := order.MarshalBinary()
	if err != nil {
		return [32]byte{}
	}
	hash := crypto.Keccak256(data)
	hash32 := [32]byte{}
	copy(hash32[:], hash)
	return hash32
}

// PrefixHash returns the data used in the prefix header for the order. Returns
// an empty slice if the order cannot be marshaled into bytes.
func (order *Order) PrefixHash() []byte {
	buf := new(bytes.Buffer)

	// Marshal the prefix data
	if err := binary.Write(buf, binary.BigEndian, order.Type); err != nil {
		return []byte{}
	}
	if err := binary.Write(buf, binary.BigEndian, uint64(order.Expiry.Unix())); err != nil {
		return []byte{}
	}
	if err := binary.Write(buf, binary.BigEndian, order.Nonce); err != nil {
		return []byte{}
	}

	return buf.Bytes()
}

// Equal returns an equality check between two Orders.
func (order *Order) Equal(other *Order) bool {
	return bytes.Equal(order.ID[:], other.ID[:]) &&
		order.Type == other.Type &&
		order.Settlement == other.Settlement &&
		order.Expiry.Equal(other.Expiry) &&
		order.Tokens == other.Tokens &&
		order.Price == other.Price &&
		order.Volume == other.Volume &&
		order.MinimumVolume == other.MinimumVolume &&
		order.Nonce == other.Nonce
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (order *Order) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Marshal the prefix data
	if err := binary.Write(buf, binary.BigEndian, order.PrefixHash()); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, order.Settlement); err != nil {
		return nil, err
	}
	if order.Parity == ParityBuy {
		if err := binary.Write(buf, binary.BigEndian, order.Tokens); err != nil {
			return nil, err
		}
	} else {
		if err := binary.Write(buf, binary.BigEndian, (order.Tokens<<32)|(order.Tokens>>32)); err != nil {
			return nil, err
		}
	}
	// Price is packed as a uint256
	if err := binary.Write(buf, binary.BigEndian, uint64(0)); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, uint64(0)); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, uint64(0)); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, order.Price); err != nil {
		return nil, err
	}
	// Volume is packed as a uint256
	if err := binary.Write(buf, binary.BigEndian, uint64(0)); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, uint64(0)); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, uint64(0)); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, order.Volume); err != nil {
		return nil, err
	}
	// Minimum volume is packed as a uint256
	if err := binary.Write(buf, binary.BigEndian, uint64(0)); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, uint64(0)); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, uint64(0)); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, order.MinimumVolume); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// PriceToCoExp converts a price from uint64 to CoExp format.
func PriceToCoExp(price uint64) CoExp {
	if price == 0 {
		return CoExp{
			Co:  0,
			Exp: 26,
		}
	}
	if price < 10 {
		return CoExp{
			Co:  price * 200,
			Exp: 26,
		}
	}
	if price < 100 {
		return CoExp{
			Co:  price * 20,
			Exp: 27,
		}
	}
	if price < 1000 {
		return CoExp{
			Co:  price * 2,
			Exp: 28,
		}
	}
	if price < 10000 {
		return CoExp{
			Co:  price / 5,
			Exp: 29,
		}
	}
	coExp := PriceToCoExp(price / 10)
	return CoExp{
		Co:  coExp.Co,
		Exp: coExp.Exp + 1,
	}
}

func VolumeToCoExp(volume uint64) CoExp {
	if volume == 0 {
		return CoExp{
			Co:  0,
			Exp: 0,
		}
	}
	if volume < 10 {
		return CoExp{
			Co:  volume * 5,
			Exp: 0,
		}
	}
	if volume < 100 {
		return CoExp{
			Co:  volume / 2,
			Exp: 1,
		}
	}
	coExp := VolumeToCoExp(volume / 10)
	return CoExp{
		Co:  coExp.Co,
		Exp: coExp.Exp + 1,
	}
}

func PriceFromCoExp(co uint64, exp uint64) uint64 {
	x := big.NewInt(10)
	y := big.NewInt(0)

	y.SetUint64(exp - 26)
	x.Exp(x, y, nil)

	y.SetUint64(co)
	x.Mul(x, y)

	y.SetUint64(200)
	x.Div(x, y)

	return x.Uint64()
}

func VolumeFromCoExp(co uint64, exp uint64) uint64 {
	x := big.NewInt(10)
	y := big.NewInt(0)

	y.SetUint64(exp)
	x.Exp(x, y, nil)

	y.SetUint64(co)
	x.Mul(x, y)

	y.SetUint64(5)
	x.Div(x, y)

	return x.Uint64()
}
