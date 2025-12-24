package container_test

import (
	"testing"

	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	"github.com/nspcc-dev/neofs-sdk-go/proto/container"
	prototest "github.com/nspcc-dev/neofs-sdk-go/proto/internal/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	"github.com/stretchr/testify/require"
)

// returns random container.Container_Attribute with all non-zero fields.
func randAttribute() *container.Container_Attribute {
	return &container.Container_Attribute{Key: prototest.RandString(), Value: prototest.RandString()}
}

// returns non-empty list of container.Container_Attribute up to 10 elements.
// Each element may be nil and pointer to zero.
func randAttributes() []*container.Container_Attribute { return prototest.RandRepeated(randAttribute) }

// returns random container.Container with all non-zero fields.
func randContainer() *container.Container {
	return &container.Container{
		Version:         prototest.RandVersion(),
		OwnerId:         prototest.RandOwnerID(),
		Nonce:           prototest.RandBytes(),
		BasicAcl:        prototest.RandUint32(),
		Attributes:      randAttributes(),
		PlacementPolicy: prototest.RandPlacementPolicy(),
	}
}

// returns random container.AnnounceUsedSpaceRequest_Body_Announcement with all
// non-zero fields.
func randAnnouncement() *container.AnnounceUsedSpaceRequest_Body_Announcement {
	return &container.AnnounceUsedSpaceRequest_Body_Announcement{
		Epoch:       prototest.RandUint64(),
		ContainerId: prototest.RandContainerID(),
		UsedSpace:   prototest.RandUint64(),
	}
}

// returns non-empty list of
// container.AnnounceUsedSpaceRequest_Body_Announcement up to 10 elements. Each
// element may be nil and pointer to zero.
func randAnnouncements() []*container.AnnounceUsedSpaceRequest_Body_Announcement {
	return prototest.RandRepeated(randAnnouncement)
}

func TestContainer_Attribute_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*container.Container_Attribute{
		{Key: prototest.RandString()},
		{Value: prototest.RandString()},
		randAttribute(),
	})
}

func TestContainer_MarshalStable(t *testing.T) {
	t.Run("nil in repeated messages", func(t *testing.T) {
		src := &container.Container{
			Attributes: []*container.Container_Attribute{nil, {}},
		}

		var dst container.Container
		require.NoError(t, neofsproto.UnmarshalMessage(neofsproto.MarshalMessage(src), &dst))

		as := dst.GetAttributes()
		require.Len(t, as, 2)
		require.Equal(t, as[0], new(container.Container_Attribute))
		require.Equal(t, as[1], new(container.Container_Attribute))
	})

	prototest.TestMarshalStable(t, []*container.Container{
		randContainer(),
	})
}

func TestPutRequest_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*container.PutRequest_Body{
		{Container: randContainer(), Signature: prototest.RandSignatureRFC6979()},
	})
}

func TestPutResponse_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*container.PutResponse_Body{
		{ContainerId: prototest.RandContainerID()},
	})
}

func TestDeleteRequest_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*container.DeleteRequest_Body{
		{ContainerId: prototest.RandContainerID(), Signature: prototest.RandSignatureRFC6979()},
	})
}

func TestDeleteResponse_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*container.DeleteResponse_Body{})
}

func TestGetRequest_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*container.GetRequest_Body{
		{ContainerId: prototest.RandContainerID()},
	})
}

func TestGetResponse_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*container.GetResponse_Body{
		{
			Container:    randContainer(),
			Signature:    prototest.RandSignatureRFC6979(),
			SessionToken: prototest.RandSessionToken(),
		},
	})
}

func TestListRequest_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*container.ListRequest_Body{
		{OwnerId: prototest.RandOwnerID()},
	})
}

func TestListResponse_Body_MarshalStable(t *testing.T) {
	t.Run("nil in repeated messages", func(t *testing.T) {
		src := &container.ListResponse_Body{
			ContainerIds: []*refs.ContainerID{nil, {}},
		}

		var dst container.ListResponse_Body
		require.NoError(t, neofsproto.UnmarshalMessage(neofsproto.MarshalMessage(src), &dst))

		cs := dst.GetContainerIds()
		require.Len(t, cs, 2)
		require.Equal(t, cs[0], new(refs.ContainerID))
		require.Equal(t, cs[1], new(refs.ContainerID))
	})

	prototest.TestMarshalStable(t, []*container.ListResponse_Body{
		{ContainerIds: prototest.RandContainerIDs()},
	})
}

func TestSetExtendedACLRequest_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*container.SetExtendedACLRequest_Body{
		{
			Eacl:      prototest.RandEACL(),
			Signature: prototest.RandSignatureRFC6979(),
		},
	})
}

func TestSetExtendedACLResponse_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*container.SetExtendedACLResponse_Body{})
}

func TestGetExtendedACLRequest_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*container.GetExtendedACLRequest_Body{
		{ContainerId: prototest.RandContainerID()},
	})
}

func TestGetExtendedACLResponse_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*container.GetExtendedACLResponse_Body{
		{
			Eacl:         prototest.RandEACL(),
			Signature:    prototest.RandSignatureRFC6979(),
			SessionToken: prototest.RandSessionToken(),
		},
	})
}

func TestAnnounceUsedSpaceRequest_Body_Announcement_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*container.AnnounceUsedSpaceRequest_Body_Announcement{
		randAnnouncement(),
	})
}

func TestAnnounceUsedSpaceRequest_Body_MarshalStable(t *testing.T) {
	t.Run("nil in repeated messages", func(t *testing.T) {
		src := &container.AnnounceUsedSpaceRequest_Body{
			Announcements: []*container.AnnounceUsedSpaceRequest_Body_Announcement{nil, {}},
		}

		var dst container.AnnounceUsedSpaceRequest_Body
		require.NoError(t, neofsproto.UnmarshalMessage(neofsproto.MarshalMessage(src), &dst))

		as := dst.GetAnnouncements()
		require.Len(t, as, 2)
		require.Equal(t, as[0], new(container.AnnounceUsedSpaceRequest_Body_Announcement))
		require.Equal(t, as[1], new(container.AnnounceUsedSpaceRequest_Body_Announcement))
	})

	prototest.TestMarshalStable(t, []*container.AnnounceUsedSpaceRequest_Body{
		{Announcements: randAnnouncements()},
	})
}

func TestAnnounceUsedSpaceResponse_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*container.AnnounceUsedSpaceResponse_Body{})
}

func TestSetAttributeRequest_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*container.SetAttributeRequest_Body{
		{
			Parameters: &container.SetAttributeRequest_Body_Parameters{
				ContainerId: prototest.RandContainerID(),
				Attribute:   prototest.RandString(),
				Value:       prototest.RandString(),
				ValidUntil:  prototest.RandUint64(),
			},
			Signature:      prototest.RandSignatureRFC6979(),
			SessionToken:   prototest.RandSessionTokenV2(true),
			SessionTokenV1: prototest.RandSessionToken(),
		},
	})
}

func TestSetAttributeResponse_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*container.SetAttributeResponse_Body{
		{Status: prototest.RandStatus()},
	})
}

func TestRemoveAttributeRequest_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*container.RemoveAttributeRequest_Body{
		{
			Parameters: &container.RemoveAttributeRequest_Body_Parameters{
				ContainerId: prototest.RandContainerID(),
				Attribute:   prototest.RandString(),
				ValidUntil:  prototest.RandUint64(),
			},
			Signature:      prototest.RandSignatureRFC6979(),
			SessionToken:   prototest.RandSessionTokenV2(true),
			SessionTokenV1: prototest.RandSessionToken(),
		},
	})
}

func TestRemoveAttributeResponse_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*container.RemoveAttributeResponse_Body{
		{Status: prototest.RandStatus()},
	})
}
