package contracts

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/republicprotocol/republic-go/ethereum/client"
	"github.com/republicprotocol/republic-go/ethereum/ganache"
	"github.com/republicprotocol/republic-go/hyperdrive"
)

var _ = Describe("Dark Node Registrar", func() {
	Context("Local test with ganache", func() {
		It("should be able to send txs which have no conflicts", func() {
			// Connect to local ganache blockchain
			conn, err := ganache.Connect("http://localhost:8545")
			Ω(err).ShouldNot(HaveOccurred())

			// Create new transactor
			ethereumPair, err := crypto.GenerateKey()
			ethereumKey := &keystore.Key{
				Address:    crypto.PubkeyToAddress(ethereumPair.PublicKey),
				PrivateKey: ethereumPair,
			}
			auth := bind.NewKeyedTransactor(ethereumKey.PrivateKey)

			// Distribute ren and eth to the address
			err = ganache.DistributeREN(conn, auth.From)
			Ω(err).ShouldNot(HaveOccurred())

			err = ganache.DistributeEth(conn, auth.From)
			Ω(err).ShouldNot(HaveOccurred())

			// Register the account
			darknodeRegistry, err := NewDarkNodeRegistry(context.Background(), conn, auth, &bind.CallOpts{})
			Ω(err).ShouldNot(HaveOccurred())

			darknodeRegistry.SetGasLimit(1000000)
			minimumBond, err := darknodeRegistry.MinimumBond()
			Ω(err).ShouldNot(HaveOccurred())

			isRegistered, err := darknodeRegistry.IsRegistered(auth.From.Bytes())
			Ω(err).ShouldNot(HaveOccurred())
			if !isRegistered {
				transaction, err := darknodeRegistry.Register(auth.From.Bytes(), []byte{}, &minimumBond)
				Ω(err).ShouldNot(HaveOccurred())
				_, err = conn.PatchedWaitMined(context.Background(), transaction)
				Ω(err).ShouldNot(HaveOccurred())
				_, err = darknodeRegistry.WaitForEpoch()
				Ω(err).ShouldNot(HaveOccurred())
			}

			//Create newHyperdriveContract for sending Txs
			hyper, err := NewHyperdriveContract(context.Background(), conn, auth, &bind.CallOpts{})
			Ω(err).ShouldNot(HaveOccurred())
			hyper.SetGasLimit(1000000)

			tx := hyperdrive.Tx{
				Nonces: [][32]byte{
					{0, 1},
				},
			}
			transaction, err := hyper.SendTx(tx)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(transaction).ShouldNot(BeNil())
		})
	})

	Context("Tests with Ropsten", func() {
		// creatorAccountAddress := "0x9b5683153ddb50b4106Ef9296a65D26E0b87002c"
		renContractAddress := "0xad6ab5ccbd2d761d11ba7e976ba7a93a6e3dd41a"
		dnrContractAddress := "0x429b5ba768e58f1a26b58742975aaeee417f3211"
		hyperdriveAddress := "0x9db5820c2c5aa57cebe502727c98d952dae8e15f"

		BeforeEach(func() {

		})

		It("should be able to send txs with no conflicts", func() {
			conn, err := client.Connect("https://ropsten.infura.io", client.NetworkRopsten, renContractAddress, dnrContractAddress, hyperdriveAddress)
			Ω(err).ShouldNot(HaveOccurred())

			testKey, err := crypto.HexToECDSA("b44a49889a79983336d15385161533868644d35c1ea670854a0a0b4b784ae40c")
			Ω(err).ShouldNot(HaveOccurred())
			auth := bind.NewKeyedTransactor(testKey)

			//Create newHyperdriveContract for sending Txs
			hyper, err := NewHyperdriveContract(context.Background(), conn, auth, &bind.CallOpts{})
			Ω(err).ShouldNot(HaveOccurred())

			tx := hyperdrive.Tx{
				Nonces: [][32]byte{
					{7}, // Make sure you increment this number before running the test
				},
			}
			transaction, err := hyper.SendTx(tx)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(transaction).ShouldNot(BeNil())
			// log.Println("transaction is ", transaction)

			blockNumber, err := hyper.GetBlockNumberOfTx(transaction.Hash())
			Ω(err).ShouldNot(HaveOccurred())
			Ω(blockNumber).Should(BeNumerically(">", 3097093))
			// log.Println("blockNumber is ", blockNumber)
		})

		FIt("should be able to get current block number and block number of certain transaction", func() {
			conn, err := client.Connect("https://ropsten.infura.io", client.NetworkRopsten, renContractAddress, dnrContractAddress, hyperdriveAddress)
			Ω(err).ShouldNot(HaveOccurred())

			testKey, err := crypto.HexToECDSA("b44a49889a79983336d15385161533868644d35c1ea670854a0a0b4b784ae40c")
			Ω(err).ShouldNot(HaveOccurred())
			auth := bind.NewKeyedTransactor(testKey)

			//Create newHyperdriveContract for sending Txs
			hyper, err := NewHyperdriveContract(context.Background(), conn, auth, &bind.CallOpts{})
			Ω(err).ShouldNot(HaveOccurred())

			number, err := hyper.GetBlockNumberOfTx(common.HexToHash("0xa587c6e316d865b8f6bbda1e18be32f35aab831ae09493d19ca81c3b7be51889"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(number).Should(Equal(uint64(3097146)))

			block, err := hyper.CurrentBlock()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(block.NumberU64()).Should(BeNumerically(">", 3097168))
		})
	})
})
