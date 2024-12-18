package acl_test

import (
	"testing"

	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	"github.com/nspcc-dev/neofs-sdk-go/proto/acl"
	prototest "github.com/nspcc-dev/neofs-sdk-go/proto/internal/test"
	"github.com/stretchr/testify/require"
)

func TestEACLRecord_Filter_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*acl.EACLRecord_Filter{
		prototest.RandEACLFilter(),
	})
}

func TestEACLRecord_Target_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*acl.EACLRecord_Target{
		prototest.RandEACLTarget(),
	})
}

func TestEACLRecord_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*acl.EACLRecord{
		prototest.RandEACLRecord(),
	})
}

func TestEACLTable_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*acl.EACLTable{
		prototest.RandEACL(),
	})
}

func TestBearerToken_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*acl.BearerToken_Body{
		prototest.RandBearerTokenBody(),
	})
}

func TestBearerToken_MarshalStable(t *testing.T) {
	t.Run("nil in repeated messages", func(t *testing.T) {
		src := &acl.BearerToken{
			Body: &acl.BearerToken_Body{
				EaclTable: &acl.EACLTable{
					Records: []*acl.EACLRecord{
						nil,
						{},
						{
							Filters: []*acl.EACLRecord_Filter{nil, {}},
							Targets: []*acl.EACLRecord_Target{nil, {}},
						},
					},
				},
			},
		}

		var dst acl.BearerToken
		require.NoError(t, neofsproto.UnmarshalMessage(neofsproto.MarshalMessage(src), &dst))

		rs := dst.GetBody().GetEaclTable().GetRecords()
		require.Len(t, rs, 3)
		require.Equal(t, rs[0], new(acl.EACLRecord))
		require.Equal(t, rs[1], new(acl.EACLRecord))
		fs := rs[2].GetFilters()
		require.Len(t, fs, 2)
		require.Equal(t, fs[0], new(acl.EACLRecord_Filter))
		require.Equal(t, fs[1], new(acl.EACLRecord_Filter))
		ts := rs[2].GetTargets()
		require.Len(t, ts, 2)
		require.Equal(t, ts[0], new(acl.EACLRecord_Target))
		require.Equal(t, ts[1], new(acl.EACLRecord_Target))
	})

	prototest.TestMarshalStable(t, []*acl.BearerToken{
		prototest.RandBearerToken(),
	})
}
