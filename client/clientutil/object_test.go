package clientutil_test

import (
	"bytes"
	"context"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/nspcc-dev/neofs-sdk-go/client/clientutil"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/stretchr/testify/require"
)

const (
	strContainer = "NZddit4fqr2vnWNzPxeJX6L2KusCVFTARZuwA63aHqE"
	strObject    = "E8XwYs71F4uRjcYbRxi75KpQYJ2xdy5uL3yoiKsmeqKs"
	endpoint     = "s01.neofs.devenv:8080"
	filePath     = "./test_file"
)

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

func TestListObjects(t *testing.T) {
	t.Skip("can be run with the correct endpoint only")

	var err error
	var prm clientutil.ListObjectsPrm

	err = prm.Container.DecodeString(strContainer)
	require.NoError(t, err)

	prm.SetHandler(func(id oid.ID) {
		log.Println("found object:", id)
	})

	err = clientutil.ListObjects(ctx, endpoint, prm)
	require.NoError(t, err)
}

func TestUploadFile(t *testing.T) {
	t.Skip("can be run with the correct endpoint and file only")

	var err error
	var cnr cid.ID

	err = cnr.DecodeString(strContainer)
	require.NoError(t, err)

	err = clientutil.UploadFileByPath(ctx, endpoint, cnr, filePath)
	require.NoError(t, err)

	log.Println("File successfully uploaded.")
}

func TestRestoreFile(t *testing.T) {
	t.Skip("can be run with the correct endpoint and file only")

	var err error
	var cnr cid.ID

	err = cnr.DecodeString(strContainer)
	require.NoError(t, err)

	err = clientutil.RestoreFileByPath(ctx, endpoint, cnr, filePath)
	require.NoError(t, err)

	log.Println("File successfully restored.")
}
