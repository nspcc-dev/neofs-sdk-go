package acl_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/acl"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestBearerToken(t *testing.T) {
	v := &acl.BearerToken{
		Body: &acl.BearerToken_Body{
			EaclTable: &acl.EACLTable{
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
			OwnerId:  &refs.OwnerID{Value: []byte("any_owner")},
			Lifetime: &acl.BearerToken_Body_TokenLifetime{Exp: 17, Nbf: 18, Iat: 19},
			Issuer:   &refs.OwnerID{Value: []byte("any_issuer")},
		},
		Signature: &refs.Signature{
			Key:    []byte("any_public_key"),
			Sign:   []byte("any_signature"),
			Scheme: 100,
		},
	}

	sz := v.MarshaledSize()
	b := make([]byte, sz)
	v.MarshalStable(b)

	var res acl.BearerToken
	err := proto.Unmarshal(b, &res)
	require.NoError(t, err)
	require.Empty(t, res.ProtoReflect().GetUnknown())
	require.Equal(t, v.Body, res.Body)
	require.Equal(t, v.Signature, res.Signature)
}
