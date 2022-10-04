package clientutil_test

import (
	"bytes"
	"context"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/nspcc-dev/neofs-sdk-go/client/clientutil"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/stretchr/testify/require"
)

const strContainer = "pQgKYhdNkHAo6GU9a8q1w5La1h5oeYE4j1RpmXkwHcL"

const strObject = "E8XwYs71F4uRjcYbRxi75KpQYJ2xdy5uL3yoiKsmeqKs"

const endpoint = "s01.neofs.devenv:8080"

// nolint: unused
var ctx = context.TODO()

func TestCreateObject(t *testing.T) {
	t.Skip("can be run with the correct endpoint only")

	var err error
	payload := []byte("Hello, world!\n")
	var prm clientutil.CreateObjectPrm

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

func TestReadObject(t *testing.T) {
	t.Skip("can be run with the correct endpoint only")

	var err error
	buf := bytes.NewBuffer(nil)
	var prm clientutil.ReadObjectPrm

	err = prm.Container.DecodeString(strContainer)
	require.NoError(t, err)

	err = prm.Object.DecodeString(strObject)
	require.NoError(t, err)

	prm.WritePayloadTo(buf)

	err = clientutil.ReadObject(ctx, endpoint, prm)
	require.NoError(t, err)

	log.Println("Object payload is successfully read.")
	log.Println(buf)
}

func TestRemoveObject(t *testing.T) {
	t.Skip("can be run with the correct endpoint only")

	var err error
	var prm clientutil.RemoveObjectPrm

	err = prm.Container.DecodeString(strContainer)
	require.NoError(t, err)

	err = prm.Object.DecodeString(strObject)
	require.NoError(t, err)

	err = clientutil.RemoveObject(ctx, endpoint, prm)
	require.NoError(t, err)

	log.Println("Object successfully removed.")
}
