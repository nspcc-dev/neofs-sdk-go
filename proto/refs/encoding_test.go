package refs_test

import (
	"testing"

	prototest "github.com/nspcc-dev/neofs-sdk-go/proto/internal/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
)

func TestOwnerID_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*refs.OwnerID{
		{Value: []byte{}},
		prototest.RandOwnerID(),
	})
}

func TestContainerID_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*refs.ContainerID{
		{Value: []byte{}},
		prototest.RandContainerID(),
	})
}

func TestObjectID_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*refs.ObjectID{
		{Value: []byte{}},
		prototest.RandObjectID(),
	})
}

func TestAddress_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*refs.Address{
		{ContainerId: new(refs.ContainerID)},
		{ContainerId: prototest.RandContainerID()},
		{ObjectId: new(refs.ObjectID)},
		{ObjectId: prototest.RandObjectID()},
		{
			ContainerId: new(refs.ContainerID),
			ObjectId:    new(refs.ObjectID),
		},
		{
			ContainerId: &refs.ContainerID{Value: []byte{}},
			ObjectId:    &refs.ObjectID{Value: []byte{}},
		},
		prototest.RandObjectAddress(),
	})
}

func TestVersion_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*refs.Version{
		{Major: prototest.RandUint32()},
		{Minor: prototest.RandUint32()},
		prototest.RandVersion(),
	})
}

func TestSignature_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*refs.Signature{
		{Key: []byte{}},
		{Sign: []byte{}},
		{Scheme: prototest.RandInteger[refs.SignatureScheme]()},
		prototest.RandSignature(),
	})
}

func TestSignatureRFC6979_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*refs.SignatureRFC6979{
		{Key: []byte{}},
		{Sign: []byte{}},
		prototest.RandSignatureRFC6979(),
	})
}

func TestChecksum_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*refs.Checksum{
		{Sum: []byte{}},
		prototest.RandChecksum(),
	})
}

func TestSubnetID(t *testing.T) {
	prototest.TestMarshalStable(t, []*refs.SubnetID{
		prototest.RandSubnetID(),
	})
}
