package miner_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Miners", func() {
	It("should fail", func() {
		Ω(true).Should(Equal(false))
	})
})
