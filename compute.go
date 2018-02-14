package compute

import (
	"math/big"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/crypto"
)

type ComputationID []byte

type Computation struct {
	ID                ComputationID
	BuyOrderFragment  *OrderFragment
	SellOrderFragment *OrderFragment
}

func NewComputation(left *OrderFragment, right *OrderFragment) (*Computation, error) {
	if err := left.IsCompatible(right); err != nil {
		return nil, err
	}
	computation := &Computation{}
	if left.OrderParity == OrderParityBuy {
		computation.BuyOrderFragment = left
		computation.SellOrderFragment = right
	} else {
		computation.BuyOrderFragment = right
		computation.SellOrderFragment = left
	}
	computation.ID = ComputationID(crypto.Keccak256(computation.BuyOrderFragment.ID[:], computation.SellOrderFragment.ID[:]))
	return computation, nil
}

func (computation *Computation) Add(prime *big.Int) (*ResultFragment, error) {
	return computation.BuyOrderFragment.Add(computation.SellOrderFragment, prime)
}

func (computation *Computation) Sub(prime *big.Int) (*ResultFragment, error) {
	return computation.BuyOrderFragment.Sub(computation.SellOrderFragment, prime)
}

type Computer struct {
	orderFragments []*OrderFragment

	computationsMu       *sync.Mutex
	computationsLeftCond *sync.Cond
	computations         []*Computation
	computationsLeft     int64
	computationsMarker   map[string]struct{}

	resultsMu       *sync.Mutex
	results         map[string]*Result
	resultFragments map[string][]*ResultFragment
}

func NewComputationMatrix() *Computer {
	return &Computer{
		orderFragments: []*OrderFragment{},

		computationsMu:       new(sync.Mutex),
		computationsLeftCond: sync.NewCond(new(sync.Mutex)),
		computations:         []*Computation{},
		computationsLeft:     0,
		computationsMarker:   map[string]struct{}{},

		resultsMu:       new(sync.Mutex),
		results:         map[string]*Result{},
		resultFragments: map[string][]*ResultFragment{},
	}
}

func (matrix *Computer) AddOrderFragment(orderFragment *OrderFragment) {
	matrix.computationsMu.Lock()
	defer matrix.computationsMu.Unlock()

	for _, rhs := range matrix.orderFragments {
		if orderFragment.ID.Equals(rhs.ID) {
			return
		}
	}

	for _, other := range matrix.orderFragments {
		if orderFragment.OrderID.Equals(other.OrderID) {
			continue
		}
		if err := orderFragment.IsCompatible(other); err != nil {
			continue
		}
		computation, err := NewComputation(orderFragment, other)
		if err != nil {
			continue
		}
		matrix.computations = append(matrix.computations, computation)
		atomic.AddInt64(&matrix.computationsLeft, 1)
	}

	matrix.orderFragments = append(matrix.orderFragments, orderFragment)
	if atomic.LoadInt64(&matrix.computationsLeft) > 0 {
		matrix.computationsLeftCond.Signal()
	}
}

func (matrix *Computer) WaitForComputations(max int) []*Computation {
	matrix.computationsLeftCond.L.Lock()
	defer matrix.computationsLeftCond.L.Unlock()
	for atomic.LoadInt64(&matrix.computationsLeft) == 0 {
		matrix.computationsLeftCond.Wait()
	}

	matrix.computationsMu.Lock()
	defer matrix.computationsMu.Unlock()

	computations := make([]*Computation, 0, max)
	for _, computation := range matrix.computations {
		if _, ok := matrix.computationsMarker[string(computation.ID)]; !ok {
			matrix.computationsMarker[string(computation.ID)] = struct{}{}
			computations = append(computations, computation)
			if len(computations) == max {
				break
			}
		}
	}
	atomic.AddInt64(&matrix.computationsLeft, -int64(len(computations)))
	return computations
}

func (matrix *Computer) AddResultFragments(resultFragments []*ResultFragment, k int64, prime *big.Int) ([]*Result, error) {
	matrix.resultsMu.Lock()
	defer matrix.resultsMu.Unlock()

	results := make([]*Result, 0, len(resultFragments))
	for _, resultFragment := range resultFragments {
		resultID := ResultID(crypto.Keccak256(resultFragment.BuyOrderID[:], resultFragment.SellOrderID[:]))

		// Check that this result fragment has not been collected yet.
		resultFragmentIsUnique := true
		for _, candidate := range matrix.resultFragments[string(resultID)] {
			if candidate.ID.Equals(resultFragment.ID) {
				resultFragmentIsUnique = false
				break
			}
		}
		if resultFragmentIsUnique {
			matrix.resultFragments[string(resultID)] = append(matrix.resultFragments[string(resultID)], resultFragment)
		}

		if int64(len(matrix.resultFragments[string(resultID)])) >= k {
			if result, ok := matrix.results[string(resultID)]; result != nil && ok {
				// FIXME: At the moment we are only returning new results. Do
				// we want to return results we have already found?
				continue
			}
			result := NewResult(matrix.resultFragments[string(resultID)], prime)
			matrix.results[string(resultID)] = result
			results = append(results, result)
		}
	}
	return results, nil
}

func (matrix *Computer) ComputationsLeft() int64 {
	matrix.computationsLeftCond.L.Lock()
	defer matrix.computationsLeftCond.L.Unlock()

	return matrix.computationsLeft
}
