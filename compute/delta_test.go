package compute_test

import (
	"math/big"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/republicprotocol/go-order-compute"
)

var _ = Describe("Finals and final fragments", func() {

	n := int64(8)
	k := int64(6)
	prime, _ := big.NewInt(0).SetString("179769313486231590772930519078902473361797697894230657273430081157732675805500963132708477322407536021120113879871393357658789768814416622492847430639474124377767893424865485276302219601246094119453082952085005768838150682342462881473913110540827237163350510684586298239947245938479716304835356329624224137859", 10)

	Context("when serializing IDs to strings", func() {

		It("should return the string representation of the ID", func() {
			final, err := computeRandomFinal(n, k, prime)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(final.ID.String()).Should(Equal(string(final.ID)))
		})
	})

	Context("when using a delta builder", func() {

		It("should only return a delta after receiving k delta fragments", func() {
			lhs, err := compute.NewOrder(compute.OrderTypeLimit, compute.OrderParityBuy, time.Now().Add(time.Hour), compute.CurrencyCodeBTC, compute.CurrencyCodeETH, big.NewInt(10), big.NewInt(1000), big.NewInt(100), big.NewInt(0)).Split(n, k, prime)
			Ω(err).ShouldNot(HaveOccurred())
			rhs, err := compute.NewOrder(compute.OrderTypeLimit, compute.OrderParitySell, time.Now().Add(time.Hour), compute.CurrencyCodeBTC, compute.CurrencyCodeETH, big.NewInt(10), big.NewInt(1000), big.NewInt(100), big.NewInt(0)).Split(n, k, prime)
			Ω(err).ShouldNot(HaveOccurred())
			deltaFragments := make([]*compute.DeltaFragment, n)
			for i := range deltaFragments {
				deltaFragment, err := compute.NewDeltaFragment(lhs[i], rhs[i], prime)
				Ω(err).ShouldNot(HaveOccurred())
				deltaFragments[i] = deltaFragment
			}

			builder := compute.NewDeltaBuilder(k, prime)
			for i := int64(0); i < k-1; i++ {
				delta, err := builder.InsertDeltaFragment(deltaFragments[i])
				Ω(err).ShouldNot(HaveOccurred())
				Ω(delta).Should(BeNil())
			}
			delta, err := builder.InsertDeltaFragment(deltaFragments[k-1])
			Ω(err).ShouldNot(HaveOccurred())
			Ω(delta).ShouldNot(BeNil())
		})

		It("should not return a delta after the first k delta fragments", func() {
			lhs, err := compute.NewOrder(compute.OrderTypeLimit, compute.OrderParityBuy, time.Now().Add(time.Hour), compute.CurrencyCodeBTC, compute.CurrencyCodeETH, big.NewInt(10), big.NewInt(1000), big.NewInt(100), big.NewInt(0)).Split(n, k, prime)
			Ω(err).ShouldNot(HaveOccurred())
			rhs, err := compute.NewOrder(compute.OrderTypeLimit, compute.OrderParitySell, time.Now().Add(time.Hour), compute.CurrencyCodeBTC, compute.CurrencyCodeETH, big.NewInt(10), big.NewInt(1000), big.NewInt(100), big.NewInt(0)).Split(n, k, prime)
			Ω(err).ShouldNot(HaveOccurred())
			deltaFragments := make([]*compute.DeltaFragment, n)
			for i := range deltaFragments {
				deltaFragment, err := compute.NewDeltaFragment(lhs[i], rhs[i], prime)
				Ω(err).ShouldNot(HaveOccurred())
				deltaFragments[i] = deltaFragment
			}

			builder := compute.NewDeltaBuilder(k, prime)
			for i := int64(0); i < k-1; i++ {
				delta, err := builder.InsertDeltaFragment(deltaFragments[i])
				Ω(err).ShouldNot(HaveOccurred())
				Ω(delta).Should(BeNil())
			}
			delta, err := builder.InsertDeltaFragment(deltaFragments[k-1])
			Ω(err).ShouldNot(HaveOccurred())
			Ω(delta).ShouldNot(BeNil())

			for i := int64(0); i < n; i++ {
				delta, err := builder.InsertDeltaFragment(deltaFragments[i])
				Ω(err).ShouldNot(HaveOccurred())
				Ω(delta).Should(BeNil())
			}
		})

		It("should not return a delta using k non-unique fragments", func() {
			lhs, err := compute.NewOrder(compute.OrderTypeLimit, compute.OrderParityBuy, time.Now().Add(time.Hour), compute.CurrencyCodeBTC, compute.CurrencyCodeETH, big.NewInt(10), big.NewInt(1000), big.NewInt(100), big.NewInt(0)).Split(n, k, prime)
			Ω(err).ShouldNot(HaveOccurred())
			rhs, err := compute.NewOrder(compute.OrderTypeLimit, compute.OrderParitySell, time.Now().Add(time.Hour), compute.CurrencyCodeBTC, compute.CurrencyCodeETH, big.NewInt(10), big.NewInt(1000), big.NewInt(100), big.NewInt(0)).Split(n, k, prime)
			Ω(err).ShouldNot(HaveOccurred())
			deltaFragments := make([]*compute.DeltaFragment, n)
			for i := range deltaFragments {
				deltaFragment, err := compute.NewDeltaFragment(lhs[i], rhs[i], prime)
				Ω(err).ShouldNot(HaveOccurred())
				deltaFragments[i] = deltaFragment
			}

			builder := compute.NewDeltaBuilder(k, prime)
			for i := int64(0); i < k-1; i++ {
				delta, err := builder.InsertDeltaFragment(deltaFragments[i])
				Ω(err).ShouldNot(HaveOccurred())
				Ω(delta).Should(BeNil())
			}
			for i := int64(0); i < k-1; i++ {
				delta, err := builder.InsertDeltaFragment(deltaFragments[i])
				Ω(err).ShouldNot(HaveOccurred())
				Ω(delta).Should(BeNil())
			}
		})
	})
})

func computeRandomFinal(n, k int64, prime *big.Int) (*compute.Delta, error) {
	lhs, err := compute.NewOrder(compute.OrderTypeLimit, compute.OrderParityBuy, time.Now().Add(time.Hour), compute.CurrencyCodeBTC, compute.CurrencyCodeETH, big.NewInt(10), big.NewInt(1000), big.NewInt(100), big.NewInt(0)).Split(n, k, prime)
	if err != nil {
		return nil, err
	}
	rhs, err := compute.NewOrder(compute.OrderTypeLimit, compute.OrderParitySell, time.Now().Add(time.Hour), compute.CurrencyCodeBTC, compute.CurrencyCodeETH, big.NewInt(10), big.NewInt(1000), big.NewInt(100), big.NewInt(0)).Split(n, k, prime)
	if err != nil {
		return nil, err
	}
	return computeFinalFromOrderFragments(lhs, rhs, n, prime)
}

func computeFinalFromOrderFragments(lhs []*compute.OrderFragment, rhs []*compute.OrderFragment, n int64, prime *big.Int) (*compute.Delta, error) {
	// Generate pairwise computations for each fragment class.
	deltaFragments := make([]*compute.DeltaFragment, n)
	for i := range deltaFragments {
		deltaFragment, err := compute.NewDeltaFragment(lhs[i], rhs[i], prime)
		if err != nil {
			return nil, err
		}
		deltaFragments[i] = deltaFragment
	}
	// Combine them into a final result.
	return compute.NewDelta(deltaFragments, prime)

}