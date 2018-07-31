package identity

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"
)

// Codes for extracting specific protocol values from a MultiAddress.
const (
	IP4Code      = 0x0004
	IP6Code      = 0x0029
	TCPCode      = 0x0006
	RepublicCode = 0x0065
)

// Add the Republic Protocol when the package is initialized.
func init() {
	republic := multiaddr.Protocol{
		Code:       RepublicCode,
		Size:       multiaddr.LengthPrefixedVarSize,
		Name:       "republic",
		Path:       false,
		Transcoder: multiaddr.NewTranscoderFromFunctions(republicStB, republicBtS, nil),
	}
	multiaddr.AddProtocol(republic)
}

// MultiAddress is an alias.
type MultiAddress struct {
	Signature []byte
	Nonce     uint64

	address          Address
	baseMultiAddress multiaddr.Multiaddr
}

type multiAddressJsonValue struct {
	Signature []byte `json:"signature"`
	Nonce     uint64 `json:"nonce"`

	Address          Address `json:"address"`
	BaseMultiAddress string  `json:"baseMultiAddress"`
}

// MultiAddresses is an alias.
type MultiAddresses []MultiAddress

// NewMultiAddressFromString parses and validates an input string. It returns a
// MultiAddress, or an error.
func NewMultiAddressFromString(s string) (MultiAddress, error) {
	multiAddress, err := multiaddr.NewMultiaddr(s)
	if err != nil {
		return MultiAddress{}, err
	}
	address, err := multiAddress.ValueForProtocol(RepublicCode)
	if err != nil {
		return MultiAddress{}, err
	}
	addressAsMultiAddress, err := multiaddr.NewMultiaddr("/republic/" + address)
	if err != nil {
		return MultiAddress{}, err
	}
	baseMultiAddress := multiAddress.Decapsulate(addressAsMultiAddress)

	return MultiAddress{[]byte{}, uint64(1), Address(address), baseMultiAddress}, err
}

// ValueForProtocol returns the value of the specific protocol in the MultiAddress
func (multiAddress MultiAddress) ValueForProtocol(code int) (string, error) {
	if code == RepublicCode {
		return multiAddress.address.String(), nil
	}
	return multiAddress.baseMultiAddress.ValueForProtocol(code)
}

// Address returns the Republic address of a MultiAddress.
func (multiAddress MultiAddress) Address() Address {
	return multiAddress.address
}

// ID returns the Republic ID of a MultiAddress.
func (multiAddress MultiAddress) ID() ID {
	return multiAddress.address.ID()
}

// String returns the MultiAddress as a plain string.
func (multiAddress MultiAddress) String() string {
	return fmt.Sprintf("%s/republic/%s", multiAddress.baseMultiAddress.String(), multiAddress.address.String())
}

// Hash returns the Keccak256 hash of a multiAddress. This hash is used to create
// signatures for a multiaddress.
func (multiAddress MultiAddress) Hash() []byte {
	nonceBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(nonceBytes, multiAddress.Nonce)
	multiaddrBytes := append([]byte(multiAddress.String()), nonceBytes...)
	return crypto.Keccak256(multiaddrBytes)
}

// MarshalJSON implements the json.Marshaler interface.
func (multiAddress MultiAddress) MarshalJSON() ([]byte, error) {
	if multiAddress.baseMultiAddress == nil {
		return []byte{}, errors.New("baseMultiAddress cannot be nil")
	}
	val := multiAddressJsonValue{
		Signature:        multiAddress.Signature,
		Nonce:            multiAddress.Nonce,
		Address:          multiAddress.address,
		BaseMultiAddress: multiAddress.baseMultiAddress.String(),
	}
	return json.Marshal(val)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (multiAddress *MultiAddress) UnmarshalJSON(data []byte) error {
	val := multiAddressJsonValue{}
	if err := json.Unmarshal(data, &val); err != nil {
		return multiAddress.UnmarshalStringJSON(data)
	}
	newMultiAddress, err := multiaddr.NewMultiaddr(val.BaseMultiAddress)
	if err != nil {
		return err
	}
	multiAddress.Signature = val.Signature
	multiAddress.Nonce = val.Nonce
	multiAddress.address = val.Address
	multiAddress.baseMultiAddress = newMultiAddress
	return nil
}

// UnmarshalStringJSON will unmarshal multi-addresses that are in string
// format to a standard multi-address struct.
func (multiAddress *MultiAddress) UnmarshalStringJSON(data []byte) error {
	multiAddressAsString := ""
	if err := json.Unmarshal(data, &multiAddressAsString); err != nil {
		return err
	}
	if multiAddressAsString == "" {
		return nil
	}
	newMultiAddress, err := NewMultiAddressFromString(multiAddressAsString)
	if err != nil {
		return err
	}
	multiAddress.baseMultiAddress = newMultiAddress.baseMultiAddress
	multiAddress.address = newMultiAddress.address
	multiAddress.Nonce = 0
	return nil
}

// ProtocolWithName returns the Protocol description with the given name.
func ProtocolWithName(s string) multiaddr.Protocol {
	return multiaddr.ProtocolWithName(s)
}

// ProtocolWithCode returns the Protocol description with the given code.
func ProtocolWithCode(c int) multiaddr.Protocol {
	return multiaddr.ProtocolWithCode(c)
}

// republicStB converts a republic address from a string to bytes.
func republicStB(s string) ([]byte, error) {
	// The address is a varint prefixed multihash string representation.
	m, err := multihash.FromB58String(s)
	if err != nil {
		return nil, fmt.Errorf("failed to parse republic addr: %s %s", s, err)
	}
	size := multiaddr.CodeToVarint(len(m))
	b := append(size, m...)
	return b, nil
}

// republicBtS converts a Republic address, encoded as bytes, to a string.
func republicBtS(b []byte) (string, error) {
	size, n, err := multiaddr.ReadVarintCode(b)
	if err != nil {
		return "", err
	}
	b = b[n:]
	if len(b) != size {
		return "", errors.New("inconsistent lengths")
	}
	m, err := multihash.Cast(b)
	if err != nil {
		return "", err
	}
	// This uses the default Bitcoin alphabet for Base58 encoding.
	return m.B58String(), nil
}
