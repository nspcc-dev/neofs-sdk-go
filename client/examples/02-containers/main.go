package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nspcc-dev/neo-go/cli/flags"
	"github.com/nspcc-dev/neo-go/cli/input"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/container"
)

const usage = `Container management example

$ ./02-containers -wallet [..] -address [..] list 
$ ./02-containers -wallet [..] -address [..] get [container-id] 
$ ./02-containers -wallet [..] -address [..] create 

`

var (
	walletPath = flag.String("wallet", "", "path to JSON wallet file")
	walletAddr = flag.String("address", "", "wallet address [optional]")
)

func main() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	ctx := context.Background()

	// First obtain client credentials: private key of request owner.
	key, err := getCredentials(*walletPath, *walletAddr)
	if err != nil {
		log.Fatal("can't read credentials:", err)
	}

	cli, err := client.New(
		// provide private key associated with request owner
		client.WithDefaultPrivateKey(key),
		// find endpoints in https://testcdn.fs.neo.org/doc/integrations/endpoints/
		client.WithURIAddress("grpcs://st01.testnet.fs.neo.org:8082", nil),
		// check client errors in go compatible way
		client.WithNeoFSErrorParsing(),
	)
	if err != nil {
		log.Fatal("can't create NeoFS client:", err)
	}

	switch flag.Arg(0) {
	case "list":
		ids, err := List(ctx, cli, key)
		if err != nil {
			log.Fatal(err)
		}

		for _, id := range ids {
			fmt.Println(id)
		}
	case "get":
		cnr, err := Get(ctx, cli, flag.Arg(1))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Container ID: ", container.CalculateID(cnr))
		for _, attribute := range cnr.Attributes() {
			fmt.Printf("Attribute %s: %s\n", attribute.Key(), attribute.Value())
		}
	case "create":
		id, err := Create(ctx, cli, key)
		if err != nil {
			log.Fatal(err)
		}

		// Poll container ID until it will be available in the network.
		for i := 0; i <= 30; i++ {
			if i == 30 {
				log.Fatalf("Timeout, container %s was not persisted in side chain\n", id)
			}
			_, err = Get(ctx, cli, id.String())
			if err == nil {
				break
			}
			time.Sleep(time.Second)
		}

		fmt.Printf("Container %s has been persisted in side chain\n", id)
	default:
		log.Fatal("unknown command", flag.Arg(0))
	}
}

func getCredentials(path, address string) (*ecdsa.PrivateKey, error) {
	w, err := wallet.NewWalletFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("can't read the wallet: %w", err)
	}

	return getKeyFromWallet(w, address)
}

// getKeyFromWallet fetches private key from neo-go wallet structure
func getKeyFromWallet(w *wallet.Wallet, addrStr string) (*ecdsa.PrivateKey, error) {
	var (
		addr util.Uint160
		err  error
	)

	if addrStr == "" {
		addr = w.GetChangeAddress()
	} else {
		addr, err = flags.ParseAddress(addrStr)
		if err != nil {
			return nil, fmt.Errorf("invalid wallet address %s: %w", addrStr, err)
		}
	}

	acc := w.GetAccount(addr)
	if acc == nil {
		return nil, fmt.Errorf("invalid wallet address %s: %w", addrStr, err)
	}

	pass, err := input.ReadPassword("Enter password > ")
	if err != nil {
		return nil, errors.New("invalid password")
	}

	if err := acc.Decrypt(pass, keys.NEP2ScryptParams()); err != nil {
		return nil, errors.New("invalid password")

	}

	return &acc.PrivateKey().PrivateKey, nil
}
