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

func TestCreateObject(t *testing.T) {
	t.Skip("can be run with the correct endpoint only")

	const strContainer = "g9FJunti7KAR5Fi8p2Gnxvu3wDueqmUYp54Sj49vEGN"
	const endpoint = "s01.neofs.devenv:8080"
	var err error
	ctx := context.Background()
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
