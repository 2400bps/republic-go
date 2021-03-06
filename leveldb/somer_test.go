package leveldb_test

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/republicprotocol/republic-go/leveldb"

	"github.com/republicprotocol/republic-go/ome"
	"github.com/republicprotocol/republic-go/order"
	"github.com/republicprotocol/republic-go/registry"
	"github.com/republicprotocol/republic-go/testutils"
)

var _ = Describe("Somer storage", func() {

	computations := make([]ome.Computation, 100)
	buyFragments := make([]order.Fragment, 100)
	sellFragments := make([]order.Fragment, 100)
	dbFolder := "./tmp/"
	dbFile := dbFolder + "db"
	epoch := registry.Epoch{}

	BeforeEach(func() {
		for i := 0; i < 100; i++ {
			buyOrd := order.NewOrder(order.ParityBuy, order.TypeMidpoint, time.Now(), order.SettlementRenEx, order.TokensETHREN, uint64(i), uint64(i), uint64(i), uint64(i))
			buyOrdFragments, err := buyOrd.Split(3, 2)
			sellOrd := order.NewOrder(order.ParitySell, order.TypeMidpoint, time.Now(), order.SettlementRenEx, order.TokensETHREN, uint64(i), uint64(i), uint64(i), uint64(i))
			sellOrdFragments, err := sellOrd.Split(3, 2)
			Expect(err).ShouldNot(HaveOccurred())
			computations[i] = ome.NewComputation([32]byte{byte(i)}, buyOrdFragments[0], sellOrdFragments[0], ome.ComputationStateMatched, true)
			buyFragments[i] = buyFragments[0]
			sellFragments[i] = sellFragments[0]
			_, epoch, err = testutils.RandomEpoch(1)
		}
	})

	AfterEach(func() {
		os.RemoveAll(dbFolder)
	})

	Context("when iterating through out of range data", func() {
		It("should trigger an out of range error", func() {
			db := newDB(dbFile)
			somerComputationTable := NewSomerComputationTable(db)
			somerOrderFragmentTable := NewSomerOrderFragmentTable(db)

			// Put the computations into the table and attempt to retrieve
			for i := 0; i < len(computations); i++ {
				err := somerComputationTable.PutComputation(computations[i])
				Expect(err).ShouldNot(HaveOccurred())
				somerOrderFragmentTable.PutBuyOrderFragment(epoch.Hash, buyFragments[i], "trader1", uint64(i), order.Open)
				Expect(err).ShouldNot(HaveOccurred())
				somerOrderFragmentTable.PutSellOrderFragment(epoch.Hash, sellFragments[i], "trader2", uint64(i), order.Open)
				Expect(err).ShouldNot(HaveOccurred())
			}
			for i := 0; i < len(computations); i++ {
				com, err := somerComputationTable.Computation(computations[i].ID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(com.Equal(&computations[i])).Should(BeTrue())
				buyFragment, trader, _, _, err := somerOrderFragmentTable.BuyOrderFragment(epoch.Hash, buyFragments[i].OrderID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(buyFragment.Equal(&buyFragments[i])).Should(BeTrue())
				Expect(trader).To(Equal("trader1"))
				sellFragment, trader, _, _, err := somerOrderFragmentTable.SellOrderFragment(epoch.Hash, sellFragments[i].OrderID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(sellFragment.Equal(&sellFragments[i])).Should(BeTrue())
				Expect(trader).To(Equal("trader2"))
			}

			comsIter, err := somerComputationTable.Computations()
			defer comsIter.Release()
			for comsIter.Next() {
				_, err := comsIter.Cursor()
				Expect(err).ShouldNot(HaveOccurred())
			}

			// This is out of range so we should expect an error
			_, err = comsIter.Cursor()
			Expect(err).Should(Equal(ome.ErrCursorOutOfRange))

			buysIter, err := somerOrderFragmentTable.BuyOrderFragments(epoch.Hash)
			Expect(err).ShouldNot(HaveOccurred())
			defer buysIter.Release()
			for buysIter.Next() {
				_, _, _, _, err := buysIter.Cursor()
				Expect(err).ShouldNot(HaveOccurred())
			}

			// This is out of range so we should expect an error
			_, _, _, _, err = buysIter.Cursor()
			Expect(err).Should(Equal(ome.ErrCursorOutOfRange))

			sellsIter, err := somerOrderFragmentTable.SellOrderFragments(epoch.Hash)
			Expect(err).ShouldNot(HaveOccurred())
			defer sellsIter.Release()
			for sellsIter.Next() {
				_, _, _, _, err := sellsIter.Cursor()
				Expect(err).ShouldNot(HaveOccurred())
			}

			// This is out of range so we should expect an error
			_, _, _, _, err = sellsIter.Cursor()
			Expect(err).Should(Equal(ome.ErrCursorOutOfRange))

		})
	})

	Context("when updating order fragment status", func() {
		It("should return updated status", func() {
			db := newDB(dbFile)
			somerOrderFragmentTable := NewSomerOrderFragmentTable(db)

			// Put the computations into the table and attempt to retrieve
			for i := 0; i < len(computations); i++ {
				err := somerOrderFragmentTable.PutBuyOrderFragment(epoch.Hash, buyFragments[i], "trader1", uint64(i), order.Open)
				Expect(err).ShouldNot(HaveOccurred())
				err = somerOrderFragmentTable.PutSellOrderFragment(epoch.Hash, sellFragments[i], "trader2", uint64(i), order.Open)
				Expect(err).ShouldNot(HaveOccurred())
			}

			for i := 0; i < len(computations); i++ {
				err := somerOrderFragmentTable.UpdateBuyOrderFragmentStatus(epoch.Hash, buyFragments[i].OrderID, order.Canceled)
				Expect(err).ShouldNot(HaveOccurred())
				err = somerOrderFragmentTable.UpdateSellOrderFragmentStatus(epoch.Hash, sellFragments[i].OrderID, order.Canceled)
				Expect(err).ShouldNot(HaveOccurred())
			}

			for i := 0; i < len(computations); i++ {
				_, _, _, status, err := somerOrderFragmentTable.BuyOrderFragment(epoch.Hash, buyFragments[i].OrderID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status).To(Equal(order.Canceled))
				_, _, _, status, err = somerOrderFragmentTable.SellOrderFragment(epoch.Hash, sellFragments[i].OrderID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(status).To(Equal(order.Canceled))
			}
		})
	})
})
