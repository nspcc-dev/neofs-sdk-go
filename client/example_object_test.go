package client_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// ExampleClient_ObjectGetInit demonstrates how to initialize an object retrieval
// operation using the NeoFS SDK for Go.
func ExampleClient_ObjectGetInit() {
	const (
		// walletPath is path to your wallet file.
		walletPath = "wallet.json"
		// walletPass is password to decrypt your wallet.
		walletPass = "password"

		// testnetStorageNode is address of NeoFS Testnet storage node.
		// It can be your own node or any public node.
		testnetStorageNode = "st1.t5.fs.neo.org:8080"

		// Example container and object IDs for testing.
		cnrIDStr    = "9wqwAcEU8Vn2tq4T1PVpBT6CbVgFb8WfF8ekSRmcQQjq"
		objectIDStr = "7sTMgE8MzpyvqwLh6EJUXdoPJJyMRRRNHyTx2ErvjgNX"
	)

	ctx := context.Background()

	// Load wallet and decrypt account
	// For public-read containers, this step is not required
	// as no signing is needed to read objects.
	w, err := wallet.NewWalletFromFile(walletPath)
	if err != nil {
		log.Fatal(err)
	}
	addr := w.GetChangeAddress()
	acc := w.GetAccount(addr)
	err = acc.Decrypt(walletPass, w.Scrypt)
	if err != nil {
		log.Fatal(err)
	}
	signer := user.NewAutoIDSigner(acc.PrivateKey().PrivateKey)

	// Create and configure client
	var prmInit client.PrmInit
	c, err := client.New(prmInit)
	if err != nil {
		log.Fatal(fmt.Errorf("client init: %w", err))
	}

	// Connect to NeoFS node
	var prmDial client.PrmDial
	prmDial.SetServerURI(testnetStorageNode)
	// Optional settings
	prmDial.SetTimeout(15 * time.Second)
	prmDial.SetStreamTimeout(15 * time.Second)

	if err = c.Dial(prmDial); err != nil {
		log.Fatal(fmt.Errorf("dial: %w", err))
	}

	// Parse container ID from string
	containerID, err := cid.DecodeString(cnrIDStr)
	if err != nil {
		log.Fatal(fmt.Errorf("parse container ID: %w", err))
	}

	// Parse object ID from string
	objectID, err := oid.DecodeString(objectIDStr)
	if err != nil {
		log.Fatal(fmt.Errorf("parse object ID: %w", err))
	}

	// Configure get parameters
	var prm client.PrmObjectGet

	// Initiate object retrieval
	objHeader, reader, err := c.ObjectGetInit(ctx, containerID, objectID, signer, prm)
	if err != nil {
		switch {
		case errors.Is(err, apistatus.ErrObjectAccessDenied):
			var errAccessDenied apistatus.ObjectAccessDenied
			if errors.As(err, &errAccessDenied) {
				reason := errAccessDenied.Reason()
				if reason != "" {
					log.Fatal(fmt.Errorf("access denied: %s", reason))
				} else {
					log.Fatal(fmt.Errorf("access denied: %w", err))
				}
			}
		case errors.Is(err, apistatus.ErrObjectNotFound):
			log.Fatal(fmt.Errorf("object not found: %w", err))
		default:
			log.Fatal(fmt.Errorf("init object get: %w", err))
		}
	}
	defer reader.Close()

	// Access object header information
	fmt.Printf("Object owner: %s\n", objHeader.Owner())
	fmt.Printf("Payload size: %d bytes\n", objHeader.PayloadSize())
	fmt.Println("Attributes:")
	for _, v := range objHeader.Attributes() {
		fmt.Printf("\t%s: %s\n", v.Key(), v.Value())
	}
	// Example output:
	// Object owner: NfeCA7AuP5zodtJEPsSdkZ2gB3hKdNEhUg
	// Payload size: 1024 bytes
	// Attributes:
	//	FileName: example.txt
	//	ContentType: text/plain

	// Read the payload
	payload, err := io.ReadAll(reader)
	if err != nil {
		log.Fatal(fmt.Errorf("read payload: %w", err))
	}

	fmt.Printf("Successfully read %d bytes\n", len(payload))
	// Example output: Successfully read 1024 bytes
}

// ExampleClient_ObjectPutInit demonstrates how to initialize an object upload
// operation using the NeoFS SDK for Go.
func ExampleClient_ObjectPutInit() {
	const (
		// walletPath is path to your wallet file.
		walletPath = "wallet.json"
		// walletPass is password to decrypt your wallet.
		walletPass = "password"

		// testnetStorageNode is address of NeoFS Testnet storage node.
		// It can be your own node or any public node.
		testnetStorageNode = "st1.t5.fs.neo.org:8080"

		// Example container ID for testing.
		cnrIDStr = "9wqwAcEU8Vn2tq4T1PVpBT6CbVgFb8WfF8ekSRmcQQjq"
	)

	ctx := context.Background()

	// Load wallet and decrypt account
	// This is required to sign the upload request.
	w, err := wallet.NewWalletFromFile(walletPath)
	if err != nil {
		log.Fatal(err)
	}
	addr := w.GetChangeAddress()
	acc := w.GetAccount(addr)
	err = acc.Decrypt(walletPass, w.Scrypt)
	if err != nil {
		log.Fatal(err)
	}
	signer := user.NewAutoIDSigner(acc.PrivateKey().PrivateKey)

	// Create and configure client
	var prmInit client.PrmInit
	c, err := client.New(prmInit)
	if err != nil {
		log.Fatal(fmt.Errorf("client init: %w", err))
	}

	// Connect to NeoFS node
	var prmDial client.PrmDial
	prmDial.SetServerURI(testnetStorageNode)
	// Optional settings
	prmDial.SetTimeout(15 * time.Second)
	prmDial.SetStreamTimeout(15 * time.Second)

	if err = c.Dial(prmDial); err != nil {
		log.Fatal(fmt.Errorf("dial: %w", err))
	}

	// Parse container ID from string
	containerID, err := cid.DecodeString(cnrIDStr)
	if err != nil {
		log.Fatal(fmt.Errorf("parse container ID: %w", err))
	}

	// Create object header with required fields
	var obj = object.New(containerID, signer.UserID())

	// Optional: set custom attributes
	obj.SetAttributes(
		object.NewAttribute(object.AttributeFileName, "example.txt"),
		object.NewAttribute(object.AttributeFilePath, "path/to/example.txt"),
		object.NewAttribute("ContentType", "text/plain"),
	)

	// We need to create a session to authorize the upload
	// Get current epoch time
	ni, err := c.NetworkInfo(ctx, client.PrmNetworkInfo{})
	if err != nil {
		log.Fatal(fmt.Errorf("get current epoch: %w", err))
	}

	currEpoch := ni.CurrentEpoch()
	expireAt := currEpoch + 10 // session valid for 10 epochs

	// Create session for the upload
	var prmSession client.PrmSessionCreate
	prmSession.SetExp(expireAt)
	res, err := c.SessionCreate(ctx, signer, prmSession)
	if err != nil {
		log.Fatal(fmt.Errorf("create session: %w", err))
	}

	var keySession neofsecdsa.PublicKey
	err = keySession.Decode(res.PublicKey())
	if err != nil {
		log.Fatal(fmt.Errorf("decode public session key: %w", err))
	}

	var idSession uuid.UUID
	err = idSession.UnmarshalBinary(res.ID())
	if err != nil {
		log.Fatal(fmt.Errorf("decode session ID: %w", err))
	}

	// Fill session parameters
	var sessionToken session.Object
	sessionToken.SetID(idSession)
	sessionToken.SetNbf(currEpoch)
	sessionToken.SetIat(currEpoch)
	sessionToken.SetExp(expireAt)
	sessionToken.SetAuthKey(&keySession)

	sessionToken.BindContainer(containerID)
	sessionToken.ForVerb(session.VerbObjectPut)
	// Sign the session token with the user's key
	err = sessionToken.Sign(signer)
	if err != nil {
		log.Fatal(fmt.Errorf("sign session token: %w", err))
	}

	// Configure put parameters
	var prm client.PrmObjectPutInit
	prm.WithinSession(sessionToken)

	// Initiate object upload
	writer, err := c.ObjectPutInit(ctx, *obj, signer, prm)
	if err != nil {
		log.Fatal(fmt.Errorf("init object put: %w", err))
	}

	// Write the payload data
	payload := []byte("Hello, NeoFS!")
	_, err = writer.Write(payload)
	if err != nil {
		_ = writer.Close()
		log.Fatal(fmt.Errorf("write payload: %w", err))
	}

	// Close the writer to finalize the upload
	if err = writer.Close(); err != nil {
		log.Fatal(fmt.Errorf("close writer: %w", err))
	}

	// Get the stored object ID
	objectID := writer.GetResult().StoredObjectID()
	fmt.Printf("Successfully uploaded object with ID: %s\n", objectID)
	// Example output: Successfully uploaded object with ID: 7sTMgE8MzpyvqwLh6EJUXdoPJJyMRRRNHyTx2ErvjgNX
}
