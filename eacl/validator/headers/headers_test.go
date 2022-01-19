package headers

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	objectV2 "github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/session"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/eacl/validator"
	objectSDK "github.com/nspcc-dev/neofs-sdk-go/object"
	"github.com/stretchr/testify/require"
)

type testLocalStorage struct {
	t *testing.T

	expAddr *objectSDK.Address

	obj *objectSDK.RawObject
}

func (s *testLocalStorage) Head(addr *objectSDK.Address) (*objectSDK.Object, error) {
	require.True(s.t, addr.ContainerID().Equal(addr.ContainerID()) && addr.ObjectID().Equal(addr.ObjectID()))

	return s.obj.Object(), nil
}

func testID(t *testing.T) *objectSDK.ID {
	cs := [sha256.Size]byte{}

	_, err := rand.Read(cs[:])
	require.NoError(t, err)

	id := objectSDK.NewID()
	id.SetSHA256(cs)

	return id
}

func testAddress(t *testing.T) *objectSDK.Address {
	addr := objectSDK.NewAddress()
	addr.SetObjectID(testID(t))
	addr.SetContainerID(cidtest.ID())

	return addr
}

func testXHeaders(strs ...string) []*session.XHeader {
	res := make([]*session.XHeader, 0, len(strs)/2)

	for i := 0; i < len(strs); i += 2 {
		x := new(session.XHeader)
		x.SetKey(strs[i])
		x.SetValue(strs[i+1])

		res = append(res, x)
	}

	return res
}

func TestHeadRequest(t *testing.T) {
	req := new(objectV2.HeadRequest)

	meta := new(session.RequestMetaHeader)
	req.SetMetaHeader(meta)

	body := new(objectV2.HeadRequestBody)
	req.SetBody(body)

	addr := testAddress(t)
	body.SetAddress(addr.ToV2())

	xKey := "x-key"
	xVal := "x-val"
	xHdrs := testXHeaders(
		xKey, xVal,
	)

	meta.SetXHeaders(xHdrs)

	obj := objectSDK.NewRaw()

	attrKey := "attr_key"
	attrVal := "attr_val"
	attr := objectSDK.NewAttribute()
	attr.SetKey(attrKey)
	attr.SetValue(attrVal)
	obj.SetAttributes(attr)

	table := new(eacl.Table)

	priv, err := keys.NewPrivateKey()
	require.NoError(t, err)
	senderKey := priv.PublicKey()

	r := eacl.NewRecord()
	r.SetOperation(eacl.OperationHead)
	r.SetAction(eacl.ActionDeny)
	r.AddFilter(eacl.HeaderFromObject, eacl.MatchStringEqual, attrKey, attrVal)
	r.AddFilter(eacl.HeaderFromRequest, eacl.MatchStringEqual, xKey, xVal)
	eacl.AddFormedTarget(r, eacl.RoleUnknown, (ecdsa.PublicKey)(*senderKey))

	table.AddRecord(r)

	lStorage := &testLocalStorage{
		t:       t,
		expAddr: addr,
		obj:     obj,
	}

	cid := addr.ContainerID()
	unit := new(validator.ValidationUnit).
		WithContainerID(cid).
		WithOperation(eacl.OperationHead).
		WithSenderKey(senderKey.Bytes()).
		WithHeaderSource(
			NewMessageHeaderSource(
				WithObjectStorage(lStorage),
				WithMessageAndRequest(req, nil),
			),
		).
		WithEACLTable(table)

	v := validator.New()

	require.Equal(t, eacl.ActionDeny, v.CalculateAction(unit))

	meta.SetXHeaders(nil)

	require.Equal(t, eacl.ActionAllow, v.CalculateAction(unit))

	meta.SetXHeaders(xHdrs)

	obj.SetAttributes(nil)

	require.Equal(t, eacl.ActionAllow, v.CalculateAction(unit))
}
