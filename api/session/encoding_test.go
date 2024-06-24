package session_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/acl"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	"github.com/nspcc-dev/neofs-sdk-go/api/session"
	"github.com/nspcc-dev/neofs-sdk-go/api/status"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestSessionToken_Body(t *testing.T) {
	v := &session.SessionToken{
		Body: &session.SessionToken_Body{
			Id:         []byte("any_ID"),
			OwnerId:    &refs.OwnerID{Value: []byte("any_owner")},
			Lifetime:   &session.SessionToken_Body_TokenLifetime{Exp: 1, Nbf: 2, Iat: 3},
			SessionKey: []byte("any_key"),
		},
		Signature: &refs.Signature{Key: []byte("any_key"), Sign: []byte("any_signature"), Scheme: 4},
	}

	testWithContext := func(setCtx func(*session.SessionToken_Body)) {
		setCtx(v.Body)
		sz := v.MarshaledSize()
		b := make([]byte, sz)
		v.MarshalStable(b)

		var res session.SessionToken
		err := proto.Unmarshal(b, &res)
		require.NoError(t, err)
		require.Empty(t, res.ProtoReflect().GetUnknown())
		require.Equal(t, v.Body, res.Body)
		require.Equal(t, v.Signature, res.Signature)
	}

	testWithContext(func(body *session.SessionToken_Body) { body.Context = nil })
	testWithContext(func(body *session.SessionToken_Body) {
		body.Context = &session.SessionToken_Body_Container{
			Container: &session.ContainerSessionContext{
				Verb: 1, Wildcard: true,
				ContainerId: &refs.ContainerID{Value: []byte("any_container")},
			},
		}
	})
	testWithContext(func(body *session.SessionToken_Body) {
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

func TestCreateRequest_Body(t *testing.T) {
	v := &session.CreateRequest_Body{
		OwnerId:    &refs.OwnerID{Value: []byte("any_owner")},
		Expiration: 1,
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res session.CreateRequest_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.OwnerId, res.OwnerId)
	require.Equal(t, v.Expiration, res.Expiration)
}

func TestCreateResponse_Body(t *testing.T) {
	v := &session.CreateResponse_Body{
		Id:         []byte("any_ID"),
		SessionKey: []byte("any_public_key"),
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res session.CreateResponse_Body
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Id, res.Id)
	require.Equal(t, v.SessionKey, res.SessionKey)
}

func TestRequestMetaHeader(t *testing.T) {
	v := &session.RequestMetaHeader{
		Version: &refs.Version{Major: 1, Minor: 2},
		Epoch:   3,
		Ttl:     4,
		XHeaders: []*session.XHeader{
			{Key: "any_key1", Value: "any_val1"},
			{Key: "any_key2", Value: "any_val2"},
		},
		SessionToken: &session.SessionToken{
			Body: &session.SessionToken_Body{
				Id:         []byte("any_ID"),
				OwnerId:    &refs.OwnerID{Value: []byte("any_owner")},
				Lifetime:   &session.SessionToken_Body_TokenLifetime{Exp: 101, Nbf: 102, Iat: 103},
				SessionKey: []byte("any_key"),
			},
			Signature: &refs.Signature{Key: []byte("any_key"), Sign: []byte("any_signature"), Scheme: 104},
		},
		BearerToken: &acl.BearerToken{
			Body: &acl.BearerToken_Body{
				EaclTable: &acl.EACLTable{
					Version:     &refs.Version{Major: 200, Minor: 201},
					ContainerId: &refs.ContainerID{Value: []byte("any_container")},
					Records: []*acl.EACLRecord{
						{
							Operation: 203, Action: 204,
							Filters: []*acl.EACLRecord_Filter{
								{HeaderType: 205, MatchType: 206, Key: "key1", Value: "val1"},
								{HeaderType: 207, MatchType: 208, Key: "key2", Value: "val2"},
							},
							Targets: []*acl.EACLRecord_Target{
								{Role: 209, Keys: [][]byte{{0}, {1}}},
								{Role: 210, Keys: [][]byte{{2}, {3}}},
							},
						},
						{
							Operation: 211, Action: 212,
							Filters: []*acl.EACLRecord_Filter{
								{HeaderType: 213, MatchType: 12, Key: "key3", Value: "val3"},
								{HeaderType: 214, MatchType: 14, Key: "key4", Value: "val4"},
							},
							Targets: []*acl.EACLRecord_Target{
								{Role: 215, Keys: [][]byte{{4}, {5}}},
								{Role: 216, Keys: [][]byte{{6}, {7}}},
							},
						},
					},
				},
				OwnerId:  &refs.OwnerID{Value: []byte("any_owner")},
				Lifetime: &acl.BearerToken_Body_TokenLifetime{Exp: 217, Nbf: 218, Iat: 219},
				Issuer:   &refs.OwnerID{Value: []byte("any_issuer")},
			},
			Signature: &refs.Signature{
				Key:    []byte("any_public_key"),
				Sign:   []byte("any_signature"),
				Scheme: 1,
			},
		},
		Origin: &session.RequestMetaHeader{
			Version: &refs.Version{Major: 300, Minor: 301},
			Epoch:   302,
			Ttl:     303,
			XHeaders: []*session.XHeader{
				{Key: "any_key3", Value: "any_val3"},
				{Key: "any_key4", Value: "any_val4"},
			},
			MagicNumber: 304,
		},
		MagicNumber: 5,
	}

	testWithSessionContext := func(setCtx func(*session.SessionToken_Body)) {
		setCtx(v.SessionToken.Body)
		sz := v.MarshaledSize()
		b := make([]byte, sz)
		v.MarshalStable(b)

		var res session.RequestMetaHeader
		err := proto.Unmarshal(b, &res)
		require.NoError(t, err)
		require.Empty(t, res.ProtoReflect().GetUnknown())
		require.Equal(t, v.Version, res.Version)
		require.Equal(t, v.Epoch, res.Epoch)
		require.Equal(t, v.Ttl, res.Ttl)
		require.Equal(t, v.XHeaders, res.XHeaders)
		require.Equal(t, v.SessionToken, res.SessionToken)
		require.Equal(t, v.BearerToken, res.BearerToken)
		require.Equal(t, v.Origin, res.Origin)
		require.Equal(t, v.MagicNumber, res.MagicNumber)
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

func TestResponseMetaHeader(t *testing.T) {
	v := &session.ResponseMetaHeader{
		Version: &refs.Version{Major: 1, Minor: 2},
		Epoch:   3,
		Ttl:     4,
		XHeaders: []*session.XHeader{
			{Key: "any_key1", Value: "any_val1"},
			{Key: "any_key2", Value: "any_val2"},
		},
		Origin: &session.ResponseMetaHeader{
			Version: &refs.Version{Major: 100, Minor: 101},
			Epoch:   102,
			Ttl:     103,
			XHeaders: []*session.XHeader{
				{Key: "any_key3", Value: "any_val3"},
				{Key: "any_key4", Value: "any_val4"},
			},
			Status: &status.Status{
				Code:    104,
				Message: "any_message",
				Details: []*status.Status_Detail{
					{Id: 105, Value: []byte("any_detail100")},
					{Id: 106, Value: []byte("any_detail101")},
				},
			},
		},
		Status: &status.Status{
			Code:    5,
			Message: "any_message",
			Details: []*status.Status_Detail{
				{Id: 6, Value: []byte("any_detail1")},
				{Id: 7, Value: []byte("any_detail2")},
			},
		},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res session.ResponseMetaHeader
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Version, res.Version)
	require.Equal(t, v.Epoch, res.Epoch)
	require.Equal(t, v.Ttl, res.Ttl)
	require.Equal(t, v.XHeaders, res.XHeaders)
	require.Equal(t, v.Origin, res.Origin)
	require.Equal(t, v.Status, res.Status)
}

func TestRequestVerificationHeader(t *testing.T) {
	v := &session.RequestVerificationHeader{
		BodySignature:   &refs.Signature{Key: []byte("any_pubkey1"), Sign: []byte("any_signature1"), Scheme: 1},
		MetaSignature:   &refs.Signature{Key: []byte("any_pubkey2"), Sign: []byte("any_signature2"), Scheme: 2},
		OriginSignature: &refs.Signature{Key: []byte("any_pubkey3"), Sign: []byte("any_signature3"), Scheme: 3},
		Origin: &session.RequestVerificationHeader{
			BodySignature:   &refs.Signature{Key: []byte("any_pubkey100"), Sign: []byte("any_signature100"), Scheme: 100},
			MetaSignature:   &refs.Signature{Key: []byte("any_pubkey101"), Sign: []byte("any_signature101"), Scheme: 101},
			OriginSignature: &refs.Signature{Key: []byte("any_pubkey102"), Sign: []byte("any_signature102"), Scheme: 102},
		},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res session.RequestVerificationHeader
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.BodySignature, res.BodySignature)
	require.Equal(t, v.MetaSignature, res.MetaSignature)
	require.Equal(t, v.OriginSignature, res.OriginSignature)
	require.Equal(t, v.Origin, res.Origin)
}

func TestResponseVerificationHeader(t *testing.T) {
	v := &session.ResponseVerificationHeader{
		BodySignature:   &refs.Signature{Key: []byte("any_pubkey1"), Sign: []byte("any_signature1"), Scheme: 1},
		MetaSignature:   &refs.Signature{Key: []byte("any_pubkey2"), Sign: []byte("any_signature2"), Scheme: 2},
		OriginSignature: &refs.Signature{Key: []byte("any_pubkey3"), Sign: []byte("any_signature3"), Scheme: 3},
		Origin: &session.ResponseVerificationHeader{
			BodySignature:   &refs.Signature{Key: []byte("any_pubkey100"), Sign: []byte("any_signature100"), Scheme: 100},
			MetaSignature:   &refs.Signature{Key: []byte("any_pubkey101"), Sign: []byte("any_signature101"), Scheme: 101},
			OriginSignature: &refs.Signature{Key: []byte("any_pubkey102"), Sign: []byte("any_signature102"), Scheme: 102},
		},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res session.ResponseVerificationHeader
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.BodySignature, res.BodySignature)
	require.Equal(t, v.MetaSignature, res.MetaSignature)
	require.Equal(t, v.OriginSignature, res.OriginSignature)
	require.Equal(t, v.Origin, res.Origin)
}
