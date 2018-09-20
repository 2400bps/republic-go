package ome_test

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/republicprotocol/republic-go/ome"

	"github.com/republicprotocol/republic-go/leveldb"
	"github.com/republicprotocol/republic-go/testutils"
)

var numberOfComputationsToTest = 100

var _ = Describe("Confirmer", func() {
	var confirmer Confirmer
	var contract *omeBinder
	var comStorer ComputationStorer
	var fragmentStorer OrderFragmentStorer

	BeforeEach(func() {
		var err error
		depth, pollInterval := uint(0), time.Second
		contract = newOmeBinder()
		db, err := leveldb.NewStore("./data.out", time.Hour, time.Hour)
		Expect(err).ShouldNot(HaveOccurred())
		comStorer = db.SomerComputationStore()
		fragmentStorer = db.SomerOrderFragmentStore()
		confirmer = NewConfirmer(comStorer, fragmentStorer, contract, pollInterval, depth)
	})

	AfterEach(func() {
		os.RemoveAll("./data.out")
	})

	XIt("should be able to confirm order on the ren ledger", func(d Done) {
		defer close(d)

		done := make(chan struct{})
		orderMatches := make(chan Computation)
		orderIDs := map[[32]byte]struct{}{}
		computations := make([]Computation, numberOfComputationsToTest)

		var err error
		for i := 0; i < numberOfComputationsToTest; i++ {
			computations[i], err = testutils.RandomComputation()
			Expect(err).ShouldNot(HaveOccurred())
			orderIDs[computations[i].Buy.OrderID] = struct{}{}
			orderIDs[computations[i].Sell.OrderID] = struct{}{}
			comStorer.PutComputation(computations[i])
		}

		// Open all the orders
		for i := 0; i < numberOfComputationsToTest; i++ {
			err := contract.OpenBuyOrder([65]byte{}, computations[i].Buy.OrderID)
			Expect(err).ShouldNot(HaveOccurred())
			err = contract.OpenSellOrder([65]byte{}, computations[i].Sell.OrderID)
			Expect(err).ShouldNot(HaveOccurred())
		}

		go func() {
			defer GinkgoRecover()
			defer close(done)

			for i := 0; i < numberOfComputationsToTest; i++ {
				orderMatches <- computations[i]
			}
			time.Sleep(5 * time.Second)
		}()

		confirmedMatches, errs := confirmer.Confirm(done, orderMatches)
		go func() {
			defer GinkgoRecover()

			for err := range errs {
				Expect(err).ShouldNot(HaveOccurred())
			}
		}()
		go func() {
			defer GinkgoRecover()

			for match := range confirmedMatches {
				_, ok := orderIDs[match.Buy.OrderID]
				Ω(ok).Should(BeTrue())
				delete(orderIDs, match.Buy.OrderID)

				_, ok = orderIDs[match.Sell.OrderID]
				Ω(ok).Should(BeTrue())
				delete(orderIDs, match.Sell.OrderID)
			}
			Ω(len(orderIDs)).Should(Equal(0))
		}()
	}, 100)

	It("should return error for invalid computations", func() {
		done := make(chan struct{})
		orderMatches := make(chan Computation)
		orderIDs := map[[32]byte]struct{}{}
		computations := make([]Computation, numberOfComputationsToTest)

		var err error
		for i := 0; i < numberOfComputationsToTest; i++ {
			computations[i], err = testutils.RandomComputation()
			Expect(err).ShouldNot(HaveOccurred())
			orderIDs[computations[i].Buy.OrderID] = struct{}{}
			orderIDs[computations[i].Sell.OrderID] = struct{}{}
			comStorer.PutComputation(computations[i])
		}

		go func() {
			defer GinkgoRecover()
			defer close(done)

			for i := 0; i < numberOfComputationsToTest; i++ {
				orderMatches <- computations[i]
			}
			time.Sleep(5 * time.Second)
		}()

		confirmedMatches, errs := confirmer.Confirm(done, orderMatches)
		go func() {
			defer GinkgoRecover()

			for err := range errs {
				Expect(err).Should(HaveOccurred())
			}
		}()

		Expect(len(confirmedMatches)).Should(BeZero())
	})
})
