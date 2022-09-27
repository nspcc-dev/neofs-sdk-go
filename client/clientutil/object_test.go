package clientutil_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/nspcc-dev/neofs-sdk-go/client/clientutil"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/stretchr/testify/require"
)

func TestCreateObject(t *testing.T) {
	t.Skip("can be run with the correct endpoint only")

	const strContainer = "9JbqitaFuaJfq7Gsy9XnwWkhg6JRDSe46FCrKUbcQHCf"
	const endpoint = "s01.neofs.devenv:8080"
	var err error
	ctx := context.Background()
	payload := []byte("Hello, world!\n")

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	prm := clientutil.CreateObjectPrm{
		Signer: *key,
	}

	err = prm.Container.DecodeString(strContainer)
	require.NoError(t, err)

	prm.SetPayload(bytes.NewBuffer(payload))

	prm.AddAttribute(object.AttributeTimestamp, strconv.FormatInt(time.Now().UTC().Unix(), 10))
	prm.AddAttribute("Test", "true")

	prm.SetIDHandler(func(id oid.ID) {
		log.Println("Object has been successfully stored:", id)
	})

	err = clientutil.CreateObject(ctx, endpoint, prm)
	require.NoError(t, err)
}
