package object_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/object"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	"github.com/nspcc-dev/neofs-sdk-go/api/session"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestObject(t *testing.T) {
	v := &object.Object{
		ObjectId: &refs.ObjectID{Value: []byte("any_object_ID")},
		Signature: &refs.Signature{
			Key:    []byte("any_public_key"),
			Sign:   []byte("any_signature"),
			Scheme: 1,
		},
		Header: &object.Header{
			Version:         &refs.Version{Major: 2, Minor: 3},
			ContainerId:     &refs.ContainerID{Value: []byte("any_container")},
			OwnerId:         &refs.OwnerID{Value: []byte("any_owner")},
			CreationEpoch:   4,
			PayloadLength:   5,
			PayloadHash:     &refs.Checksum{Type: 6, Sum: []byte("any_checksum")},
			ObjectType:      7,
			HomomorphicHash: &refs.Checksum{Type: 8, Sum: []byte("any_homomorphic_checksum")},
			SessionToken: &session.SessionToken{
				Body: &session.SessionToken_Body{
					Id:         []byte("any_ID"),
					OwnerId:    &refs.OwnerID{Value: []byte("any_owner")},
					Lifetime:   &session.SessionToken_Body_TokenLifetime{Exp: 9, Nbf: 10, Iat: 11},
					SessionKey: []byte("any_key"),
				},
				Signature: &refs.Signature{Key: []byte("any_key"), Sign: []byte("any_signature"), Scheme: 12},
			},
			Attributes: []*object.Header_Attribute{
				{Key: "attr_key1", Value: "attr_val1"},
				{Key: "attr_key2", Value: "attr_val2"},
			},
			Split: &object.Header_Split{
				Parent:   &refs.ObjectID{Value: []byte("any_parent_ID")},
				Previous: &refs.ObjectID{Value: []byte("any_previous")},
				ParentSignature: &refs.Signature{
					Key:    []byte("any_parent_public_key"),
					Sign:   []byte("any_parent_signature"),
					Scheme: 13,
				},
				ParentHeader: &object.Header{
					Version:         &refs.Version{Major: 100, Minor: 101},
					ContainerId:     &refs.ContainerID{Value: []byte("any_parent_container")},
					OwnerId:         &refs.OwnerID{Value: []byte("any_parent_owner")},
					CreationEpoch:   102,
					PayloadLength:   103,
					PayloadHash:     &refs.Checksum{Type: 104, Sum: []byte("any_parent_checksum")},
					ObjectType:      105,
					HomomorphicHash: &refs.Checksum{Type: 106, Sum: []byte("any_parent_homomorphic_checksum")},
					Attributes: []*object.Header_Attribute{
						{Key: "parent_attr_key1", Value: "parent_attr_val2"},
						{Key: "parent_attr_key2", Value: "parent_attr_val2"},
					},
				},
				Children: []*refs.ObjectID{{Value: []byte("any_child1")}, {Value: []byte("any_child2")}},
				SplitId:  []byte("any_split_ID"),
				First:    &refs.ObjectID{Value: []byte("any_first_ID")},
			},
		},
		Payload: []byte("any_payload"),
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res object.Object
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.ObjectId, res.ObjectId)
	require.Equal(t, v.Signature, res.Signature)
	require.Equal(t, v.Header, res.Header)
	require.Equal(t, v.Payload, res.Payload)
}

func TestGetRequest_Body(t *testing.T) {
	v := &object.GetRequest_Body{
		Address: &refs.Address{
			ContainerId: &refs.ContainerID{Value: []byte("any_container")},
			ObjectId:    &refs.ObjectID{Value: []byte("any_object")},
		},
		Raw: true,
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res object.GetRequest_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Address, res.Address)
	require.Equal(t, v.Raw, res.Raw)
}

func TestGetResponse_Body(t *testing.T) {
	var v object.GetResponse_Body

	testWithPart := func(setPart func(*object.GetResponse_Body)) {
		setPart(&v)
		sz := v.MarshaledSize()
		b := make([]byte, sz)
		v.MarshalStable(b)

		var res object.GetResponse_Body
		err := proto.Unmarshal(b, &res)
		require.NoError(t, err)
		require.Empty(t, res.ProtoReflect().GetUnknown())
		require.Equal(t, v.ObjectPart, res.ObjectPart)
	}

	testWithPart(func(body *object.GetResponse_Body) { body.ObjectPart = nil })
	testWithPart(func(body *object.GetResponse_Body) {
		body.ObjectPart = &object.GetResponse_Body_Chunk{Chunk: []byte("any_chunk")}
	})
	testWithPart(func(body *object.GetResponse_Body) {
		body.ObjectPart = &object.GetResponse_Body_SplitInfo{
			SplitInfo: &object.SplitInfo{
				SplitId:   []byte("any_split_ID"),
				LastPart:  &refs.ObjectID{Value: []byte("any_last")},
				Link:      &refs.ObjectID{Value: []byte("any_link")},
				FirstPart: &refs.ObjectID{Value: []byte("any_first")},
			},
		}
	})
	testWithPart(func(body *object.GetResponse_Body) {
		body.ObjectPart = &object.GetResponse_Body_Init_{
			Init: &object.GetResponse_Body_Init{
				ObjectId: &refs.ObjectID{Value: []byte("any_object_ID")},
				Signature: &refs.Signature{
					Key:    []byte("any_public_key"),
					Sign:   []byte("any_signature"),
					Scheme: 1,
				},
				Header: &object.Header{
					Version:         &refs.Version{Major: 2, Minor: 3},
					ContainerId:     &refs.ContainerID{Value: []byte("any_container")},
					OwnerId:         &refs.OwnerID{Value: []byte("any_owner")},
					CreationEpoch:   4,
					PayloadLength:   5,
					PayloadHash:     &refs.Checksum{Type: 6, Sum: []byte("any_checksum")},
					ObjectType:      7,
					HomomorphicHash: &refs.Checksum{Type: 8, Sum: []byte("any_homomorphic_checksum")},
					SessionToken: &session.SessionToken{
						Body: &session.SessionToken_Body{
							Id:         []byte("any_ID"),
							OwnerId:    &refs.OwnerID{Value: []byte("any_owner")},
							Lifetime:   &session.SessionToken_Body_TokenLifetime{Exp: 9, Nbf: 10, Iat: 11},
							SessionKey: []byte("any_key"),
						},
						Signature: &refs.Signature{Key: []byte("any_key"), Sign: []byte("any_signature"), Scheme: 12},
					},
					Attributes: []*object.Header_Attribute{
						{Key: "attr_key1", Value: "attr_val1"},
						{Key: "attr_key2", Value: "attr_val2"},
					},
					Split: &object.Header_Split{
						Parent:   &refs.ObjectID{Value: []byte("any_parent_ID")},
						Previous: &refs.ObjectID{Value: []byte("any_previous")},
						ParentSignature: &refs.Signature{
							Key:    []byte("any_parent_public_key"),
							Sign:   []byte("any_parent_signature"),
							Scheme: 13,
						},
						ParentHeader: &object.Header{
							Version:         &refs.Version{Major: 100, Minor: 101},
							ContainerId:     &refs.ContainerID{Value: []byte("any_parent_container")},
							OwnerId:         &refs.OwnerID{Value: []byte("any_parent_owner")},
							CreationEpoch:   102,
							PayloadLength:   103,
							PayloadHash:     &refs.Checksum{Type: 104, Sum: []byte("any_parent_checksum")},
							ObjectType:      105,
							HomomorphicHash: &refs.Checksum{Type: 106, Sum: []byte("any_parent_homomorphic_checksum")},
							Attributes: []*object.Header_Attribute{
								{Key: "parent_attr_key1", Value: "parent_attr_val2"},
								{Key: "parent_attr_key2", Value: "parent_attr_val2"},
							},
						},
						Children: []*refs.ObjectID{{Value: []byte("any_child1")}, {Value: []byte("any_child2")}},
						SplitId:  []byte("any_split_ID"),
						First:    &refs.ObjectID{Value: []byte("any_first_ID")},
					},
				},
			},
		}
	})
}

func TestHeadRequest_Body(t *testing.T) {
	v := &object.HeadRequest_Body{
		Address: &refs.Address{
			ContainerId: &refs.ContainerID{Value: []byte("any_container")},
			ObjectId:    &refs.ObjectID{Value: []byte("any_object")},
		},
		MainOnly: true,
		Raw:      true,
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res object.HeadRequest_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Address, res.Address)
	require.Equal(t, v.MainOnly, res.MainOnly)
	require.Equal(t, v.Raw, res.Raw)
}

func TestHeadResponse_Body(t *testing.T) {
	var v object.HeadResponse_Body

	testWithPart := func(setPart func(body *object.HeadResponse_Body)) {
		setPart(&v)
		sz := v.MarshaledSize()
		b := make([]byte, sz)
		v.MarshalStable(b)

		var res object.HeadResponse_Body
		err := proto.Unmarshal(b, &res)
		require.NoError(t, err)
		require.Empty(t, res.ProtoReflect().GetUnknown())
		require.Equal(t, v.Head, res.Head)
	}

	testWithPart(func(body *object.HeadResponse_Body) { body.Head = nil })

	testWithPart(func(body *object.HeadResponse_Body) {
		body.Head = &object.HeadResponse_Body_ShortHeader{
			ShortHeader: &object.ShortHeader{
				Version:         &refs.Version{Major: 2, Minor: 3},
				OwnerId:         &refs.OwnerID{Value: []byte("any_owner")},
				CreationEpoch:   4,
				PayloadLength:   5,
				PayloadHash:     &refs.Checksum{Type: 6, Sum: []byte("any_checksum")},
				ObjectType:      7,
				HomomorphicHash: &refs.Checksum{Type: 8, Sum: []byte("any_homomorphic_checksum")},
			},
		}
	})

	testWithPart(func(body *object.HeadResponse_Body) {
		body.Head = &object.HeadResponse_Body_SplitInfo{
			SplitInfo: &object.SplitInfo{
				SplitId:   []byte("any_split_ID"),
				LastPart:  &refs.ObjectID{Value: []byte("any_last")},
				Link:      &refs.ObjectID{Value: []byte("any_link")},
				FirstPart: &refs.ObjectID{Value: []byte("any_first")},
			},
		}
	})

	testWithPart(func(body *object.HeadResponse_Body) {
		body.Head = &object.HeadResponse_Body_Header{
			Header: &object.HeaderWithSignature{
				Header: &object.Header{
					Version:         &refs.Version{Major: 2, Minor: 3},
					ContainerId:     &refs.ContainerID{Value: []byte("any_container")},
					OwnerId:         &refs.OwnerID{Value: []byte("any_owner")},
					CreationEpoch:   4,
					PayloadLength:   5,
					PayloadHash:     &refs.Checksum{Type: 6, Sum: []byte("any_checksum")},
					ObjectType:      7,
					HomomorphicHash: &refs.Checksum{Type: 8, Sum: []byte("any_homomorphic_checksum")},
					SessionToken: &session.SessionToken{
						Body: &session.SessionToken_Body{
							Id:         []byte("any_ID"),
							OwnerId:    &refs.OwnerID{Value: []byte("any_owner")},
							Lifetime:   &session.SessionToken_Body_TokenLifetime{Exp: 9, Nbf: 10, Iat: 11},
							SessionKey: []byte("any_key"),
						},
						Signature: &refs.Signature{Key: []byte("any_key"), Sign: []byte("any_signature"), Scheme: 12},
					},
					Attributes: []*object.Header_Attribute{
						{Key: "attr_key1", Value: "attr_val1"},
						{Key: "attr_key2", Value: "attr_val2"},
					},
					Split: &object.Header_Split{
						Parent:   &refs.ObjectID{Value: []byte("any_parent_ID")},
						Previous: &refs.ObjectID{Value: []byte("any_previous")},
						ParentSignature: &refs.Signature{
							Key:    []byte("any_parent_public_key"),
							Sign:   []byte("any_parent_signature"),
							Scheme: 13,
						},
						ParentHeader: &object.Header{
							Version:         &refs.Version{Major: 100, Minor: 101},
							ContainerId:     &refs.ContainerID{Value: []byte("any_parent_container")},
							OwnerId:         &refs.OwnerID{Value: []byte("any_parent_owner")},
							CreationEpoch:   102,
							PayloadLength:   103,
							PayloadHash:     &refs.Checksum{Type: 104, Sum: []byte("any_parent_checksum")},
							ObjectType:      105,
							HomomorphicHash: &refs.Checksum{Type: 106, Sum: []byte("any_parent_homomorphic_checksum")},
							Attributes: []*object.Header_Attribute{
								{Key: "parent_attr_key1", Value: "parent_attr_val2"},
								{Key: "parent_attr_key2", Value: "parent_attr_val2"},
							},
						},
						Children: []*refs.ObjectID{{Value: []byte("any_child1")}, {Value: []byte("any_child2")}},
						SplitId:  []byte("any_split_ID"),
						First:    &refs.ObjectID{Value: []byte("any_first_ID")},
					},
				},
				Signature: &refs.Signature{
					Key:    []byte("any_public_key"),
					Sign:   []byte("any_signature"),
					Scheme: 987,
				},
			},
		}
	})
}

func TestGetRangeRequest_Body(t *testing.T) {
	v := &object.GetRangeRequest_Body{
		Address: &refs.Address{
			ContainerId: &refs.ContainerID{Value: []byte("any_container")},
			ObjectId:    &refs.ObjectID{Value: []byte("any_object")},
		},
		Range: &object.Range{Offset: 1, Length: 2},
		Raw:   true,
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res object.GetRangeRequest_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Address, res.Address)
	require.Equal(t, v.Range, res.Range)
	require.Equal(t, v.Raw, res.Raw)
}

func TestGetRangeResponse_Body(t *testing.T) {
	var v object.GetRangeResponse_Body

	testWithPart := func(setPart func(body *object.GetRangeResponse_Body)) {
		setPart(&v)
		sz := v.MarshaledSize()
		b := make([]byte, sz)
		v.MarshalStable(b)

		var res object.GetRangeResponse_Body
		err := proto.Unmarshal(b, &res)
		require.NoError(t, err)
		require.Empty(t, res.ProtoReflect().GetUnknown())
		require.Equal(t, v.RangePart, res.RangePart)
	}

	testWithPart(func(body *object.GetRangeResponse_Body) { body.RangePart = nil })

	testWithPart(func(body *object.GetRangeResponse_Body) {
		body.RangePart = &object.GetRangeResponse_Body_Chunk{Chunk: []byte("any_chunk")}
	})

	testWithPart(func(body *object.GetRangeResponse_Body) {
		body.RangePart = &object.GetRangeResponse_Body_SplitInfo{
			SplitInfo: &object.SplitInfo{
				SplitId:   []byte("any_split_ID"),
				LastPart:  &refs.ObjectID{Value: []byte("any_last")},
				Link:      &refs.ObjectID{Value: []byte("any_link")},
				FirstPart: &refs.ObjectID{Value: []byte("any_first")},
			},
		}
	})
}

func TestGetRangeHashRequest_Body(t *testing.T) {
	v := &object.GetRangeHashRequest_Body{
		Address: &refs.Address{
			ContainerId: &refs.ContainerID{Value: []byte("any_container")},
			ObjectId:    &refs.ObjectID{Value: []byte("any_object")},
		},
		Ranges: []*object.Range{
			{Offset: 1, Length: 2},
			{Offset: 3, Length: 4},
		},
		Salt: []byte("any_salt"),
		Type: 5,
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res object.GetRangeHashRequest_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Address, res.Address)
	require.Equal(t, v.Ranges, res.Ranges)
	require.Equal(t, v.Salt, res.Salt)
	require.Equal(t, v.Type, res.Type)
}

func TestGetRangeHashResponse_Body(t *testing.T) {
	v := &object.GetRangeHashResponse_Body{
		Type:     1,
		HashList: [][]byte{[]byte("any_hash1"), []byte("any_hash2")},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res object.GetRangeHashResponse_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Type, res.Type)
	require.Equal(t, v.HashList, res.HashList)
}

func TestDeleteRequest_Body(t *testing.T) {
	v := &object.DeleteRequest_Body{
		Address: &refs.Address{
			ContainerId: &refs.ContainerID{Value: []byte("any_container")},
			ObjectId:    &refs.ObjectID{Value: []byte("any_object")},
		},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res object.DeleteRequest_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Address, res.Address)
}

func TestDeleteResponse_Body(t *testing.T) {
	v := &object.DeleteResponse_Body{
		Tombstone: &refs.Address{
			ContainerId: &refs.ContainerID{Value: []byte("any_container")},
			ObjectId:    &refs.ObjectID{Value: []byte("any_object")},
		},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res object.DeleteResponse_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Tombstone, res.Tombstone)
}

func TestPutRequest_Body(t *testing.T) {
	var v object.PutRequest_Body

	testWithPart := func(setPart func(body *object.PutRequest_Body)) {
		setPart(&v)
		sz := v.MarshaledSize()
		b := make([]byte, sz)
		v.MarshalStable(b)

		var res object.PutRequest_Body
		err := proto.Unmarshal(b, &res)
		require.NoError(t, err)
		require.Empty(t, res.ProtoReflect().GetUnknown())
		require.Equal(t, v.ObjectPart, res.ObjectPart)
	}

	testWithPart(func(body *object.PutRequest_Body) { body.ObjectPart = nil })

	testWithPart(func(body *object.PutRequest_Body) {
		body.ObjectPart = &object.PutRequest_Body_Init_{
			Init: &object.PutRequest_Body_Init{
				ObjectId: &refs.ObjectID{Value: []byte("any_object_ID")},
				Signature: &refs.Signature{
					Key:    []byte("any_public_key"),
					Sign:   []byte("any_signature"),
					Scheme: 1,
				},
				Header: &object.Header{
					Version:         &refs.Version{Major: 2, Minor: 3},
					ContainerId:     &refs.ContainerID{Value: []byte("any_container")},
					OwnerId:         &refs.OwnerID{Value: []byte("any_owner")},
					CreationEpoch:   4,
					PayloadLength:   5,
					PayloadHash:     &refs.Checksum{Type: 6, Sum: []byte("any_checksum")},
					ObjectType:      7,
					HomomorphicHash: &refs.Checksum{Type: 8, Sum: []byte("any_homomorphic_checksum")},
					SessionToken: &session.SessionToken{
						Body: &session.SessionToken_Body{
							Id:         []byte("any_ID"),
							OwnerId:    &refs.OwnerID{Value: []byte("any_owner")},
							Lifetime:   &session.SessionToken_Body_TokenLifetime{Exp: 9, Nbf: 10, Iat: 11},
							SessionKey: []byte("any_key"),
						},
						Signature: &refs.Signature{Key: []byte("any_key"), Sign: []byte("any_signature"), Scheme: 12},
					},
					Attributes: []*object.Header_Attribute{
						{Key: "attr_key1", Value: "attr_val1"},
						{Key: "attr_key2", Value: "attr_val2"},
					},
					Split: &object.Header_Split{
						Parent:   &refs.ObjectID{Value: []byte("any_parent_ID")},
						Previous: &refs.ObjectID{Value: []byte("any_previous")},
						ParentSignature: &refs.Signature{
							Key:    []byte("any_parent_public_key"),
							Sign:   []byte("any_parent_signature"),
							Scheme: 13,
						},
						ParentHeader: &object.Header{
							Version:         &refs.Version{Major: 100, Minor: 101},
							ContainerId:     &refs.ContainerID{Value: []byte("any_parent_container")},
							OwnerId:         &refs.OwnerID{Value: []byte("any_parent_owner")},
							CreationEpoch:   102,
							PayloadLength:   103,
							PayloadHash:     &refs.Checksum{Type: 104, Sum: []byte("any_parent_checksum")},
							ObjectType:      105,
							HomomorphicHash: &refs.Checksum{Type: 106, Sum: []byte("any_parent_homomorphic_checksum")},
							Attributes: []*object.Header_Attribute{
								{Key: "parent_attr_key1", Value: "parent_attr_val2"},
								{Key: "parent_attr_key2", Value: "parent_attr_val2"},
							},
						},
						Children: []*refs.ObjectID{{Value: []byte("any_child1")}, {Value: []byte("any_child2")}},
						SplitId:  []byte("any_split_ID"),
						First:    &refs.ObjectID{Value: []byte("any_first_ID")},
					},
				},
				CopiesNumber: 1,
			},
		}
	})

	testWithPart(func(body *object.PutRequest_Body) {
		body.ObjectPart = &object.PutRequest_Body_Chunk{Chunk: []byte("any_chunk")}
	})
}

func TestPutResponse_Body(t *testing.T) {
	v := &object.PutResponse_Body{
		ObjectId: &refs.ObjectID{Value: []byte("any_object")},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res object.PutResponse_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.ObjectId, res.ObjectId)
}

func TestSearchRequest_Body(t *testing.T) {
	v := &object.SearchRequest_Body{
		ContainerId: &refs.ContainerID{Value: []byte("any_container")},
		Version:     123,
		Filters: []*object.SearchRequest_Body_Filter{
			{MatchType: 456, Key: "k1", Value: "v1"},
			{MatchType: 789, Key: "k2", Value: "v2"},
		},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res object.SearchRequest_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.ContainerId, res.ContainerId)
	require.Equal(t, v.Version, res.Version)
	require.Equal(t, v.Filters, res.Filters)
}

func TestSearchResponse_Body(t *testing.T) {
	v := &object.SearchResponse_Body{
		IdList: []*refs.ObjectID{
			{Value: []byte("any_object1")},
			{Value: []byte("any_object2")},
			{Value: []byte("any_object3")},
		},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res object.SearchResponse_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.IdList, res.IdList)
}
