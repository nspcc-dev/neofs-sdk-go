package container_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/acl"
	"github.com/nspcc-dev/neofs-sdk-go/api/container"
	"github.com/nspcc-dev/neofs-sdk-go/api/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	"github.com/nspcc-dev/neofs-sdk-go/api/session"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestPutRequest_Body(t *testing.T) {
	v := &container.PutRequest_Body{
		Container: &container.Container{
			Version:  &refs.Version{Major: 1, Minor: 2},
			OwnerId:  &refs.OwnerID{Value: []byte("any_owner")},
			Nonce:    []byte("any_nonce"),
			BasicAcl: 3,
			Attributes: []*container.Container_Attribute{
				{Key: "attr_key1", Value: "attr_val1"},
				{Key: "attr_key2", Value: "attr_val2"},
			},
			PlacementPolicy: &netmap.PlacementPolicy{
				Replicas: []*netmap.Replica{
					{Count: 4, Selector: "selector1"},
					{Count: 5, Selector: "selector2"},
				},
				ContainerBackupFactor: 6,
				Selectors: []*netmap.Selector{
					{Name: "selector3", Count: 7, Clause: 8, Attribute: "attr1", Filter: "filter1"},
					{Name: "selector4", Count: 9, Clause: 10, Attribute: "attr2", Filter: "filter2"},
				},
				Filters: []*netmap.Filter{
					{Name: "filter3", Key: "filter_key1", Op: 11, Value: "filter_val1", Filters: []*netmap.Filter{
						{Name: "filter4", Key: "filter_key2", Op: 12, Value: "filter_val2"},
						{Name: "filter5", Key: "filter_key3", Op: 13, Value: "filter_val3"},
					}},
					{Name: "filter6", Key: "filter_key4", Op: 14, Value: "filter_val4", Filters: []*netmap.Filter{
						{Name: "filter7", Key: "filter_key5", Op: 15, Value: "filter_val5"},
						{Name: "filter8", Key: "filter_key6", Op: 16, Value: "filter_val6"},
					}},
				},
				SubnetId: &refs.SubnetID{Value: 17},
			},
		},
		Signature: &refs.SignatureRFC6979{Key: []byte("any_pubkey"), Sign: []byte("any_signature")},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res container.PutRequest_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Container, res.Container)
	require.Equal(t, v.Signature, res.Signature)
}

func TestPutResponse_Body(t *testing.T) {
	v := &container.PutResponse_Body{
		ContainerId: &refs.ContainerID{Value: []byte("any_container")},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res container.PutResponse_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.ContainerId, res.ContainerId)
}

func TestDeleteRequest_Body(t *testing.T) {
	v := &container.DeleteRequest_Body{
		ContainerId: &refs.ContainerID{Value: []byte("any_container")},
		Signature:   &refs.SignatureRFC6979{Key: []byte("any_pubkey"), Sign: []byte("any_signature")},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res container.DeleteRequest_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.ContainerId, res.ContainerId)
	require.Equal(t, v.Signature, res.Signature)
}

func TestDeleteResponse_Body(t *testing.T) {
	var v container.DeleteResponse_Body
	require.Zero(t, v.MarshaledSize())
	require.NotPanics(t, func() { v.MarshalStable(nil) })
	b := []byte("not_a_protobuf")
	v.MarshalStable(b)
	require.EqualValues(t, "not_a_protobuf", b)
}

func TestGetRequest_Body(t *testing.T) {
	v := &container.GetRequest_Body{
		ContainerId: &refs.ContainerID{Value: []byte("any_container")},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res container.GetRequest_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.ContainerId, res.ContainerId)
}

func TestGetResponse_Body(t *testing.T) {
	v := &container.GetResponse_Body{
		Container: &container.Container{
			Version:  &refs.Version{Major: 1, Minor: 2},
			OwnerId:  &refs.OwnerID{Value: []byte("any_owner")},
			Nonce:    []byte("any_nonce"),
			BasicAcl: 3,
			Attributes: []*container.Container_Attribute{
				{Key: "attr_key1", Value: "attr_val1"},
				{Key: "attr_key2", Value: "attr_val2"},
			},
			PlacementPolicy: &netmap.PlacementPolicy{
				Replicas: []*netmap.Replica{
					{Count: 4, Selector: "selector1"},
					{Count: 5, Selector: "selector2"},
				},
				ContainerBackupFactor: 6,
				Selectors: []*netmap.Selector{
					{Name: "selector3", Count: 7, Clause: 8, Attribute: "attr1", Filter: "filter1"},
					{Name: "selector4", Count: 9, Clause: 10, Attribute: "attr2", Filter: "filter2"},
				},
				Filters: []*netmap.Filter{
					{Name: "filter3", Key: "filter_key1", Op: 11, Value: "filter_val1", Filters: []*netmap.Filter{
						{Name: "filter4", Key: "filter_key2", Op: 12, Value: "filter_val2"},
						{Name: "filter5", Key: "filter_key3", Op: 13, Value: "filter_val3"},
					}},
					{Name: "filter6", Key: "filter_key4", Op: 14, Value: "filter_val4", Filters: []*netmap.Filter{
						{Name: "filter7", Key: "filter_key5", Op: 15, Value: "filter_val5"},
						{Name: "filter8", Key: "filter_key6", Op: 16, Value: "filter_val6"},
					}},
				},
				SubnetId: &refs.SubnetID{Value: 17},
			},
		},
		Signature: &refs.SignatureRFC6979{Key: []byte("any_pubkey"), Sign: []byte("any_signature")},
		SessionToken: &session.SessionToken{
			Body: &session.SessionToken_Body{
				Id:         []byte("any_ID"),
				OwnerId:    &refs.OwnerID{Value: []byte("any_owner")},
				Lifetime:   &session.SessionToken_Body_TokenLifetime{Exp: 18, Nbf: 19, Iat: 20},
				SessionKey: []byte("any_key"),
			},
			Signature: &refs.Signature{Key: []byte("any_key"), Sign: []byte("any_signature"), Scheme: 21},
		},
	}

	testWithSessionContext := func(setCtx func(*session.SessionToken_Body)) {
		setCtx(v.SessionToken.Body)
		sz := v.MarshaledSize()
		b := make([]byte, sz)
		v.MarshalStable(b)

		var res container.GetResponse_Body
		err := proto.Unmarshal(b, &res)
		require.NoError(t, err)
		require.Empty(t, res.ProtoReflect().GetUnknown())
		require.Equal(t, v.Container, res.Container)
		require.Equal(t, v.Signature, res.Signature)
		require.Equal(t, v.SessionToken, res.SessionToken)
	}

	testWithSessionContext(func(body *session.SessionToken_Body) { body.Context = nil })
	testWithSessionContext(func(body *session.SessionToken_Body) {
		body.Context = &session.SessionToken_Body_Container{
			Container: &session.ContainerSessionContext{
				Verb: 1, Wildcard: true,
				ContainerId: &refs.ContainerID{Value: []byte("any_container")},
			},
		}
	})
	testWithSessionContext(func(body *session.SessionToken_Body) {
		body.Context = &session.SessionToken_Body_Object{
			Object: &session.ObjectSessionContext{
				Verb: 1,
				Target: &session.ObjectSessionContext_Target{
					Container: &refs.ContainerID{Value: []byte("any_container")},
					Objects: []*refs.ObjectID{
						{Value: []byte("any_object1")},
						{Value: []byte("any_object2")},
					},
				},
			},
		}
	})
}

func TestListRequest_Body(t *testing.T) {
	v := &container.ListRequest_Body{
		OwnerId: &refs.OwnerID{Value: []byte("any_owner")},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res container.ListRequest_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.OwnerId, res.OwnerId)
}

func TestListResponse_Body(t *testing.T) {
	v := &container.ListResponse_Body{
		ContainerIds: []*refs.ContainerID{
			{Value: []byte("any_container1")},
			{Value: []byte("any_container2")},
		},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res container.ListResponse_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.ContainerIds, res.ContainerIds)
}

func TestSetExtendedACLRequest_Body(t *testing.T) {
	v := &container.SetExtendedACLRequest_Body{
		Eacl: &acl.EACLTable{
			Version:     &refs.Version{Major: 123, Minor: 456},
			ContainerId: &refs.ContainerID{Value: []byte("any_container")},
			Records: []*acl.EACLRecord{
				{
					Operation: 1, Action: 2,
					Filters: []*acl.EACLRecord_Filter{
						{HeaderType: 3, MatchType: 4, Key: "key1", Value: "val1"},
						{HeaderType: 5, MatchType: 6, Key: "key2", Value: "val2"},
					},
					Targets: []*acl.EACLRecord_Target{
						{Role: 7, Keys: [][]byte{{0}, {1}}},
						{Role: 8, Keys: [][]byte{{2}, {3}}},
					},
				},
				{
					Operation: 9, Action: 10,
					Filters: []*acl.EACLRecord_Filter{
						{HeaderType: 11, MatchType: 12, Key: "key3", Value: "val3"},
						{HeaderType: 13, MatchType: 14, Key: "key4", Value: "val4"},
					},
					Targets: []*acl.EACLRecord_Target{
						{Role: 15, Keys: [][]byte{{4}, {5}}},
						{Role: 16, Keys: [][]byte{{6}, {7}}},
					},
				},
			},
		},
		Signature: &refs.SignatureRFC6979{Key: []byte("any_pubkey"), Sign: []byte("any_signature")},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res container.SetExtendedACLRequest_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Eacl, res.Eacl)
	require.Equal(t, v.Signature, res.Signature)
}

func TestSetExtendedACLResponse_Body(t *testing.T) {
	var v container.SetExtendedACLResponse_Body
	require.Zero(t, v.MarshaledSize())
	require.NotPanics(t, func() { v.MarshalStable(nil) })
	b := []byte("not_a_protobuf")
	v.MarshalStable(b)
	require.EqualValues(t, "not_a_protobuf", b)
}

func TestGetExtendedACLRequest_Body(t *testing.T) {
	v := &container.GetExtendedACLRequest_Body{
		ContainerId: &refs.ContainerID{Value: []byte("any_container")},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res container.GetExtendedACLRequest_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.ContainerId, res.ContainerId)
}

func TestGetExtendedACLResponse(t *testing.T) {
	v := &container.GetExtendedACLResponse_Body{
		Eacl: &acl.EACLTable{
			Version:     &refs.Version{Major: 123, Minor: 456},
			ContainerId: &refs.ContainerID{Value: []byte("any_container")},
			Records: []*acl.EACLRecord{
				{
					Operation: 1, Action: 2,
					Filters: []*acl.EACLRecord_Filter{
						{HeaderType: 3, MatchType: 4, Key: "key1", Value: "val1"},
						{HeaderType: 5, MatchType: 6, Key: "key2", Value: "val2"},
					},
					Targets: []*acl.EACLRecord_Target{
						{Role: 7, Keys: [][]byte{{0}, {1}}},
						{Role: 8, Keys: [][]byte{{2}, {3}}},
					},
				},
				{
					Operation: 9, Action: 10,
					Filters: []*acl.EACLRecord_Filter{
						{HeaderType: 11, MatchType: 12, Key: "key3", Value: "val3"},
						{HeaderType: 13, MatchType: 14, Key: "key4", Value: "val4"},
					},
					Targets: []*acl.EACLRecord_Target{
						{Role: 15, Keys: [][]byte{{4}, {5}}},
						{Role: 16, Keys: [][]byte{{6}, {7}}},
					},
				},
			},
		},
		Signature: &refs.SignatureRFC6979{Key: []byte("any_pubkey"), Sign: []byte("any_signature")},
		SessionToken: &session.SessionToken{
			Body: &session.SessionToken_Body{
				Id:         []byte("any_ID"),
				OwnerId:    &refs.OwnerID{Value: []byte("any_owner")},
				Lifetime:   &session.SessionToken_Body_TokenLifetime{Exp: 18, Nbf: 19, Iat: 20},
				SessionKey: []byte("any_key"),
			},
			Signature: &refs.Signature{Key: []byte("any_key"), Sign: []byte("any_signature"), Scheme: 21},
		},
	}

	testWithSessionContext := func(setCtx func(*session.SessionToken_Body)) {
		setCtx(v.SessionToken.Body)
		sz := v.MarshaledSize()
		b := make([]byte, sz)
		v.MarshalStable(b)

		var res container.GetExtendedACLResponse_Body
		err := proto.Unmarshal(b, &res)
		require.NoError(t, err)
		require.Empty(t, res.ProtoReflect().GetUnknown())
		require.Equal(t, v.Eacl, res.Eacl)
		require.Equal(t, v.Signature, res.Signature)
		require.Equal(t, v.SessionToken, res.SessionToken)
	}

	testWithSessionContext(func(body *session.SessionToken_Body) { body.Context = nil })
	testWithSessionContext(func(body *session.SessionToken_Body) {
		body.Context = &session.SessionToken_Body_Container{
			Container: &session.ContainerSessionContext{
				Verb: 1, Wildcard: true,
				ContainerId: &refs.ContainerID{Value: []byte("any_container")},
			},
		}
	})
	testWithSessionContext(func(body *session.SessionToken_Body) {
		body.Context = &session.SessionToken_Body_Object{
			Object: &session.ObjectSessionContext{
				Verb: 1,
				Target: &session.ObjectSessionContext_Target{
					Container: &refs.ContainerID{Value: []byte("any_container")},
					Objects: []*refs.ObjectID{
						{Value: []byte("any_object1")},
						{Value: []byte("any_object2")},
					},
				},
			},
		}
	})
}

func TestAnnounceUsedSpaceRequest_Body(t *testing.T) {
	v := &container.AnnounceUsedSpaceRequest_Body{
		Announcements: []*container.AnnounceUsedSpaceRequest_Body_Announcement{
			{Epoch: 1, ContainerId: &refs.ContainerID{Value: []byte("any_container1")}, UsedSpace: 2},
			{Epoch: 3, ContainerId: &refs.ContainerID{Value: []byte("any_container2")}, UsedSpace: 4},
		},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res container.AnnounceUsedSpaceRequest_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Announcements, res.Announcements)
}

func TestAnnounceUsedSpaceResponse_Body(t *testing.T) {
	var v container.AnnounceUsedSpaceResponse_Body
	require.Zero(t, v.MarshaledSize())
	require.NotPanics(t, func() { v.MarshalStable(nil) })
	b := []byte("not_a_protobuf")
	v.MarshalStable(b)
	require.EqualValues(t, "not_a_protobuf", b)
}
