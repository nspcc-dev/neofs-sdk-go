package object

import (
	"crypto/rand"
	"crypto/sha256"
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	ownertest "github.com/nspcc-dev/neofs-sdk-go/owner/test"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	"github.com/nspcc-dev/neofs-sdk-go/signature"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/stretchr/testify/require"
)

func newOID() oid.ID {
	return oid.ID{}
}

func newObj() Object {
	return Object{}
}

func randID(t *testing.T) oid.ID {
	id := newOID()
	id.SetSHA256(randSHA256Checksum(t))

	return id
}

func randSHA256Checksum(t *testing.T) (cs [sha256.Size]byte) {
	_, err := rand.Read(cs[:])
	require.NoError(t, err)

	return
}

func randTZChecksum(t *testing.T) (cs [64]byte) {
	_, err := rand.Read(cs[:])
	require.NoError(t, err)

	return
}

func TestObject_SetID(t *testing.T) {
	obj := newObj()

	id := randID(t)

	obj.SetID(id)

	require.Equal(t, id, obj.ID())
}

func TestObject_SetSignature(t *testing.T) {
	obj := newObj()

	sig := signature.New()
	sig.SetKey([]byte{1, 2, 3})
	sig.SetSign([]byte{4, 5, 6})

	obj.SetSignature(*sig)

	require.Equal(t, *sig, obj.Signature())
}

func TestObject_SetPayload(t *testing.T) {
	obj := newObj()

	payload := make([]byte, 10)
	_, _ = rand.Read(payload)

	obj.SetPayload(payload)

	require.Equal(t, payload, obj.Payload())
}

func TestObject_SetVersion(t *testing.T) {
	obj := newObj()

	ver := version.New()
	ver.SetMajor(1)
	ver.SetMinor(2)

	obj.SetVersion(*ver)

	require.Equal(t, *ver, obj.Version())
}

func TestObject_SetPayloadSize(t *testing.T) {
	obj := newObj()

	sz := uint64(133)
	obj.SetPayloadSize(sz)

	require.Equal(t, sz, obj.PayloadSize())
}

func TestObject_SetContainerID(t *testing.T) {
	obj := newObj()

	cid := cidtest.ID()

	obj.SetContainerID(cid)

	require.Equal(t, cid, obj.ContainerID())
}

func TestObject_SetOwnerID(t *testing.T) {
	obj := newObj()

	ownerID := ownertest.ID()

	obj.SetOwnerID(*ownerID)

	require.Equal(t, *ownerID, obj.OwnerID())
}

func TestObject_SetCreationEpoch(t *testing.T) {
	obj := newObj()

	creat := uint64(228)
	obj.SetCreationEpoch(creat)

	require.Equal(t, creat, obj.CreationEpoch())
}

func TestObject_SetPayloadChecksum(t *testing.T) {
	obj := newObj()
	var cs checksum.Checksum
	cs.SetSHA256(randSHA256Checksum(t))

	obj.SetPayloadChecksum(cs)

	require.Equal(t, cs, obj.PayloadChecksum())
}

func TestObject_SetPayloadHomomorphicHash(t *testing.T) {
	obj := newObj()

	var cs checksum.Checksum
	cs.SetTillichZemor(randTZChecksum(t))

	obj.SetPayloadHomomorphicHash(cs)

	require.Equal(t, cs, obj.PayloadHomomorphicHash())
}

func TestObject_SetAttributes(t *testing.T) {
	obj := newObj()

	var aa Attributes

	var a1 Attribute
	a1.SetKey("key1")
	a1.SetValue("val1")

	var a2 Attribute
	a2.SetKey("key2")
	a2.SetValue("val2")

	aa.Append(a1, a2)

	obj.SetAttributes(aa)

	require.Equal(t, Attributes{a1, a2}, obj.Attributes())
}

func TestObject_SetPreviousID(t *testing.T) {
	obj := newObj()

	prev := randID(t)

	obj.SetPreviousID(prev)

	require.Equal(t, prev, obj.PreviousID())
}

func TestObject_SetChildren(t *testing.T) {
	obj := newObj()

	id1 := randID(t)
	id2 := randID(t)

	obj.SetChildren(id1, id2)

	require.Equal(t, []oid.ID{id1, id2}, obj.Children())
}

func TestObject_SetSplitID(t *testing.T) {
	obj := newObj()

	require.True(t, obj.SplitID().Empty())

	splitID := NewSplitID()
	obj.SetSplitID(splitID)

	require.Equal(t, obj.SplitID(), splitID)
}

func TestObject_SetParent(t *testing.T) {
	obj := newObj()

	require.Nil(t, obj.Parent())

	par := newObj()
	par.SetID(randID(t))
	par.SetContainerID(cidtest.ID())
	par.SetSignature(*signature.New())

	obj.SetParent(&par)

	require.Equal(t, par, *obj.Parent())
}

func TestObject_ToV2(t *testing.T) {
	objV2 := new(object.Object)
	objV2.SetPayload([]byte{1, 2, 3})

	var obj Object
	obj.ReadFromV2(*objV2)

	objV2New := new(object.Object)
	obj.WriteToV2(objV2New)

	require.Equal(t, objV2, objV2New)
}

func TestObject_SetSessionToken(t *testing.T) {
	obj := newObj()

	tok := sessiontest.Token()

	obj.SetSessionToken(*tok)

	require.Equal(t, *tok, obj.SessionToken())
}

func TestObject_SetType(t *testing.T) {
	obj := newObj()

	typ := TypeStorageGroup

	obj.SetType(typ)

	require.Equal(t, typ, obj.Type())
}

func TestObject_CutPayload(t *testing.T) {
	o1 := newObj()

	p1 := []byte{12, 3}
	o1.SetPayload(p1)

	sz := uint64(13)
	o1.SetPayloadSize(sz)

	o2 := o1.CutPayload()

	require.Equal(t, sz, o2.PayloadSize())
	require.Empty(t, o2.Payload())

	sz++
	o1.SetPayloadSize(sz)

	require.Equal(t, sz, o1.PayloadSize())
	require.Equal(t, sz, o2.PayloadSize())

	p2 := []byte{4, 5, 6}
	o2.SetPayload(p2)

	require.Equal(t, p2, o2.Payload())
	require.Equal(t, p1, o1.Payload())
}

func TestObject_SetParentID(t *testing.T) {
	obj := newObj()

	id := randID(t)
	obj.SetParentID(id)

	require.Equal(t, id, obj.ParentID())
}

func TestObject_ResetRelations(t *testing.T) {
	obj := newObj()

	obj.SetPreviousID(randID(t))

	obj.ResetRelations()

	require.True(t, obj.PreviousID().Empty())
}

func TestObject_HasParent(t *testing.T) {
	obj := newObj()

	obj.InitRelations()

	require.True(t, obj.HasParent())

	obj.ResetRelations()

	require.False(t, obj.HasParent())
}
