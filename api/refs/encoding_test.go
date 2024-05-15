package refs_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestOwnerID(t *testing.T) {
	v := &refs.OwnerID{
		Value: []byte("any_owner"),
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res refs.OwnerID
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Value, res.Value)
}

func TestVersion(t *testing.T) {
	v := &refs.Version{
		Major: 123,
		Minor: 456,
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res refs.Version
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.EqualValues(t, v.Major, res.Major)
	require.EqualValues(t, v.Minor, res.Minor)
}

func TestContainerID(t *testing.T) {
	v := &refs.ContainerID{
		Value: []byte("any_container"),
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res refs.ContainerID
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Value, res.Value)
}

func TestObjectID(t *testing.T) {
	v := &refs.ObjectID{
		Value: []byte("any_object"),
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res refs.ObjectID
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Value, res.Value)
}

func TestAddress(t *testing.T) {
	v := &refs.Address{
		ContainerId: &refs.ContainerID{Value: []byte("any_container")},
		ObjectId:    &refs.ObjectID{Value: []byte("any_object")},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res refs.Address
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.ContainerId, res.ContainerId)
	require.Equal(t, v.ObjectId, res.ObjectId)
}

func TestSubnetID(t *testing.T) {
	v := &refs.SubnetID{
		Value: 123456,
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res refs.SubnetID
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Value, res.Value)
}

func TestSignature(t *testing.T) {
	v := &refs.Signature{
		Key:    []byte("any_key"),
		Sign:   []byte("any_val"),
		Scheme: 123,
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res refs.Signature
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Key, res.Key)
	require.Equal(t, v.Sign, res.Sign)
	require.Equal(t, v.Scheme, res.Scheme)
}

func TestSignatureRFC6979(t *testing.T) {
	v := &refs.SignatureRFC6979{
		Key:  []byte("any_key"),
		Sign: []byte("any_val"),
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res refs.SignatureRFC6979
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Key, res.Key)
	require.Equal(t, v.Sign, res.Sign)
}

func TestChecksum(t *testing.T) {
	v := &refs.Checksum{
		Type: 321,
		Sum:  []byte("any_checksum"),
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res refs.Checksum
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Type, res.Type)
	require.Equal(t, v.Sum, res.Sum)
}
