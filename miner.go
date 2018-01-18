package miner

import (
	"log"
	"math/big"
	"runtime"

	"github.com/republicprotocol/go-do"
	"github.com/republicprotocol/go-identity"
	"github.com/republicprotocol/go-network"
	"github.com/republicprotocol/go-order-compute"
)

// TODO: Do not make this values constant.
var (
	N        = int64(3)
	K        = int64(2)
	Prime, _ = big.NewInt(0).SetString("179769313486231590772930519078902473361797697894230657273430081157732675805500963132708477322407536021120113879871393357658789768814416622492847430639474124377767893424865485276302219601246094119453082952085005768838150682342462881473913110540827237163350510684586298239947245938479716304835356329624224137859", 10)
)

type Miner struct {
	*network.Node
	*compute.ComputationMatrix
}

func NewMiner(config Config) (*Miner, error) {
	miner := &Miner{}
	node, err := network.NewNode(config.Multi, config.BootstrapMultis, miner)
	if err != nil {
		return nil, err
	}
	miner.Node = node
	return miner, nil
}

func (miner *Miner) OnPingReceived(peer identity.MultiAddress) {
}

func (miner *Miner) OnOrderFragmentReceived(orderFragment compute.OrderFragment) {
	// miner.ComputationMatrix.FillComputations(&orderFragment)
}

func (miner *Miner) OnComputedOrderFragmentReceived(orderFragment compute.OrderFragment) {
	// miner.ComputeReconstruction(&orderFragment)
}

func (miner *Miner) Mine(quit chan struct{}) {
	go func() {
		if err := miner.Serve(); err != nil {
			// TODO: Do something other than die.
			log.Fatal(err)
		}
	}()
	for {
		select {
		case <-quit:
			miner.Stop()
			return
		default:
			miner.ComputeAll()
		}
	}
}

func (miner Miner) ComputeAll() {
	numberOfCPUs := runtime.NumCPU()
	computations := miner.ComputationMatrix.WaitForComputations(numberOfCPUs)
	resultFragments := make([]*compute.ResultFragment, len(computations))
	do.CoForAll(computations, func(i int) {
		resultFragment, err := miner.Compute(computations[i])
		if err != nil {
			return
		}
		resultFragments[i] = resultFragment
	})
	go func() {
		resultFragmentsOk := make([]*compute.ResultFragment, 0, len(resultFragments))
		for _, resultFragment := range resultFragments {
			if resultFragment != nil {
				resultFragmentsOk = append(resultFragmentsOk, resultFragment)
			}
		}
		results, _ := miner.ComputationMatrix.AddResultFragments(K, Prime, resultFragmentsOk)
		for _, result := range results {
			if result.IsMatch() {
				log.Println("buy =", result.BuyOrderID, ",", "sell =", result.SellOrderID)
			}
		}
	}()
}

// Compute the required computation on two OrderFragments and send the result
// to all Miners in the M Network.
// TODO: Send computed order fragments to the M Network instead of all peers.
func (miner Miner) Compute(computation *compute.Computation) (*compute.ResultFragment, error) {
	resultFragment, err := computation.Sub(Prime)
	if err != nil {
		return nil, err
	}
	go func() {
		for _, multi := range miner.DHT.MultiAddresses() {
			network.RPCSendComputedOrderFragment(multi, resultFragment)
		}
	}()
	return resultFragment, nil
}
