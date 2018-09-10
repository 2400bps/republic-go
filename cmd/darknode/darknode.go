package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	netHttp "net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/republicprotocol/republic-go/cmd/darknode/config"
	"github.com/republicprotocol/republic-go/contract"
	"github.com/republicprotocol/republic-go/crypto"
	"github.com/republicprotocol/republic-go/dispatch"
	"github.com/republicprotocol/republic-go/grpc"
	"github.com/republicprotocol/republic-go/http"
	"github.com/republicprotocol/republic-go/http/adapter"
	"github.com/republicprotocol/republic-go/identity"
	"github.com/republicprotocol/republic-go/leveldb"
	"github.com/republicprotocol/republic-go/logger"
	"github.com/republicprotocol/republic-go/ome"
	"github.com/republicprotocol/republic-go/oracle"
	"github.com/republicprotocol/republic-go/orderbook"
	"github.com/republicprotocol/republic-go/registry"
	"github.com/republicprotocol/republic-go/smpc"
	"github.com/republicprotocol/republic-go/status"
	"github.com/republicprotocol/republic-go/swarm"
	"golang.org/x/time/rate"
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

	// Connect to Ethereum
	conn, err := contract.Connect(config.Ethereum)
	if err != nil {
		log.Fatalf("cannot connect to ethereum: %v", err)
	}

	auth := bind.NewKeyedTransactor(config.Keystore.EcdsaKey.PrivateKey)

	// Get ethereum bindings
	contractBinder, err := contract.NewBinder(auth, conn)
	if err != nil {
		log.Fatalf("cannot get ethereum bindings: %v", err)
	}

	// New database for persistent storage
	store, err := leveldb.NewStore(*dataParam, 25*time.Hour, time.Hour)
	if err != nil {
		log.Fatalf("cannot open leveldb: %v", err)
	}
	defer store.Release()
	store.Prune()

	midpointPriceStorer := leveldb.NewMidpointPriceStorer()

	// New crypter for signing and verification
	crypter := registry.NewCrypter(config.Keystore, &contractBinder, 256, time.Minute)
	updateOwnAddress := func() {
		signature, err := crypter.Sign(multiAddr.Hash())
		if err != nil {
			log.Fatalf("cannot sign own multiAddress: %v", err)
		}
		multiAddr.Signature = signature
		if err := store.SwarmMultiAddressStore().InsertMultiAddress(multiAddr); err != nil {
			log.Fatalf("cannot store own multiAddress in leveldb: %v", err)
		}
	}

	// Insert self multiAddress if not in the storer.
	oldMulti, err := store.SwarmMultiAddressStore().MultiAddress(multiAddr.Address())
	if err != nil {
		if err == swarm.ErrMultiAddressNotFound {
			updateOwnAddress()
		} else {
			logger.Network(logger.LevelError, fmt.Sprintf("error retrieving own multiAddress from store: %v", err))
		}
	} else {
		// Update own multiAddress if the ip address has been changed.
		oldIP, err := oldMulti.ValueForProtocol(identity.IP4Code)
		if err != nil {
			log.Fatalf("cannot read own ip address from storer: %v", err)
		}
		if oldIP != ipAddr {
			updateOwnAddress()
		}
	}

	// New gRPC components
	unaryLimiter := grpc.NewRateLimiter(rate.NewLimiter(20, 40), 5, 50)
	streamLimiter := grpc.NewRateLimiter(rate.NewLimiter(40, 80), 4.0, 20)
	server := grpc.NewServerwithLimiter(unaryLimiter, streamLimiter)

	swarmClient := grpc.NewSwarmClient(store.SwarmMultiAddressStore(), multiAddr.Address())
	swarmer := swarm.NewSwarmer(swarmClient, store.SwarmMultiAddressStore(), config.Alpha, &crypter)
	swarmService := grpc.NewSwarmService(swarm.NewServer(swarmer, store.SwarmMultiAddressStore(), config.Alpha, &crypter))
	swarmService.Register(server)

	oracleClient := grpc.NewOracleClient(multiAddr.Address(), store.SwarmMultiAddressStore())
	oracler := oracle.NewOracler(oracleClient, &config.Keystore.EcdsaKey, store.SwarmMultiAddressStore(), config.Alpha)
	oracleService := grpc.NewOracleService(oracle.NewServer(oracler, config.OracleAddress, store.SwarmMultiAddressStore(), midpointPriceStorer, config.Alpha), time.Millisecond)
	oracleService.Register(server)

	orderbook := orderbook.NewOrderbook(config.Address, config.Keystore.RsaKey, store.OrderbookPointerStore(), store.OrderbookOrderStore(), store.OrderbookOrderFragmentStore(), &contractBinder, 5*time.Second, 32)
	orderbookService := grpc.NewOrderbookService(orderbook)
	orderbookService.Register(server)

	connectorListener := grpc.NewConnectorListener(config.Address, &crypter, &crypter)
	streamerService := grpc.NewStreamerService(config.Address, &crypter, &crypter, connectorListener.Listener)
	streamerService.Register(server)

	var ethNetwork string
	if config.Ethereum.Network == "mainnet" {
		ethNetwork = "mainnet"
	} else {
		ethNetwork = "kovan"
	}

	// Populate status information
	statusProvider := status.NewProvider(swarmer)
	statusProvider.WriteNetwork(string(conn.Config.Network))
	statusProvider.WriteMultiAddress(multiAddr)
	statusProvider.WriteEthereumNetwork(ethNetwork)
	statusProvider.WriteEthereumAddress(auth.From.Hex())
	statusProvider.WriteDarknodeRegistryAddress(conn.Config.DarknodeRegistryAddress)
	statusProvider.WriteRewardVaultAddress(conn.Config.DarknodeRewardVaultAddress)
	statusProvider.WriteInfuraURL(conn.Config.URI)
	statusProvider.WriteTokens(contract.TokenAddresses(conn.Config.Network))

	pk, err := crypto.BytesFromRsaPublicKey(&config.Keystore.RsaKey.PublicKey)
	if err != nil {
		log.Fatalf("could not determine public key: %v", err)
	}
	statusProvider.WritePublicKey(pk)

	// Start the status server
	go func() {
		bindParam := "0.0.0.0"
		portParam := "18515"
		log.Printf("HTTP listening on %v:%v...", bindParam, portParam)

		statusAdapter := adapter.NewStatusAdapter(statusProvider)
		if err := netHttp.ListenAndServe(fmt.Sprintf("%v:%v", bindParam, portParam), http.NewStatusServer(statusAdapter)); err != nil {
			log.Fatalf("error listening and serving: %v", err)
		}
	}()

	// Start the secure order matching engine
	go func() {
		// Wait for the gRPC server to boot
		time.Sleep(time.Second)
		rand.Seed(time.Now().UnixNano())

		// Wait until registration
		isRegistered, err := contractBinder.IsRegistered(config.Address)
		if err != nil {
			logger.Network(logger.LevelError, fmt.Sprintf("cannot get registration status: %v", err))
		}
		for !isRegistered {
			time.Sleep(10 * time.Second)
			isRegistered, err = contractBinder.IsRegistered(config.Address)
			if err != nil {
				logger.Network(logger.LevelError, fmt.Sprintf("cannot get registration status: %v", err))
			}
		}

		// Bootstrap into the network
		fmtStr := "bootstrapping\n"
		for _, bootstrapMulti := range config.BootstrapMultiAddresses {
			fmtStr += "  " + bootstrapMulti.String() + "\n"
			oldBootstrapAddr, err := store.SwarmMultiAddressStore().MultiAddress(bootstrapMulti.Address())
			if err != nil {
				if err == swarm.ErrMultiAddressNotFound {
					if err := store.SwarmMultiAddressStore().InsertMultiAddress(bootstrapMulti); err != nil {
						logger.Network(logger.LevelError, fmt.Sprintf("cannot store bootstrap multiaddress in store: %v", err))
					}
				} else {
					logger.Network(logger.LevelError, fmt.Sprintf("cannot get bootstrap multi-address from store: %v", err))
				}
			} else {
				// Update bootstrap multiAddress if the ip address has been changed.
				oldBootstrapIP, err := oldBootstrapAddr.ValueForProtocol(identity.IP4Code)
				if err != nil {
					log.Printf("cannot read own ip address from storer: %v", err)
					continue
				}
				newBootstrapIP, err := bootstrapMulti.ValueForProtocol(identity.IP4Code)
				if err != nil {
					log.Printf("cannot read own ip address from storer: %v", err)
					continue
				}
				if oldBootstrapIP != newBootstrapIP {
					if err := store.SwarmMultiAddressStore().InsertMultiAddress(bootstrapMulti); err != nil {
						logger.Network(logger.LevelError, fmt.Sprintf("cannot store bootstrap multiaddress in store: %v", err))
					}
				}
			}
		}
		log.Printf(fmtStr)
		if err := pingNetwork(swarmer); err != nil {
			log.Fatalf("[error] (bootstrap) cannot ping network: %v", err)
		}

		// New secure multi-party computer
		smpcer := smpc.NewSmpcer(connectorListener, swarmer)

		// New OME
		epoch, err := contractBinder.PreviousEpoch()
		if err != nil {
			logger.Error(fmt.Sprintf("cannot get previous epoch: %v", err))
		}
		gen := ome.NewComputationGenerator(config.Address, store.SomerOrderFragmentStore())
		matcher := ome.NewMatcher(store.SomerComputationStore(), smpcer)
		confirmer := ome.NewConfirmer(store.SomerComputationStore(), &contractBinder, 5*time.Second, 4)
		settler := ome.NewSettler(store.SomerComputationStore(), smpcer, &contractBinder, 1e12)
		ome := ome.NewOme(config.Address, gen, matcher, confirmer, settler, store.SomerComputationStore(), orderbook, smpcer, epoch)

		dispatch.CoBegin(func() {
			// Synchronizing the OME
			errs := ome.Run(done)
			for err := range errs {
				logger.Error(fmt.Sprintf("error in running the ome: %v", err))
			}
		}, func() {
			// Periodically sync the next ξ
			for {
				time.Sleep(5 * time.Second)

				// Get the epoch
				nextEpoch, err := contractBinder.Epoch()
				if err != nil {
					logger.Error(fmt.Sprintf("cannot sync epoch: %v", err))
					continue
				}

				// Check whether or not ξ has changed
				if nextEpoch.Equal(&epoch) {
					continue
				}
				epoch = nextEpoch
				logger.Epoch(epoch.Hash)

				// Notify the Ome
				ome.OnChangeEpoch(epoch)
			}
		}, func() {
			// Prune the database every hour and update the network with the
			// darknode address
			for {
				time.Sleep(time.Hour)
				if err := pingNetwork(swarmer); err != nil {
					log.Printf("[error] (prune) cannot ping network: %v", err)
					continue
				}
				if err := store.Prune(); err != nil {
					log.Printf("[error] (prune) cannot prune the storer: %v", err)
					continue
				}
			}
		})
	}()

	// Start gRPC server and run until the server is stopped
	log.Printf("gRPC listening on %v:%v...", config.Host, config.Port)
	lis, err := net.Listen("tcp", fmt.Sprintf("%v:%v", config.Host, config.Port))
	if err != nil {
		log.Fatalf("cannot listen on %v:%v: %v", config.Host, config.Port, err)
	}
	if err := server.Serve(lis); err != nil {
		log.Fatalf("cannot serve on %v:%v: %v", config.Host, config.Port, err)
	}
}

func getIPAddress() (string, error) {
	out, err := exec.Command("dig", "+short", "myip.opendns.com", "@resolver1.opendns.com").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// pingNetwork start ping the entire network with a new multiAddress with an
// updated nonce
func pingNetwork(swarmer swarm.Swarmer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := swarmer.Ping(ctx); err != nil {
		return err
	}
	peers, err := swarmer.Peers()
	if err != nil {
		return err
	}
	log.Printf("[info] connected to %v peers", len(peers)-1)

	return nil
}
