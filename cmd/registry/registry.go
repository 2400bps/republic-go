package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jbenet/go-base58"
	"github.com/republicprotocol/republic-go/blockchain/ethereum"
	"github.com/republicprotocol/republic-go/blockchain/ethereum/dnr"
	"github.com/republicprotocol/republic-go/darkocean"
	"github.com/republicprotocol/republic-go/stackint"
	"github.com/urfave/cli"
)

const (
	reset  = "\x1b[0m"
	yellow = "\x1b[33;1m"
	green  = "\x1b[32;1m"
	red    = "\x1b[31;1m"
)

// Registry command-line tool for interacting with the darknodeRegister contract
// on Ropsten testnet.
// Set up ren contract address:
//   $ registry --ren 0xContractAddress
// Set up dnr contract address:
//   $ registry --dnr 0xContractAddress
// Register nodes:
//   $ registry register 0xaddress1 0xaddress2 0xaddress3
// Deregister nodes:
//   $ registry deregister 0xaddress1 0xaddress2 0xaddress3
// Calling epoch:
//   $ registry epoch

func main() {

	// Load ethereum key
	key, err := LoadKey()
	if err != nil {
		log.Fatal("failt to load key from file", err)
	}

	// Create new cli application
	app := cli.NewApp()

	// Define flags
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "ren",
			Value: "0x65d54eda5f032f2275caa557e50c029cfbccbb54",
			Usage: "republic token contract address",
		},
		cli.StringFlag{
			Name:  "dnr",
			Value: "0x69eb8d26157b9e12f959ea9f189A5D75991b59e3",
			Usage: "dark node registry address",
		},
	}

	// Define subcommands
	app.Commands = []cli.Command{
		{
			Name:    "epoch",
			Aliases: []string{"e"},
			Usage:   "calling epoch",
			Action: func(c *cli.Context) error {
				registry, err := NewRegistry(c, key)
				if err != nil {
					return err
				}
				_, err = registry.Epoch()
				log.Println("Epoch called.")
				return err
			},
		},
		{
			Name:    "checkreg",
			Aliases: []string{"c"},
			Usage:   "check if the node is registered or not",
			Action: func(c *cli.Context) error {
				registrar, err := NewRegistry(c, key)
				if err != nil {
					return err
				}
				return CheckRegistration(c.Args(), registrar)
			},
		},
		{
			Name:    "register",
			Aliases: []string{"r"},
			Usage:   "register nodes in the dark node registry",
			Action: func(c *cli.Context) error {
				registry, err := NewRegistry(c, key)
				if err != nil {
					return err
				}
				return RegisterAll(c.Args(), registry)
			},
		},
		{
			Name:    "approve",
			Aliases: []string{"a"},
			Usage:   "approve nodes with enough REN token",
			Action: func(c *cli.Context) error {
				registry, err := NewRegistry(c, key)
				if err != nil {
					return err
				}
				return Approve(registry)
			},
		},
		{
			Name:    "deregister",
			Aliases: []string{"d"},
			Usage:   "deregister nodes in the dark node registry",
			Action: func(c *cli.Context) error {
				registry, err := NewRegistry(c, key)
				if err != nil {
					return err
				}
				return DeregisterAll(c.Args(), registry)
			},
		},
		{
			Name:  "refund",
			Usage: "refund ren",
			Action: func(c *cli.Context) error {
				registry, err := NewRegistry(c, key)
				if err != nil {
					return err
				}
				return Refund(c.Args(), registry)
			},
		},
		{
			Name:    "pool",
			Aliases: []string{"p"},
			Usage:   "get the index of the pool the node is in, return -1 if no pool found",
			Action: func(c *cli.Context) error {
				registrar, err := NewRegistry(c, key)
				if err != nil {
					return err
				}
				return GetPool(c.Args(), registrar)
			},
		},
	}
	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func LoadKey() (*keystore.Key, error) {
	var keyJSON string = `{"address":"0066ed1af055a568e49e5a20f5b63e9741c81967",
  					"privatekey":"f5ec4b010d7cc3fbf75d71f3ea4700a64bb091e05531ef65424415207c661a39",
  					"id":"06a15d42-1d52-42dd-b714-82582a13782a",
  					"version":3}`
	key := new(keystore.Key)
	err := key.UnmarshalJSON([]byte(keyJSON))

	return key, err
}

func NewRegistry(c *cli.Context, key *keystore.Key) (dnr.DarknodeRegistry, error) {
	config := ethereum.Config{
		Network:                 ethereum.NetworkRopsten,
		URI:                     "https://ropsten.infura.io",
		RepublicTokenAddress:    c.String("ren"),
		DarknodeRegistryAddress: c.String("dnr"),
	}

	auth := bind.NewKeyedTransactor(key.PrivateKey)
	auth.GasPrice = big.NewInt(40000000000)
	client, err := ethereum.Connect(config)
	if err != nil {
		log.Fatal("fail to connect to ethereum")
	}

	return dnr.NewDarknodeRegistry(context.Background(), client, auth, &bind.CallOpts{})
}

func RegisterAll(addresses []string, registry dnr.DarknodeRegistry) error {
	for i := range addresses {
		address, err := republicAddressToEthAddress(addresses[i])
		if err != nil {
			return err
		}

		// Check if node has already been registered
		isRegistered, err := registry.IsRegistered(address.Bytes())
		if err != nil {
			return fmt.Errorf("[%v] %sCouldn't check node's registration%s: %v\n", []byte(addresses[i]), red, reset, err)
		}

		// Register the node if not registered
		if !isRegistered {
			minimumBond, err := registry.MinimumBond()
			if err != nil {
				return err
			}

			_, err = registry.Register(address.Bytes(), []byte{}, &minimumBond)
			if err != nil {
				return fmt.Errorf("[%v] %sCouldn't register node%s: %v\n", address.Hex(), red, reset, err)
			} else {
				return fmt.Errorf("[%v] %sNode will be registered next epoch%s\n", address.Hex(), green, reset)
			}
		} else if isRegistered {
			log.Printf("[%v] %sNode already registered%s\n", address.Hex(), yellow, reset)
		}
	}

	return nil
}

// DeregisterAll takes a slice of republic private keys and registers them
func DeregisterAll(addresses []string, registry dnr.DarknodeRegistry) error {
	for i := range addresses {
		address, err := republicAddressToEthAddress(addresses[i])
		if err != nil {
			return err
		}

		// Check if node has already been registered
		isRegistered, err := registry.IsRegistered(address.Bytes())
		if err != nil {
			return fmt.Errorf("[%v] %sCouldn't check node's registration%s: %v\n", address.Hex(), red, reset, err)
		}

		if isRegistered {
			_, err = registry.Deregister(address.Bytes())
			if err != nil {
				return fmt.Errorf("[%v] %sCouldn't deregister node%s: %v\n", address.Hex(), red, reset, err)
			} else {
				log.Printf("[%v] %sNode will be deregistered next epoch%s\n", address.Hex(), green, reset)
			}
		} else {
			fmt.Println(fmt.Errorf("[%v] %sNode hasn't been registered yet.%s\n", address.Hex(), red, reset))
		}
	}

	return nil
}

func Approve(registry dnr.DarknodeRegistry) error {

	bond, err := stackint.FromString("100000000000000000000000")
	if err != nil {
		return err
	}
	_, err = registry.ApproveRen(&bond)
	if err != nil {
		return err
	}

	return nil
}

// GetPool will get the index of the pool the node is in.
// The address should be the ethereum address
func GetPool(addresses []string, registry dnr.DarknodeRegistry) error {
	if len(addresses) != 1 {
		return fmt.Errorf("%sPlease provide one node address.%s\n", red, reset)
	}
	address, err := republicAddressToEthAddress(addresses[0])
	if err != nil {
		return err
	}

	currentEpoch, err := registry.CurrentEpoch()
	if err != nil {
		return err
	}
	nodes, err := registry.GetAllNodes()
	if err != nil {
		return err
	}

	ocean, err := darkocean.NewDarkOcean(&registry, currentEpoch.Blockhash, nodes)
	if err != nil {
		return err
	}
	poolIndex := ocean.PoolIndex(address.Bytes())
	fmt.Println(poolIndex)

	return nil
}

// CheckRegistration will check if the node with given address is registered with
// the darknode registry. The address will be the ethereum address.
func CheckRegistration(addresses []string, registrar dnr.DarknodeRegistry) error {
	if len(addresses) != 1 {
		return fmt.Errorf("%sPlease provide one node address.%s\n", red, reset)
	}
	address, err := republicAddressToEthAddress(addresses[0])
	if err != nil {
		return err
	}

	isRegistered, err := registrar.IsRegistered(address.Bytes())
	if err != nil {
		return err
	}
	fmt.Println(isRegistered)

	return nil
}

func Refund(addresses []string, registry dnr.DarknodeRegistry) error {
	for i := range addresses {
		address, err := republicAddressToEthAddress(addresses[i])
		if err != nil {
			return err
		}
		_, err = registry.Refund(address.Bytes())
		if err != nil {
			return err
		}
		log.Printf("[%v] %sNode has been refunded%s\n", address.Hex(), green, reset)
	}

	return nil
}

// Convert republic address to ethereum address
func republicAddressToEthAddress(repAddress string) (common.Address, error) {
	addByte := base58.DecodeAlphabet(repAddress, base58.BTCAlphabet)[2:]
	if len(addByte) == 0 {
		return common.Address{}, errors.New("fail to decode the address")
	}
	address := common.BytesToAddress(addByte)
	return address, nil
}
