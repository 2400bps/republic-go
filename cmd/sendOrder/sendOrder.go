package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	netHttp "net/http"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/republicprotocol/republic-go/blockchain/ethereum"
	"github.com/republicprotocol/republic-go/blockchain/ethereum/dnr"
	"github.com/republicprotocol/republic-go/cal"
	"github.com/republicprotocol/republic-go/crypto"
	"github.com/republicprotocol/republic-go/http"
	"github.com/republicprotocol/republic-go/http/adapter"
	"github.com/republicprotocol/republic-go/order"
)

func main() {
	relayParam := flag.String("relay", "https://kovan-renexchange.herokuapp.com", "Binding and port of the relay")
	keystoreParam := flag.String("keystore", "", "Optionally encrypted keystore file")
	configParam := flag.String("config", "", "Ethereum configuration file")
	passphraseParam := flag.String("passphrase", "", "Optional passphrase to decrypt the keystore file")

	flag.Parse()

	keystore, err := loadKeystore(*keystoreParam, *passphraseParam)
	if err != nil {
		log.Fatalf("cannot load keystore: %v", err)
	}

	config, err := loadConfig(*configParam)
	if err != nil {
		log.Fatalf("cannot load config: %v", err)
	}

	_, darkpool, err := loadSmartContracts(config, keystore)
	if err != nil {
		log.Fatalf("cannot load smart contracts: %v", err)
	}
	onePrice := order.CoExp{
		Co:  2,
		Exp: 40,
	}
	oneVol := order.CoExp{
		Co:  5,
		Exp: 12,
	}
	buy := order.NewOrder(order.TypeLimit, order.ParityBuy, order.SettlementRenEx, time.Now().Add(1*time.Hour), order.TokensDGXREN, onePrice, oneVol, oneVol, rand.Uint64())
	sell := order.NewOrder(order.TypeLimit, order.ParitySell, order.SettlementRenEx, time.Now().Add(1*time.Hour), order.TokensDGXREN, onePrice, oneVol, oneVol, rand.Uint64())
	ords := []order.Order{buy, sell}

	for _, ord := range ords {
		signatureData := crypto.Keccak256([]byte("Republic Protocol: open: "), ord.ID[:])
		signatureData = crypto.Keccak256([]byte("\x19Ethereum Signed Message:\n32"), signatureData)
		signature, err := keystore.Sign(signatureData)

		log.Printf("id = %v; sig = %v", base64.StdEncoding.EncodeToString(ord.ID[:]), base64.StdEncoding.EncodeToString(signature))

		if err != nil {
			log.Fatalf("cannot sign order: %v", err)
		}

		request := http.OpenOrderRequest{
			Signature:            base64.StdEncoding.EncodeToString(signature),
			OrderFragmentMapping: map[string][]adapter.OrderFragment{},
		}
		log.Printf("order signature %v", request.Signature)

		pods, err := darkpool.Pods()
		if err != nil {
			log.Fatalf("cannot get pods from darkpool: %v", err)
		}
		for _, pod := range pods {
			n := int64(len(pod.Darknodes))
			k := int64(2 * (len(pod.Darknodes) + 1) / 3)
			hash := base64.StdEncoding.EncodeToString(pod.Hash[:])
			ordFragments, err := ord.Split(n, k)
			if err != nil {
				log.Fatalf("cannot split order to %v: %v", hash, err)
			}
			request.OrderFragmentMapping[hash] = []adapter.OrderFragment{}
			for i, ordFragment := range ordFragments {
				marshaledOrdFragment := adapter.OrderFragment{
					Index: int64(i + 1),
				}

				pubKey, err := darkpool.PublicKey(pod.Darknodes[i])
				if err != nil {
					log.Fatalf("cannot get public key of %v: %v", pod.Darknodes[i], err)
				}

				encryptedFragment, err := ordFragment.Encrypt(pubKey)
				marshaledOrdFragment.ID = base64.StdEncoding.EncodeToString(encryptedFragment.ID[:])
				marshaledOrdFragment.OrderID = base64.StdEncoding.EncodeToString(encryptedFragment.OrderID[:])
				marshaledOrdFragment.OrderParity = encryptedFragment.OrderParity
				marshaledOrdFragment.OrderSettlement = encryptedFragment.OrderSettlement
				marshaledOrdFragment.OrderType = encryptedFragment.OrderType
				marshaledOrdFragment.OrderExpiry = encryptedFragment.OrderExpiry.Unix()
				marshaledOrdFragment.Tokens = base64.StdEncoding.EncodeToString(encryptedFragment.Tokens)
				marshaledOrdFragment.Price = []string{
					base64.StdEncoding.EncodeToString(encryptedFragment.Price.Co),
					base64.StdEncoding.EncodeToString(encryptedFragment.Price.Exp),
				}
				marshaledOrdFragment.Volume = []string{
					base64.StdEncoding.EncodeToString(encryptedFragment.Volume.Co),
					base64.StdEncoding.EncodeToString(encryptedFragment.Volume.Exp),
				}
				marshaledOrdFragment.MinimumVolume = []string{
					base64.StdEncoding.EncodeToString(encryptedFragment.MinimumVolume.Co),
					base64.StdEncoding.EncodeToString(encryptedFragment.MinimumVolume.Exp),
				}
				marshaledOrdFragment.Nonce = base64.StdEncoding.EncodeToString(encryptedFragment.Nonce)
				request.OrderFragmentMapping[hash] = append(request.OrderFragmentMapping[hash], marshaledOrdFragment)
			}
		}

		data, err := json.MarshalIndent(request, "", "  ")
		if err != nil {
			log.Fatalf("cannot marshal request: %v", err)
		}
		buf := bytes.NewBuffer(data)

		res, err := netHttp.DefaultClient.Post(fmt.Sprintf("%v/orders", *relayParam), "application/json", buf)
		if err != nil {
			log.Fatalf("cannot send request: %v", err)
		}
		defer res.Body.Close()

		log.Printf("status: %v", res.StatusCode)
		resText, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("cannot read response body: %v", err)
		}
		log.Printf("body: %v", string(resText))
	}
}

func loadKeystore(keystoreFile, passphrase string) (crypto.Keystore, error) {
	file, err := os.Open(keystoreFile)
	if err != nil {
		return crypto.Keystore{}, err
	}
	defer file.Close()

	if passphrase == "" {
		keystore := crypto.Keystore{}
		if err := json.NewDecoder(file).Decode(&keystore); err != nil {
			return keystore, err
		}
		return keystore, nil
	}

	keystore := crypto.Keystore{}
	keystoreData, err := ioutil.ReadAll(file)
	if err != nil {
		return keystore, err
	}
	if err := keystore.DecryptFromJSON(keystoreData, passphrase); err != nil {
		return keystore, err
	}
	return keystore, nil
}

func loadConfig(configFile string) (ethereum.Config, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return ethereum.Config{}, err
	}
	defer file.Close()
	config := ethereum.Config{}
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return ethereum.Config{}, err
	}
	return config, nil
}

func loadSmartContracts(ethereumConfig ethereum.Config, keystore crypto.Keystore) (*bind.TransactOpts, cal.Darkpool, error) {
	conn, err := ethereum.Connect(ethereumConfig)
	if err != nil {
		fmt.Println(fmt.Errorf("cannot connect to ethereum: %v", err))
		return nil, nil, err
	}
	auth := bind.NewKeyedTransactor(keystore.EcdsaKey.PrivateKey)

	registry, err := dnr.NewDarknodeRegistry(context.Background(), conn, auth, &bind.CallOpts{})
	if err != nil {
		fmt.Println(fmt.Errorf("cannot bind to darkpool: %v", err))
		return auth, nil, err
	}
	return auth, &registry, nil
}
