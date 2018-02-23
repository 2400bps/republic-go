package compute

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/republicprotocol/go-sss"
)

type FinalShard struct {
	Signature []byte
	Finals    []*DeltaFragment
}

func NewFinalShard() FinalShard {
	return FinalShard{
		Finals: []*DeltaFragment{},
	}
}

// A FinalID is the Keccak256 hash of the OrderIDs that were used to compute
// the respective Result.
type FinalID []byte

// Equals checks if two ResultIDs are equal in value.
func (id FinalID) Equals(other FinalID) bool {
	return bytes.Equal(id, other)
}

// String returns the ResultID as a string.
func (id FinalID) String() string {
	return string(id)
}

// A Final is the publicly computed value of comparing two Orders.
type Final struct {
	ID          FinalID
	BuyOrderID  OrderID
	SellOrderID OrderID
	FstCode     *big.Int
	SndCode     *big.Int
	Price       *big.Int
	MaxVolume   *big.Int
	MinVolume   *big.Int
}

func NewFinal(deltaFragments []*DeltaFragment, prime *big.Int) (*Final, error) {
	// Check that all DeltaFragments are compatible with each other.
	err := isCompatible(deltaFragments)
	if err != nil {
		return nil, err
	}

	// Collect sss.Shares across all DeltaFragments.
	k := len(deltaFragments)
	fstCodeShares := make(sss.Shares, k)
	sndCodeShares := make(sss.Shares, k)
	priceShares := make(sss.Shares, k)
	maxVolumeShares := make(sss.Shares, k)
	minVolumeShares := make(sss.Shares, k)
	for i, resultFragment := range deltaFragments {
		fstCodeShares[i] = resultFragment.FstCodeShare
		sndCodeShares[i] = resultFragment.SndCodeShare
		priceShares[i] = resultFragment.PriceShare
		maxVolumeShares[i] = resultFragment.MaxVolumeShare
		minVolumeShares[i] = resultFragment.MinVolumeShare
	}

	// Join the sss.Shares into a Final.
	final := &Final{
		BuyOrderID:  deltaFragments[0].BuyOrderID,
		SellOrderID: deltaFragments[0].SellOrderID,
	}
	final.FstCode = sss.Join(prime, fstCodeShares)
	final.SndCode = sss.Join(prime, sndCodeShares)
	final.Price = sss.Join(prime, priceShares)
	final.MaxVolume = sss.Join(prime, maxVolumeShares)
	final.MinVolume = sss.Join(prime, minVolumeShares)

	// Compute the FinalID and return the Final.
	final.ID = FinalID(crypto.Keccak256(final.BuyOrderID[:], final.SellOrderID[:]))
	return final, nil
}

func (final *Final) IsMatch(prime *big.Int) bool {
	zeroThreshold := big.NewInt(0).Div(prime, big.NewInt(2))
	if final.FstCode.Cmp(big.NewInt(0)) != 0 {
		return false
	}
	if final.SndCode.Cmp(big.NewInt(0)) != 0 {
		return false
	}
	if final.Price.Cmp(zeroThreshold) == 1 {
		return false
	}
	if final.MaxVolume.Cmp(zeroThreshold) == 1 {
		return false
	}
	if final.MinVolume.Cmp(zeroThreshold) == 1 {
		return false
	}
	return true
}

// A DeltaFragmentID is the Keccak256 hash of its OrderFragmentIDs.
type DeltaFragmentID []byte

// Equals checks if two DeltaFragmentIDs are equal in value.
func (id DeltaFragmentID) Equals(other DeltaFragmentID) bool {
	return bytes.Equal(id, other)
}

// String returns the DeltaFragmentID as a string.
func (id DeltaFragmentID) String() string {
	return string(id)
}

// A DeltaFragment is a secret share of a Final. Is is performing a
// computation over two OrderFragments.
type DeltaFragment struct {
	// Public data.
	ID                  DeltaFragmentID
	BuyOrderID          OrderID
	SellOrderID         OrderID
	BuyOrderFragmentID  OrderFragmentID
	SellOrderFragmentID OrderFragmentID

	// Private data.
	FstCodeShare   sss.Share
	SndCodeShare   sss.Share
	PriceShare     sss.Share
	MaxVolumeShare sss.Share
	MinVolumeShare sss.Share
}

func NewDeltaFragment(left *OrderFragment, right *OrderFragment, prime *big.Int) (*DeltaFragment, error) {
	if err := left.IsCompatible(right); err != nil {
		return nil, err
	}
	var buyOrderFragment *OrderFragment
	var sellOrderFragment *OrderFragment
	if left.OrderParity == OrderParityBuy {
		buyOrderFragment = left
		sellOrderFragment = right
	} else {
		buyOrderFragment = right
		sellOrderFragment = left
	}

	deltaFragment, err := buyOrderFragment.Sub(sellOrderFragment, prime)
	if err != nil {
		return nil, err
	}
	return deltaFragment, nil
}

// Equals checks if two DeltaFragments are equal in value.
func (deltaFragment *DeltaFragment) Equals(other *DeltaFragment) bool {
	return deltaFragment.ID.Equals(other.ID) &&
		deltaFragment.BuyOrderID.Equals(other.BuyOrderID) &&
		deltaFragment.SellOrderID.Equals(other.SellOrderID) &&
		deltaFragment.BuyOrderFragmentID.Equals(other.BuyOrderFragmentID) &&
		deltaFragment.SellOrderFragmentID.Equals(other.SellOrderFragmentID) &&
		deltaFragment.FstCodeShare.Key == other.FstCodeShare.Key &&
		deltaFragment.FstCodeShare.Value.Cmp(other.FstCodeShare.Value) == 0 &&
		deltaFragment.SndCodeShare.Key == other.SndCodeShare.Key &&
		deltaFragment.SndCodeShare.Value.Cmp(other.SndCodeShare.Value) == 0 &&
		deltaFragment.PriceShare.Key == other.PriceShare.Key &&
		deltaFragment.PriceShare.Value.Cmp(other.PriceShare.Value) == 0 &&
		deltaFragment.MaxVolumeShare.Key == other.MaxVolumeShare.Key &&
		deltaFragment.MaxVolumeShare.Value.Cmp(other.MaxVolumeShare.Value) == 0 &&
		deltaFragment.MinVolumeShare.Key == other.MinVolumeShare.Key &&
		deltaFragment.MinVolumeShare.Value.Cmp(other.MinVolumeShare.Value) == 0
}

// IsCompatible returns an error when the two deltaFragments do not have
// the same share indices.
func isCompatible(deltaFragments []*DeltaFragment) error {
	if len(deltaFragments) == 0 {
		return NewEmptySliceError("result fragments")
	}
	buyOrderID := deltaFragments[0].BuyOrderID
	sellOrderID := deltaFragments[0].SellOrderID
	for i := range deltaFragments {
		if !deltaFragments[i].BuyOrderID.Equals(buyOrderID) {
			return NewOrderFragmentationError(0, int64(i))
		}
		if !deltaFragments[i].SellOrderID.Equals(sellOrderID) {
			return NewOrderFragmentationError(0, int64(i))
		}
	}
	return nil
}
