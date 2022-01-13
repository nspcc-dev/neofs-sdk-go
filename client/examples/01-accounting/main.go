package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/nspcc-dev/neo-go/cli/flags"
	"github.com/nspcc-dev/neo-go/cli/input"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
)

const usage = `NeoFS Balance requests

$ ./01-accounting -wallet [..] -address [..]

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

	// First obtain client credentials: private key of request owner
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

	result, err := cli.GetBalance(ctx, ownerIDFromPrivateKey(key))
	if err != nil {
		log.Fatal("can't get NeoFS Balance:", err)
	}

	fmt.Println("value:", result.Amount().Value())
	fmt.Println("precision:", result.Amount().Precision())
}

func getCredentials(path, address string) (*ecdsa.PrivateKey, error) {
	w, err := wallet.NewWalletFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("can't read the wallet: %walletPath", err)
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

func ownerIDFromPrivateKey(key *ecdsa.PrivateKey) *owner.ID {
	w, err := owner.NEO3WalletFromPublicKey(&key.PublicKey)
	if err != nil {
		panic(fmt.Errorf("invalid private key"))
	}

	return owner.NewIDFromNeo3Wallet(w)
}
