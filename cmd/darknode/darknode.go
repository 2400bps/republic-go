package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/republicprotocol/republic-go/blockchain/ethereum"
	"github.com/republicprotocol/republic-go/blockchain/ethereum/accounts"
	"github.com/republicprotocol/republic-go/blockchain/ethereum/dnr"
	"github.com/republicprotocol/republic-go/blockchain/ethereum/ledger"
	"github.com/republicprotocol/republic-go/cal"
	"github.com/republicprotocol/republic-go/cmd/darknode/config"
	"github.com/republicprotocol/republic-go/crypto"
	"github.com/republicprotocol/republic-go/dht"
	"github.com/republicprotocol/republic-go/dispatch"
	"github.com/republicprotocol/republic-go/grpc"
	"github.com/republicprotocol/republic-go/identity"
	"github.com/republicprotocol/republic-go/leveldb"
	"github.com/republicprotocol/republic-go/logger"
	"github.com/republicprotocol/republic-go/ome"
	"github.com/republicprotocol/republic-go/orderbook"
	"github.com/republicprotocol/republic-go/registry"
	"github.com/republicprotocol/republic-go/smpc"
	"github.com/republicprotocol/republic-go/stream"
	"github.com/republicprotocol/republic-go/swarm"
)

func main() {
	done := make(chan struct{})
	defer close(done)

	logger.SetFilterLevel(logger.LevelDebugLow)

	// Parse command-line arguments
	configParam := flag.String("config", path.Join(os.Getenv("HOME"), ".darknode/config.json"), "JSON configuration file")
	dataParam := flag.String("data", path.Join(os.Getenv("HOME"), ".darknode/data"), "Data directory")
	flag.Parse()

	// Load configuration file
	config, err := config.NewConfigFromJSONFile(*configParam)
	if err != nil {
		log.Fatalf("cannot load config: %v", err)
	}

	// Get IP-address
	ipAddr, err := getIPAddress()
	if err != nil {
		log.Fatalf("cannot get ip-address: %v", err)
	}

	// Get multi-address
	multiAddr, err := identity.NewMultiAddressFromString(fmt.Sprintf("/ip4/%s/tcp/18514/republic/%s", ipAddr, config.Address))
	if err != nil {
		log.Fatalf("cannot get multiaddress: %v", err)
	}
	log.Printf("address %v", multiAddr)

	// Get ethereum bindings
	auth, darkPool, darkPoolAccounts, _, renLedger, err := getEthereumBindings(config.Keystore, config.Ethereum)
	if err != nil {
		log.Fatalf("cannot get ethereum bindings: %v", err)
	}
	log.Printf("ethereum %v", auth.From.Hex())

	// New crypter for signing and verification
	crypter := registry.NewCrypter(config.Keystore, darkPool, 256, time.Minute)

	// New database for persistent storage
	store, err := leveldb.NewStore(*dataParam)
	if err != nil {
		log.Fatalf("cannot open leveldb: %v", err)
	}
	defer store.Close()

	// New DHT
	dht := dht.NewDHT(config.Address, 64)

	// New gRPC components
	server := grpc.NewServer()

	statusService := grpc.NewStatusService(&dht)
	statusService.Register(server)

	swarmClient := grpc.NewSwarmClient(multiAddr)
	swarmService := grpc.NewSwarmService(swarm.NewServer(&crypter, swarmClient, &dht))
	swarmer := swarm.NewSwarmer(swarmClient, &dht)
	swarmService.Register(server)

	orderbook := orderbook.NewOrderbook(config.Keystore.RsaKey, orderbook.NewSyncer(renLedger, 32), &store)
	orderbookService := grpc.NewOrderbookService(orderbook)
	orderbookService.Register(server)

	streamClient := grpc.NewStreamClient(&crypter, config.Address)
	streamService := grpc.NewStreamService(&crypter, config.Address)
	streamer := stream.NewStreamRecycler(stream.NewStreamer(config.Address, streamClient, &streamService))
	streamService.Register(server)

	// Start the secure order matching engine
	go func() {
		// Wait for the gRPC server to boot
		time.Sleep(time.Second)

		// FIXME: Wait until registration has been approved.

		// Bootstrap into the network
		fmtStr := "bootstrapping\n"
		for _, multiAddr := range config.BootstrapMultiAddresses {
			fmtStr += "  " + multiAddr.String() + "\n"
		}
		log.Printf(fmtStr)
		if err := swarmer.Bootstrap(context.Background(), config.BootstrapMultiAddresses); err != nil {
			log.Printf("bootstrap: %v", err)
		}
		log.Printf("connected to %v peers", len(dht.MultiAddresses()))

		// New secure multi-party computer
		smpcer := smpc.NewSmpcer(swarmer, streamer)

		// New OME
		epoch, err := darkPool.Epoch()
		if err != nil {
			log.Fatalf("cannot get current epoch: %v", err)
		}
		ranker, err := ome.NewRanker(done, config.Address, epoch)
		if err != nil {
			log.Fatalf("cannot create new ranker: %v", err)
		}
		matcher := ome.NewMatcher(&store, smpcer)
		confirmer := ome.NewConfirmer(&store, renLedger, 14*time.Second, 1)
		settler := ome.NewSettler(&store, smpcer, darkPoolAccounts)
		ome := ome.NewOme(config.Address, ranker, matcher, confirmer, settler, &store, orderbook, smpcer, epoch)

		dispatch.CoBegin(func() {
			// Synchronizing the OME
			errs := ome.Run(done)
			for err := range errs {
				logger.Error(fmt.Sprintf("error in running the ome: %v", err))
			}
		}, func() {

			// Periodically sync the next ξ
			for {
				time.Sleep(14 * time.Second)

				// Get the epoch
				nextEpoch, err := darkPool.Epoch()
				if err != nil {
					logger.Error(fmt.Sprintf("cannot sync epoch: %v", err))
					continue
				}

				// Check whether or not ξ has changed
				if epoch.Equal(&nextEpoch) {
					continue
				}
				epoch = nextEpoch
				logger.Epoch(epoch.Hash)

				// Notify the Ome
				ome.OnChangeEpoch(epoch)
			}
		})
	}()

	// Start gRPC server and run until the server is stopped
	log.Printf("listening on %v:%v...", config.Host, config.Port)
	lis, err := net.Listen("tcp", fmt.Sprintf("%v:%v", config.Host, config.Port))
	if err != nil {
		log.Fatalf("cannot listen on %v:%v: %v", config.Host, config.Port, err)
	}
	if err := server.Serve(lis); err != nil {
		log.Fatalf("cannot serve on %v:%v: %v", config.Host, config.Port, err)
	}
}

func getIPAddress() (string, error) {

	out, err := exec.Command("curl", "https://ipinfo.io/ip").Output()
	if err != nil {
		return "", err
	}
	out = []byte(strings.Trim(string(out), "\n "))
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func getEthereumBindings(keystore crypto.Keystore, conf ethereum.Config) (*bind.TransactOpts, cal.Darkpool, cal.DarkpoolAccounts, cal.DarkpoolFees, cal.RenLedger, error) {
	conn, err := ethereum.Connect(conf)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("cannot connect to ethereum: %v", err)
	}
	auth := bind.NewKeyedTransactor(keystore.EcdsaKey.PrivateKey)
	auth.GasPrice = big.NewInt(1000000000)

	darkpool, err := dnr.NewDarknodeRegistry(context.Background(), conn, auth, &bind.CallOpts{})
	if err != nil {
		fmt.Println(fmt.Errorf("cannot bind to darkpool: %v", err))
		return auth, nil, nil, nil, nil, err
	}

	renLedger, err := ledger.NewRenLedgerContract(context.Background(), conn, auth, &bind.CallOpts{})
	if err != nil {
		fmt.Println(fmt.Errorf("cannot bind to ren ledger: %v", err))
		return auth, nil, nil, nil, nil, err
	}

	acts, err := accounts.NewRenExAccounts(context.Background(), conn, auth, &bind.CallOpts{})
	if err != nil {
		fmt.Println(fmt.Errorf("cannot bind to RenEx accounts: %v", err))
		return auth, nil, nil, nil, nil, err
	}

	return auth, &darkpool, &acts, nil, &renLedger, nil
}
